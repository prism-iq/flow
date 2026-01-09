#pragma once

#include <cstddef>
#include <cstdint>
#include <vector>
#include <mutex>
#include <unordered_map>

namespace flow {

class MemoryPool {
public:
    static MemoryPool& instance();

    void* allocate(size_t size, size_t alignment = 64);
    void deallocate(void* ptr);
    void release_all();

    size_t allocated_bytes() const { return allocated_bytes_; }
    size_t pool_size() const { return pool_size_; }
    size_t allocation_count() const { return allocation_count_; }

    void set_max_pool_size(size_t max_bytes);

private:
    MemoryPool();
    ~MemoryPool();
    MemoryPool(const MemoryPool&) = delete;
    MemoryPool& operator=(const MemoryPool&) = delete;

    struct Block {
        void* ptr;
        size_t size;
        bool in_use;
    };

    std::vector<Block> blocks_;
    std::unordered_map<void*, size_t> ptr_to_block_;
    mutable std::mutex mutex_;

    size_t allocated_bytes_ = 0;
    size_t pool_size_ = 0;
    size_t max_pool_size_ = 1024 * 1024 * 1024;  // 1GB default
    size_t allocation_count_ = 0;

    void* find_free_block(size_t size, size_t alignment);
    void* allocate_new_block(size_t size, size_t alignment);
};

template<typename T>
class PoolAllocator {
public:
    using value_type = T;

    PoolAllocator() noexcept = default;

    template<typename U>
    PoolAllocator(const PoolAllocator<U>&) noexcept {}

    T* allocate(size_t n) {
        return static_cast<T*>(MemoryPool::instance().allocate(n * sizeof(T), alignof(T)));
    }

    void deallocate(T* ptr, size_t) noexcept {
        MemoryPool::instance().deallocate(ptr);
    }
};

template<typename T, typename U>
bool operator==(const PoolAllocator<T>&, const PoolAllocator<U>&) noexcept {
    return true;
}

template<typename T, typename U>
bool operator!=(const PoolAllocator<T>&, const PoolAllocator<U>&) noexcept {
    return false;
}

}  // namespace flow
