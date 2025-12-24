# Performance Guide

[繁體中文](TW/performance.md)

## Design Principles

Images Filters is built for high throughput and low latency.

1. **Asynchronous Processing**: HTTP handling is decoupled from heavy image processing via a Worker Pool.
2. **Memory Efficiency**: Streaming processing where possible; efficient memory allocation for image buffers.
3. **Caching**: Aggressive caching (Layer 1 Memory, Layer 2 Redis).

### Benchmarks (Reference)

**Hardware**: 4 vCPU, 8GB RAM, SSD

| Operation | RPS | P99 Latency |
| --------- | --- | ----------- |
| Health Check | 20k+ | < 1ms |
| Cached Image (Redis) | 5k+ | < 5ms |
| Resize (1000px -> 300px) | 200+ | ~150ms |
| Filter (Blur) | 100+ | ~300ms |

### Tuning Guide

#### 1. Worker Pool (`processing.workers`)

- Default: 4
- Recommendation: Set to `NumCPU()`. Setting higher does not improve performance for CPU-bound tasks and increases context switching.

#### 2. Memory Cache (`cache.memory.max_size`)

- Default: 512MB
- Recommendation: Set to 50-70% of available container memory if running without Redis.

#### 3. Redis Connection Pool

- Default: 10
- Recommendation: Increase `pool.size` if you see high `GetConn` latency in metrics under load.
