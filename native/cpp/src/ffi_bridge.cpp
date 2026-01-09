#include "ffi_bridge.hpp"
#include "synapse.hpp"
#include "tensor.hpp"
#include "memory_pool.hpp"
#include <cstring>

#define FLOW_VERSION "1.0.0"

using namespace flow;

extern "C" {

// Synapse FFI

FlowError flow_synapse_create(FlowSynapseHandle* handle,
                               size_t input_dim,
                               size_t output_dim,
                               float learning_rate,
                               int use_bias) {
    if (!handle) return FLOW_ERROR_INVALID_PARAM;

    try {
        SynapseConfig config{
            input_dim,
            output_dim,
            learning_rate,
            use_bias != 0
        };
        *handle = new Synapse(config);
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_OUT_OF_MEMORY;
    }
}

FlowError flow_synapse_forward(FlowSynapseHandle handle,
                                const float* input,
                                float* output) {
    if (!handle || !input || !output) return FLOW_ERROR_INVALID_PARAM;

    try {
        static_cast<Synapse*>(handle)->forward(input, output);
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_UNKNOWN;
    }
}

FlowError flow_synapse_backward(FlowSynapseHandle handle,
                                 const float* grad_output,
                                 float* grad_input) {
    if (!handle || !grad_output) return FLOW_ERROR_INVALID_PARAM;

    try {
        static_cast<Synapse*>(handle)->backward(grad_output, grad_input);
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_UNKNOWN;
    }
}

FlowError flow_synapse_update(FlowSynapseHandle handle) {
    if (!handle) return FLOW_ERROR_INVALID_HANDLE;

    try {
        static_cast<Synapse*>(handle)->update_weights();
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_UNKNOWN;
    }
}

FlowError flow_synapse_destroy(FlowSynapseHandle handle) {
    if (!handle) return FLOW_ERROR_INVALID_HANDLE;
    delete static_cast<Synapse*>(handle);
    return FLOW_OK;
}

// Tensor FFI

FlowError flow_tensor_create(FlowTensorHandle* handle,
                              const size_t* shape,
                              size_t ndim,
                              int dtype) {
    if (!handle || !shape || ndim == 0) return FLOW_ERROR_INVALID_PARAM;

    try {
        std::vector<size_t> shape_vec(shape, shape + ndim);
        *handle = new Tensor(shape_vec, static_cast<DType>(dtype));
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_OUT_OF_MEMORY;
    }
}

FlowError flow_tensor_create_zeros(FlowTensorHandle* handle,
                                    const size_t* shape,
                                    size_t ndim) {
    if (!handle || !shape || ndim == 0) return FLOW_ERROR_INVALID_PARAM;

    try {
        std::vector<size_t> shape_vec(shape, shape + ndim);
        *handle = new Tensor(Tensor::zeros(shape_vec));
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_OUT_OF_MEMORY;
    }
}

FlowError flow_tensor_create_rand(FlowTensorHandle* handle,
                                   const size_t* shape,
                                   size_t ndim) {
    if (!handle || !shape || ndim == 0) return FLOW_ERROR_INVALID_PARAM;

    try {
        std::vector<size_t> shape_vec(shape, shape + ndim);
        *handle = new Tensor(Tensor::rand(shape_vec));
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_OUT_OF_MEMORY;
    }
}

FlowError flow_tensor_get_data(FlowTensorHandle handle,
                                float** data,
                                size_t* size) {
    if (!handle || !data || !size) return FLOW_ERROR_INVALID_PARAM;

    auto tensor = static_cast<Tensor*>(handle);
    *data = tensor->data();
    *size = tensor->size();
    return FLOW_OK;
}

FlowError flow_tensor_set_data(FlowTensorHandle handle,
                                const float* data,
                                size_t size) {
    if (!handle || !data) return FLOW_ERROR_INVALID_PARAM;

    auto tensor = static_cast<Tensor*>(handle);
    if (size != tensor->size()) return FLOW_ERROR_DIMENSION_MISMATCH;

    std::memcpy(tensor->data(), data, size * sizeof(float));
    return FLOW_OK;
}

FlowError flow_tensor_matmul(FlowTensorHandle a,
                              FlowTensorHandle b,
                              FlowTensorHandle* result) {
    if (!a || !b || !result) return FLOW_ERROR_INVALID_PARAM;

    try {
        auto ta = static_cast<Tensor*>(a);
        auto tb = static_cast<Tensor*>(b);
        *result = new Tensor(ta->matmul(*tb));
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_DIMENSION_MISMATCH;
    }
}

FlowError flow_tensor_add(FlowTensorHandle a,
                           FlowTensorHandle b,
                           FlowTensorHandle* result) {
    if (!a || !b || !result) return FLOW_ERROR_INVALID_PARAM;

    try {
        auto ta = static_cast<Tensor*>(a);
        auto tb = static_cast<Tensor*>(b);
        *result = new Tensor(*ta + *tb);
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_DIMENSION_MISMATCH;
    }
}

FlowError flow_tensor_destroy(FlowTensorHandle handle) {
    if (!handle) return FLOW_ERROR_INVALID_HANDLE;
    delete static_cast<Tensor*>(handle);
    return FLOW_OK;
}

// Network FFI

FlowError flow_network_create(FlowNetworkHandle* handle) {
    if (!handle) return FLOW_ERROR_INVALID_PARAM;

    try {
        *handle = new SynapseNetwork();
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_OUT_OF_MEMORY;
    }
}

FlowError flow_network_add_layer(FlowNetworkHandle handle,
                                  size_t input_dim,
                                  size_t output_dim,
                                  float learning_rate) {
    if (!handle) return FLOW_ERROR_INVALID_HANDLE;

    try {
        auto network = static_cast<SynapseNetwork*>(handle);
        SynapseConfig config{input_dim, output_dim, learning_rate, true};
        network->add_layer(config);
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_OUT_OF_MEMORY;
    }
}

FlowError flow_network_forward(FlowNetworkHandle handle,
                                const float* input,
                                float* output) {
    if (!handle || !input || !output) return FLOW_ERROR_INVALID_PARAM;

    try {
        static_cast<SynapseNetwork*>(handle)->forward(input, output);
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_UNKNOWN;
    }
}

FlowError flow_network_backward(FlowNetworkHandle handle,
                                 const float* grad_output) {
    if (!handle || !grad_output) return FLOW_ERROR_INVALID_PARAM;

    try {
        static_cast<SynapseNetwork*>(handle)->backward(grad_output);
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_UNKNOWN;
    }
}

FlowError flow_network_update(FlowNetworkHandle handle) {
    if (!handle) return FLOW_ERROR_INVALID_HANDLE;

    try {
        static_cast<SynapseNetwork*>(handle)->update();
        return FLOW_OK;
    } catch (...) {
        return FLOW_ERROR_UNKNOWN;
    }
}

FlowError flow_network_destroy(FlowNetworkHandle handle) {
    if (!handle) return FLOW_ERROR_INVALID_HANDLE;
    delete static_cast<SynapseNetwork*>(handle);
    return FLOW_OK;
}

// Memory management

size_t flow_memory_allocated() {
    return MemoryPool::instance().allocated_bytes();
}

void flow_memory_release_all() {
    MemoryPool::instance().release_all();
}

// Version info

const char* flow_version() {
    return FLOW_VERSION;
}

}  // extern "C"
