package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/pkg/logger"
)

// MemoryCache 實作基於 Ristretto 的記憶體快取
type MemoryCache struct {
	cache      *ristretto.Cache
	defaultTTL time.Duration
}

// NewMemoryCache 建立新的記憶體快取實例
func NewMemoryCache(cfg config.MemoryCacheConfig) (*MemoryCache, error) {
	// Ristretto config
	// NumCounters 建議為預期 BufferItems 的 10 倍，提升命中率追蹤準確度
	// BufferItems 建議設為預期項目數，這裡我們沒有直接設定項目數，而是用 MaxSize (bytes)
	// 但 Ristretto 需要 BufferItems (for admission buffers)。
	// 假設每個項目平均 100KB (大圖)，則 512MB 可存約 5000 張。
	// 若 MaxSize 為 512MB (536870912 bytes)，且平均 100KB。
	// 我們可以概估一個合理的 BufferItems。
	// 簡單起見，我們將 MaxSize 作為 Cost，BufferItems 設為一個夠大的值 (例如 64) 用於並發寫入緩衝。
	// 修正：BufferItems 是 "Expected number of items in admission buffer".
	// 官方建議 NumCounters = 10 * MaxCost 這是針對 Item Count based cost。
	// 但我們是 Size based cost。
	// 讓我們根據 MaxSize (bytes) 來設定 MaxCost。
	// NumCounters: 1e7 (1000萬) is good for standard usage.

	// 優化配置：
	// MaxCost: cfg.MaxSize (以 byte 計算)
	// NumCounters: 1e7 (足以追蹤大量 keys)
	// BufferItems: 64 (預設值通常足夠)

	c, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     cfg.MaxSize, // maximum cost of cache (In bytes).
		BufferItems: 64,      // number of keys per Get buffer.
		Metrics:     false, // enable if we want metrics later
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
	}

	ttl := time.Duration(cfg.TTL) * time.Second
	if ttl == 0 {
		ttl = time.Hour // 預設 1 小時
	}

	logger.Info("memory cache initialized (ristretto)",
		logger.Int64("max_size_bytes", cfg.MaxSize),
		logger.Int("ttl_sec", cfg.TTL),
	)

	return &MemoryCache{
		cache:      c,
		defaultTTL: ttl,
	}, nil
}

// Get 取得快取值
func (m *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, found := m.cache.Get(key)
	if !found {
		// Ristretto Get returns false/nil if expired or not found
		return nil, ErrCacheMiss
	}

	// Type assertion
	data, ok := val.([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid cache value type")
	}

	return data, nil
}

// Set 設定快取值
func (m *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = m.defaultTTL
	}

	// Cost 計算：使用 byte length
	cost := int64(len(value))

	// 使用 SetWithTTL
	if added := m.cache.SetWithTTL(key, value, cost, ttl); !added {
		// Set returns true if the item was accepted or updated.
		// Returns false if the item was rejected (e.g. too large).
		// 對於 Ristretto 來說，rejected 不一定是錯誤 (可能只是 policy 決定不收)，但在我們的 case 可能希望知道。
		// 不過，cache 是 best-effort，rejected 我們通常只能 log warn 或 ignore。
		// 這裡我們暫時當作成功 (不回傳 error)，因為快取本來就不是保證寫入。
		// 但可以 log 一個 debug。
		// logger.Debug("cache item rejected or dropped", logger.String("key", key))
		return nil
	}

	// Wait for value to pass through buffers (Ristretto is eventually consistent)
	// But we don't need to wait in Set.

	return nil
}

// Delete 刪除快取值
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.cache.Del(key)
	return nil
}

// Exists 檢查 Key 是否存在
func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	_, found := m.cache.Get(key)
	return found, nil
}

// Close 關閉快取 (釋放資源)
func (m *MemoryCache) Close() {
	m.cache.Close()
}
