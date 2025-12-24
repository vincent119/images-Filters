package cache

import (
	"context"
	"time"
)

// NoOpCache 不執行任何操作的快取實作
// 用於快取未啟用時
type NoOpCache struct{}

// NewNoOpCache 建立新的 NoOp 快取實例
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

func (c *NoOpCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, ErrCacheMiss
}

func (c *NoOpCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

func (c *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *NoOpCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}
