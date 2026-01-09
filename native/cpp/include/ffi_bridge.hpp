#pragma once

#include <cstdint>
#include <cstddef>

#ifdef __cplusplus
extern "C" {
#endif

// Opaque handles for FFI
typedef void* FlowSynapseHandle;
typedef void* FlowTensorHandle;
typedef void* FlowNetworkHandle;

// Error codes
typedef enum {
    FLOW_OK = 0,
    FLOW_ERROR_INVALID_HANDLE = -1,
    FLOW_ERROR_INVALID_PARAM = -2,
    FLOW_ERROR_OUT_OF_MEMORY = -3,
    FLOW_ERROR_DIMENSION_MISMATCH = -4,
    FLOW_ERROR_UNKNOWN = -99
} FlowError;

// Synapse FFI
FlowError flow_synapse_create(FlowSynapseHandle* handle,
                               size_t input_dim,
                               size_t output_dim,
                               float learning_rate,
                               int use_bias);

FlowError flow_synapse_forward(FlowSynapseHandle handle,
                                const float* input,
                                float* output);

FlowError flow_synapse_backward(FlowSynapseHandle handle,
                                 const float* grad_output,
                                 float* grad_input);

FlowError flow_synapse_update(FlowSynapseHandle handle);

FlowError flow_synapse_destroy(FlowSynapseHandle handle);

// Tensor FFI
FlowError flow_tensor_create(FlowTensorHandle* handle,
                              const size_t* shape,
                              size_t ndim,
                              int dtype);

FlowError flow_tensor_create_zeros(FlowTensorHandle* handle,
                                    const size_t* shape,
                                    size_t ndim);

FlowError flow_tensor_create_rand(FlowTensorHandle* handle,
                                   const size_t* shape,
                                   size_t ndim);

FlowError flow_tensor_get_data(FlowTensorHandle handle,
                                float** data,
                                size_t* size);

FlowError flow_tensor_set_data(FlowTensorHandle handle,
                                const float* data,
                                size_t size);

FlowError flow_tensor_matmul(FlowTensorHandle a,
                              FlowTensorHandle b,
                              FlowTensorHandle* result);

FlowError flow_tensor_add(FlowTensorHandle a,
                           FlowTensorHandle b,
                           FlowTensorHandle* result);

FlowError flow_tensor_destroy(FlowTensorHandle handle);

// Network FFI
FlowError flow_network_create(FlowNetworkHandle* handle);

FlowError flow_network_add_layer(FlowNetworkHandle handle,
                                  size_t input_dim,
                                  size_t output_dim,
                                  float learning_rate);

FlowError flow_network_forward(FlowNetworkHandle handle,
                                const float* input,
                                float* output);

FlowError flow_network_backward(FlowNetworkHandle handle,
                                 const float* grad_output);

FlowError flow_network_update(FlowNetworkHandle handle);

FlowError flow_network_destroy(FlowNetworkHandle handle);

// Memory management
size_t flow_memory_allocated();
void flow_memory_release_all();

// Version info
const char* flow_version();

#ifdef __cplusplus
}
#endif
