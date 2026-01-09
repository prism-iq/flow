#include "synapse.hpp"
#include "memory_pool.hpp"
#include <random>
#include <cmath>
#include <algorithm>

namespace flow {

Synapse::Synapse(const SynapseConfig& config)
    : config_(config)
    , weights_(config.input_dim * config.output_dim)
    , bias_(config.use_bias ? config.output_dim : 0)
    , grad_weights_(config.input_dim * config.output_dim)
    , grad_bias_(config.use_bias ? config.output_dim : 0)
    , input_cache_(config.input_dim) {

    std::random_device rd;
    std::mt19937 gen(rd());
    float scale = std::sqrt(2.0f / static_cast<float>(config.input_dim));
    std::normal_distribution<float> dist(0.0f, scale);

    for (auto& w : weights_) {
        w = dist(gen);
    }
    std::fill(bias_.begin(), bias_.end(), 0.0f);
    std::fill(grad_weights_.begin(), grad_weights_.end(), 0.0f);
    std::fill(grad_bias_.begin(), grad_bias_.end(), 0.0f);
}

Synapse::~Synapse() = default;

Synapse::Synapse(Synapse&&) noexcept = default;
Synapse& Synapse::operator=(Synapse&&) noexcept = default;

void Synapse::forward(const float* input, float* output) const {
    std::copy(input, input + config_.input_dim,
              const_cast<std::vector<float>&>(input_cache_).begin());

    for (size_t o = 0; o < config_.output_dim; ++o) {
        float sum = config_.use_bias ? bias_[o] : 0.0f;
        for (size_t i = 0; i < config_.input_dim; ++i) {
            sum += input[i] * weights_[i * config_.output_dim + o];
        }
        output[o] = sum;
    }
}

void Synapse::backward(const float* grad_output, float* grad_input) {
    for (size_t i = 0; i < config_.input_dim; ++i) {
        float sum = 0.0f;
        for (size_t o = 0; o < config_.output_dim; ++o) {
            sum += grad_output[o] * weights_[i * config_.output_dim + o];
            grad_weights_[i * config_.output_dim + o] +=
                input_cache_[i] * grad_output[o];
        }
        if (grad_input) {
            grad_input[i] = sum;
        }
    }

    if (config_.use_bias) {
        for (size_t o = 0; o < config_.output_dim; ++o) {
            grad_bias_[o] += grad_output[o];
        }
    }
}

void Synapse::update_weights() {
    for (size_t i = 0; i < weights_.size(); ++i) {
        weights_[i] -= config_.learning_rate * grad_weights_[i];
        grad_weights_[i] = 0.0f;
    }

    if (config_.use_bias) {
        for (size_t i = 0; i < bias_.size(); ++i) {
            bias_[i] -= config_.learning_rate * grad_bias_[i];
            grad_bias_[i] = 0.0f;
        }
    }
}

size_t Synapse::weight_count() const {
    return weights_.size() + bias_.size();
}

// SynapseNetwork implementation

void SynapseNetwork::add_layer(const SynapseConfig& config) {
    layers_.push_back(std::make_unique<Synapse>(config));
    activations_.emplace_back();
}

void SynapseNetwork::forward(const float* input, float* output) const {
    if (layers_.empty()) return;

    activations_[0].resize(layers_[0]->output_dim());
    layers_[0]->forward(input, activations_[0].data());

    for (size_t i = 1; i < layers_.size(); ++i) {
        activations_[i].resize(layers_[i]->output_dim());
        layers_[i]->forward(activations_[i-1].data(), activations_[i].data());
    }

    std::copy(activations_.back().begin(), activations_.back().end(), output);
}

void SynapseNetwork::backward(const float* grad_output) {
    if (layers_.empty()) return;

    std::vector<float> grad(grad_output, grad_output + layers_.back()->output_dim());

    for (int i = static_cast<int>(layers_.size()) - 1; i >= 0; --i) {
        std::vector<float> grad_input(layers_[i]->input_dim());
        layers_[i]->backward(grad.data(), grad_input.data());
        grad = std::move(grad_input);
    }
}

void SynapseNetwork::update() {
    for (auto& layer : layers_) {
        layer->update_weights();
    }
}

}  // namespace flow
