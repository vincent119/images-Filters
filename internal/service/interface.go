// Package service 提供圖片處理業務邏輯層
package service

import (
	"context"
	"image"
	"io"

	"github.com/vincent119/images-filters/internal/metrics"
	"github.com/vincent119/images-filters/internal/parser"
)

// ImageService 圖片處理服務介面
type ImageService interface {
	// ProcessImage 處理圖片
	// 根據解析後的 URL 參數處理圖片並返回處理結果
	ProcessImage(ctx context.Context, parsedURL *parser.ParsedURL) ([]byte, string, error)

	// UploadImage 上傳圖片
	// 儲存圖片並回傳儲存路徑與簽名 URL
	UploadImage(ctx context.Context, filename string, contentType string, reader io.Reader) (*UploadResult, error)
}

// UploadResult 上傳結果
type UploadResult struct {
	// Path 儲存路徑
	Path string `json:"path"`
	// SignedURL 簽名後的存取 URL
	SignedURL string `json:"url"`
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

// ServiceOption 服務選項
type ServiceOption func(*serviceOptions)

type serviceOptions struct {
	metrics metrics.Metrics
}

// WithMetrics 設定 metrics 收集器
func WithMetrics(m metrics.Metrics) ServiceOption {
	return func(o *serviceOptions) {
		o.metrics = m
	}
}
