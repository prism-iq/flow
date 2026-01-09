#pragma once

#include <cstddef>
#include <cstdint>
#include <vector>
#include <memory>
#include <span>

namespace flow {

enum class DType : uint8_t {
    Float32,
    Float16,
    Int32,
    Int8
};

class Tensor {
public:
    Tensor(std::vector<size_t> shape, DType dtype = DType::Float32);
    ~Tensor();

    Tensor(const Tensor&);
    Tensor& operator=(const Tensor&);
    Tensor(Tensor&&) noexcept;
    Tensor& operator=(Tensor&&) noexcept;

    static Tensor zeros(std::vector<size_t> shape, DType dtype = DType::Float32);
    static Tensor ones(std::vector<size_t> shape, DType dtype = DType::Float32);
    static Tensor rand(std::vector<size_t> shape, DType dtype = DType::Float32);

    float* data() { return data_.get(); }
    const float* data() const { return data_.get(); }

    const std::vector<size_t>& shape() const { return shape_; }
    size_t ndim() const { return shape_.size(); }
    size_t size() const { return size_; }
    size_t bytes() const;
    DType dtype() const { return dtype_; }

    float& at(std::vector<size_t> indices);
    const float& at(std::vector<size_t> indices) const;

    Tensor reshape(std::vector<size_t> new_shape) const;
    Tensor transpose() const;
    Tensor slice(size_t dim, size_t start, size_t end) const;

    Tensor operator+(const Tensor& other) const;
    Tensor operator-(const Tensor& other) const;
    Tensor operator*(const Tensor& other) const;
    Tensor operator/(const Tensor& other) const;

    Tensor matmul(const Tensor& other) const;
    Tensor sum(int axis = -1) const;
    Tensor mean(int axis = -1) const;

private:
    std::vector<size_t> shape_;
    std::vector<size_t> strides_;
    size_t size_;
    DType dtype_;
    std::unique_ptr<float[]> data_;

    void compute_strides();
    size_t flat_index(const std::vector<size_t>& indices) const;
};

void matmul_simd(const float* a, const float* b, float* c,
                 size_t m, size_t n, size_t k);

void add_simd(const float* a, const float* b, float* c, size_t n);
void mul_simd(const float* a, const float* b, float* c, size_t n);

}  // namespace flow
