cmake_minimum_required(VERSION 3.16)
project(flow_synapses VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_POSITION_INDEPENDENT_CODE ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

option(BUILD_SHARED_LIBS "Build shared libraries" ON)
option(ENABLE_SIMD "Enable SIMD optimizations" ON)

if(ENABLE_SIMD)
    if(CMAKE_CXX_COMPILER_ID MATCHES "GNU|Clang")
        add_compile_options(-mavx2 -mfma)
    endif()
endif()

add_library(flow_synapses
    src/synapse.cpp
    src/tensor.cpp
    src/memory_pool.cpp
    src/ffi_bridge.cpp
    src/pattern_matcher.cpp
    src/pattern_ffi.cpp
)

find_package(Threads REQUIRED)
target_link_libraries(flow_synapses PRIVATE Threads::Threads)

target_include_directories(flow_synapses PUBLIC
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include>
)

target_compile_options(flow_synapses PRIVATE
    -Wall -Wextra -Wpedantic -O3
)

install(TARGETS flow_synapses
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
)

install(DIRECTORY include/ DESTINATION include)
