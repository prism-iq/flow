#pragma once

#include <cstdint>
#include <cstddef>
#include <vector>
#include <memory>
#include <functional>

namespace flow {

struct SynapseConfig {
    size_t input_dim;
    size_t output_dim;
    float learning_rate;
    bool use_bias;
};

class Synapse {
public:
    explicit Synapse(const SynapseConfig& config);
    ~Synapse();

    Synapse(const Synapse&) = delete;
    Synapse& operator=(const Synapse&) = delete;
    Synapse(Synapse&&) noexcept;
    Synapse& operator=(Synapse&&) noexcept;

    void forward(const float* input, float* output) const;
    void backward(const float* grad_output, float* grad_input);
    void update_weights();

    size_t input_dim() const { return config_.input_dim; }
    size_t output_dim() const { return config_.output_dim; }
    size_t weight_count() const;

    const float* weights() const { return weights_.data(); }
    const float* bias() const { return bias_.data(); }

private:
    SynapseConfig config_;
    std::vector<float> weights_;
    std::vector<float> bias_;
    std::vector<float> grad_weights_;
    std::vector<float> grad_bias_;
    std::vector<float> input_cache_;
};

class SynapseNetwork {
public:
    SynapseNetwork() = default;

    void add_layer(const SynapseConfig& config);
    void forward(const float* input, float* output) const;
    void backward(const float* grad_output);
    void update();

    size_t layer_count() const { return layers_.size(); }

private:
    std::vector<std::unique_ptr<Synapse>> layers_;
    mutable std::vector<std::vector<float>> activations_;
};

}  // namespace flow
