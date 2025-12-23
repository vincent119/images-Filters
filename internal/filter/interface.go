// Package filter 提供圖片濾鏡處理功能
package filter

import "image"

// Filter 濾鏡介面
// 所有濾鏡都必須實作此介面
type Filter interface {
	// Name 返回濾鏡名稱（用於 URL 解析）
	Name() string

	// Apply 對圖片應用濾鏡
	// params 為濾鏡參數（從 URL 解析）
	Apply(img image.Image, params []string) (image.Image, error)
}

// FilterFunc 函數式濾鏡包裝器
// 方便快速建立簡單濾鏡
type FilterFunc struct {
	name      string
	applyFunc func(img image.Image, params []string) (image.Image, error)
}

// NewFilterFunc 建立函數式濾鏡
func NewFilterFunc(name string, fn func(img image.Image, params []string) (image.Image, error)) *FilterFunc {
	return &FilterFunc{
		name:      name,
		applyFunc: fn,
	}
}

// Name 返回濾鏡名稱
func (f *FilterFunc) Name() string {
	return f.name
}

// Apply 應用濾鏡
func (f *FilterFunc) Apply(img image.Image, params []string) (image.Image, error) {
	return f.applyFunc(img, params)
}
