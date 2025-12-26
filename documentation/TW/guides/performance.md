# Performance Guide

[English](../performance.md)

## 設計原則

Images Filters 專為高吞吐量與低延遲而建構。

1. **非同步處理**: HTTP 請求處理與繁重的圖片運算透過 Worker Pool 解耦。
2. **記憶體效率**: 盡可能使用串流處理；高效的影像緩衝區記憶體配置。
3. **快取機制**: 積極的快取策略（第一層記憶體、第二層 Redis）。

### 基準測試 (參考)

**硬體規格**: 4 vCPU, 8GB RAM, SSD

| 操作 | RPS | P99 延遲 |
| ----------- | --- | ------------- |
| 健康檢查 | 20k+ | < 1ms |
| 快取圖片 (Redis) | 5k+ | < 5ms |
| 縮放 (1000px -> 300px) | 200+ | ~150ms |
| 濾鏡 (模糊) | 100+ | ~300ms |

### 調優指南

#### 1. Worker Pool (`processing.workers`)

- 預設值: 4
- 建議: 設定為 CPU 核心數 `NumCPU()`。設定過高不會改善 CPU 密集型任務的效能，反而增加 Context Switching。

#### 2. 記憶體快取 (`cache.memory.max_size`)

- 預設值: 512MB
- 建議: 若未啟用 Redis，建議設定為容器可用記憶體的 50-70%。

#### 3. Redis 連線池

- 預設值: 10
- 建議: 若在高負載下觀察到 `GetConn` 延遲過高，請增加 `pool.size`。
