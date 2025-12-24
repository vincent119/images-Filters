package cache

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss 當快取 Key 不存在時回傳此錯誤
var ErrCacheMiss = errors.New("cache: key not found")

// Cache 定義通用快取介面
// 支援不同的快取實作（如 Redis, Memory 等）
type Cache interface {
	// Get 取得快取值
	// 若 Key 不存在，應回傳錯誤
	Get(ctx context.Context, key string) ([]byte, error)

	// Set 設定快取值
	// ttl 為過期時間，若為 0 表示不過期（或使用預設值）
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete 刪除快取值
	Delete(ctx context.Context, key string) error

	// Exists 檢查 Key 是否存在
	Exists(ctx context.Context, key string) (bool, error)
}
