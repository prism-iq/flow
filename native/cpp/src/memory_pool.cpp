#include "memory_pool.hpp"
#include <cstdlib>
#include <stdexcept>
#include <algorithm>

namespace flow {

MemoryPool& MemoryPool::instance() {
    static MemoryPool pool;
    return pool;
}

MemoryPool::MemoryPool() = default;

MemoryPool::~MemoryPool() {
    release_all();
}

void* MemoryPool::allocate(size_t size, size_t alignment) {
    std::lock_guard<std::mutex> lock(mutex_);

    if (void* ptr = find_free_block(size, alignment)) {
        return ptr;
    }

    return allocate_new_block(size, alignment);
}

void* MemoryPool::find_free_block(size_t size, size_t alignment) {
    for (auto& block : blocks_) {
        if (!block.in_use && block.size >= size) {
            uintptr_t addr = reinterpret_cast<uintptr_t>(block.ptr);
            uintptr_t aligned = (addr + alignment - 1) & ~(alignment - 1);

            if (aligned == addr && block.size >= size) {
                block.in_use = true;
                return block.ptr;
            }
        }
    }
    return nullptr;
}

void* MemoryPool::allocate_new_block(size_t size, size_t alignment) {
    if (pool_size_ + size > max_pool_size_) {
        throw std::bad_alloc();
    }

    void* ptr = nullptr;

#if defined(_WIN32)
    ptr = _aligned_malloc(size, alignment);
#else
    if (posix_memalign(&ptr, alignment, size) != 0) {
        ptr = nullptr;
    }
#endif

    if (!ptr) {
        throw std::bad_alloc();
    }

    blocks_.push_back({ptr, size, true});
    ptr_to_block_[ptr] = blocks_.size() - 1;
    pool_size_ += size;
    allocated_bytes_ += size;
    ++allocation_count_;

    return ptr;
}

void MemoryPool::deallocate(void* ptr) {
    if (!ptr) return;

    std::lock_guard<std::mutex> lock(mutex_);

    auto it = ptr_to_block_.find(ptr);
    if (it != ptr_to_block_.end()) {
        blocks_[it->second].in_use = false;
        allocated_bytes_ -= blocks_[it->second].size;
    }
}

void MemoryPool::release_all() {
    std::lock_guard<std::mutex> lock(mutex_);

    for (auto& block : blocks_) {
#if defined(_WIN32)
        _aligned_free(block.ptr);
#else
        free(block.ptr);
#endif
    }

    blocks_.clear();
    ptr_to_block_.clear();
    allocated_bytes_ = 0;
    pool_size_ = 0;
}

void MemoryPool::set_max_pool_size(size_t max_bytes) {
    std::lock_guard<std::mutex> lock(mutex_);
    max_pool_size_ = max_bytes;
}

}  // namespace flow
