// Package loader 提供圖片載入功能
// 支援 HTTP/HTTPS 遠端載入和本地檔案載入
package loader

import (
	"context"
	"fmt"
	"io"
)

// Loader 圖片載入器介面
type Loader interface {
	// Load 載入圖片資料
	// 參數 source 可以是 URL 或檔案路徑
	Load(ctx context.Context, source string) ([]byte, error)

	// CanLoad 檢查是否可以載入指定來源
	CanLoad(source string) bool
}

// LoaderFactory 載入器工廠
type LoaderFactory struct {
	loaders []Loader
}

// NewLoaderFactory 建立載入器工廠
func NewLoaderFactory(loaders ...Loader) *LoaderFactory {
	return &LoaderFactory{
		loaders: loaders,
	}
}

// GetLoader 根據來源取得適合的載入器
func (f *LoaderFactory) GetLoader(source string) (Loader, error) {
	for _, loader := range f.loaders {
		if loader.CanLoad(source) {
			return loader, nil
		}
	}
	return nil, fmt.Errorf("no suitable loader found for: %s", source)
}

// Load 載入圖片
func (f *LoaderFactory) Load(ctx context.Context, source string) ([]byte, error) {
	loader, err := f.GetLoader(source)
	if err != nil {
		return nil, err
	}
	return loader.Load(ctx, source)
}

// readAll 讀取所有資料（帶大小限制）
func readAll(r io.Reader, maxSize int64) ([]byte, error) {
	if maxSize > 0 {
		r = io.LimitReader(r, maxSize)
	}
	return io.ReadAll(r)
}
