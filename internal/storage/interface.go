// Package storage 提供圖片儲存功能
// 支援本地儲存、S3 儲存、無儲存等多種模式
package storage

import (
	"context"
)

// Storage 儲存介面
type Storage interface {
	// Get 取得圖片資料
	Get(ctx context.Context, key string) ([]byte, error)

	// Put 儲存圖片資料
	Put(ctx context.Context, key string, data []byte) error

	// Exists 檢查圖片是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Delete 刪除圖片
	Delete(ctx context.Context, key string) error
}
