// Package service 提供圖片處理業務邏輯層
package service

import (
	"context"
	"image"

	"github.com/vincent119/images-filters/internal/parser"
)

// ImageService 圖片處理服務介面
type ImageService interface {
	// ProcessImage 處理圖片
	// 根據解析後的 URL 參數處理圖片並返回處理結果
	ProcessImage(ctx context.Context, parsedURL *parser.ParsedURL) ([]byte, string, error)
}

// ImageResult 圖片處理結果
type ImageResult struct {
	// 處理後的圖片資料
	Data []byte
	// Content-Type
	ContentType string
	// 處理後的圖片（可選，用於進一步處理）
	Image image.Image
}
