#include "tensor.hpp"
#include <random>
#include <stdexcept>
#include <cstring>
#include <numeric>

#ifdef __AVX2__
#include <immintrin.h>
#endif

namespace flow {

Tensor::Tensor(std::vector<size_t> shape, DType dtype)
    : shape_(std::move(shape))
    , dtype_(dtype) {
    compute_strides();
    size_ = std::accumulate(shape_.begin(), shape_.end(),
                            size_t(1), std::multiplies<>());
    data_ = std::make_unique<float[]>(size_);
}

Tensor::~Tensor() = default;

Tensor::Tensor(const Tensor& other)
    : shape_(other.shape_)
    , strides_(other.strides_)
    , size_(other.size_)
    , dtype_(other.dtype_)
    , data_(std::make_unique<float[]>(size_)) {
    std::memcpy(data_.get(), other.data_.get(), size_ * sizeof(float));
}

Tensor& Tensor::operator=(const Tensor& other) {
    if (this != &other) {
        shape_ = other.shape_;
        strides_ = other.strides_;
        size_ = other.size_;
        dtype_ = other.dtype_;
        data_ = std::make_unique<float[]>(size_);
        std::memcpy(data_.get(), other.data_.get(), size_ * sizeof(float));
    }
    return *this;
}

Tensor::Tensor(Tensor&&) noexcept = default;
Tensor& Tensor::operator=(Tensor&&) noexcept = default;

void Tensor::compute_strides() {
    strides_.resize(shape_.size());
    if (shape_.empty()) return;

    strides_.back() = 1;
    for (int i = static_cast<int>(shape_.size()) - 2; i >= 0; --i) {
        strides_[i] = strides_[i + 1] * shape_[i + 1];
    }
}

size_t Tensor::flat_index(const std::vector<size_t>& indices) const {
    size_t idx = 0;
    for (size_t i = 0; i < indices.size(); ++i) {
        idx += indices[i] * strides_[i];
    }
    return idx;
}

size_t Tensor::bytes() const {
    size_t elem_size = 4;
    switch (dtype_) {
        case DType::Float16: elem_size = 2; break;
        case DType::Int8: elem_size = 1; break;
        default: break;
    }
    return size_ * elem_size;
}

float& Tensor::at(std::vector<size_t> indices) {
    return data_[flat_index(indices)];
}

const float& Tensor::at(std::vector<size_t> indices) const {
    return data_[flat_index(indices)];
}

Tensor Tensor::zeros(std::vector<size_t> shape, DType dtype) {
    Tensor t(std::move(shape), dtype);
    std::memset(t.data_.get(), 0, t.size_ * sizeof(float));
    return t;
}

Tensor Tensor::ones(std::vector<size_t> shape, DType dtype) {
    Tensor t(std::move(shape), dtype);
    std::fill(t.data_.get(), t.data_.get() + t.size_, 1.0f);
    return t;
}

Tensor Tensor::rand(std::vector<size_t> shape, DType dtype) {
    Tensor t(std::move(shape), dtype);
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_real_distribution<float> dist(0.0f, 1.0f);
    for (size_t i = 0; i < t.size_; ++i) {
        t.data_[i] = dist(gen);
    }
    return t;
}

Tensor Tensor::reshape(std::vector<size_t> new_shape) const {
    size_t new_size = std::accumulate(new_shape.begin(), new_shape.end(),
                                       size_t(1), std::multiplies<>());
    if (new_size != size_) {
        throw std::runtime_error("reshape: incompatible sizes");
    }
    Tensor result(new_shape, dtype_);
    std::memcpy(result.data_.get(), data_.get(), size_ * sizeof(float));
    return result;
}

Tensor Tensor::operator+(const Tensor& other) const {
    if (size_ != other.size_) {
        throw std::runtime_error("add: size mismatch");
    }
    Tensor result(shape_, dtype_);
    add_simd(data_.get(), other.data_.get(), result.data_.get(), size_);
    return result;
}

Tensor Tensor::operator*(const Tensor& other) const {
    if (size_ != other.size_) {
        throw std::runtime_error("mul: size mismatch");
    }
    Tensor result(shape_, dtype_);
    mul_simd(data_.get(), other.data_.get(), result.data_.get(), size_);
    return result;
}

Tensor Tensor::matmul(const Tensor& other) const {
    if (ndim() != 2 || other.ndim() != 2 || shape_[1] != other.shape_[0]) {
        throw std::runtime_error("matmul: dimension mismatch");
    }

    size_t m = shape_[0];
    size_t k = shape_[1];
    size_t n = other.shape_[1];

    Tensor result({m, n}, dtype_);
    matmul_simd(data_.get(), other.data_.get(), result.data_.get(), m, n, k);
    return result;
}

// SIMD implementations

void matmul_simd(const float* a, const float* b, float* c,
                 size_t m, size_t n, size_t k) {
    std::memset(c, 0, m * n * sizeof(float));

#ifdef __AVX2__
    for (size_t i = 0; i < m; ++i) {
        for (size_t p = 0; p < k; ++p) {
            __m256 a_val = _mm256_set1_ps(a[i * k + p]);
            size_t j = 0;
            for (; j + 8 <= n; j += 8) {
                __m256 b_val = _mm256_loadu_ps(&b[p * n + j]);
                __m256 c_val = _mm256_loadu_ps(&c[i * n + j]);
                c_val = _mm256_fmadd_ps(a_val, b_val, c_val);
                _mm256_storeu_ps(&c[i * n + j], c_val);
            }
            for (; j < n; ++j) {
                c[i * n + j] += a[i * k + p] * b[p * n + j];
            }
        }
    }
#else
    for (size_t i = 0; i < m; ++i) {
        for (size_t j = 0; j < n; ++j) {
            for (size_t p = 0; p < k; ++p) {
                c[i * n + j] += a[i * k + p] * b[p * n + j];
            }
        }
    }
#endif
}

void add_simd(const float* a, const float* b, float* c, size_t n) {
#ifdef __AVX2__
    size_t i = 0;
    for (; i + 8 <= n; i += 8) {
        __m256 va = _mm256_loadu_ps(&a[i]);
        __m256 vb = _mm256_loadu_ps(&b[i]);
        __m256 vc = _mm256_add_ps(va, vb);
        _mm256_storeu_ps(&c[i], vc);
    }
    for (; i < n; ++i) {
        c[i] = a[i] + b[i];
    }
#else
    for (size_t i = 0; i < n; ++i) {
        c[i] = a[i] + b[i];
    }
#endif
}

void mul_simd(const float* a, const float* b, float* c, size_t n) {
#ifdef __AVX2__
    size_t i = 0;
    for (; i + 8 <= n; i += 8) {
        __m256 va = _mm256_loadu_ps(&a[i]);
        __m256 vb = _mm256_loadu_ps(&b[i]);
        __m256 vc = _mm256_mul_ps(va, vb);
        _mm256_storeu_ps(&c[i], vc);
    }
    for (; i < n; ++i) {
        c[i] = a[i] * b[i];
    }
#else
    for (size_t i = 0; i < n; ++i) {
        c[i] = a[i] * b[i];
    }
#endif
}

}  // namespace flow
