package loader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vincent119/images-filters/pkg/logger"
)

// HTTPLoader HTTP/HTTPS 圖片載入器
type HTTPLoader struct {
	client    *http.Client
	maxSize   int64 // 最大檔案大小（位元組）
	timeout   time.Duration
	userAgent string
}

// HTTPLoaderOption HTTP 載入器選項
type HTTPLoaderOption func(*HTTPLoader)

// WithHTTPTimeout 設定請求逾時
func WithHTTPTimeout(timeout time.Duration) HTTPLoaderOption {
	return func(l *HTTPLoader) {
		l.timeout = timeout
		l.client.Timeout = timeout
	}
}

// WithMaxSize 設定最大檔案大小
func WithMaxSize(maxSize int64) HTTPLoaderOption {
	return func(l *HTTPLoader) {
		l.maxSize = maxSize
	}
}

// WithUserAgent 設定 User-Agent
func WithUserAgent(userAgent string) HTTPLoaderOption {
	return func(l *HTTPLoader) {
		l.userAgent = userAgent
	}
}

// NewHTTPLoader 建立 HTTP 載入器
func NewHTTPLoader(opts ...HTTPLoaderOption) *HTTPLoader {
	loader := &HTTPLoader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxSize:   10 * 1024 * 1024, // 預設 10MB
		timeout:   30 * time.Second,
		userAgent: "ImageProcessor/1.0",
	}

	for _, opt := range opts {
		opt(loader)
	}

	return loader
}

// CanLoad 檢查是否可以載入指定來源
func (l *HTTPLoader) CanLoad(source string) bool {
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")
}

// Load 從 HTTP/HTTPS 載入圖片
func (l *HTTPLoader) Load(ctx context.Context, source string) ([]byte, error) {
	logger.Debug("HTTP loader starting",
		logger.String("url", source),
		logger.Int64("max_size", l.maxSize),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		logger.Debug("failed to create HTTP request", logger.Err(err))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", l.userAgent)

	resp, err := l.client.Do(req)
	if err != nil {
		logger.Debug("HTTP request failed",
			logger.String("url", source),
			logger.Err(err),
		)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Debug("HTTP response status error",
			logger.String("url", source),
			logger.Int("status", resp.StatusCode),
		)
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// 檢查 Content-Length
	if l.maxSize > 0 && resp.ContentLength > l.maxSize {
		logger.Debug("file too large",
			logger.Int64("content_length", resp.ContentLength),
			logger.Int64("max_size", l.maxSize),
		)
		return nil, fmt.Errorf("file too large: %d bytes (limit: %d)", resp.ContentLength, l.maxSize)
	}

	// 驗證 Content-Type
	contentType := resp.Header.Get("Content-Type")
	if !isValidImageContentType(contentType) {
		logger.Debug("invalid Content-Type",
			logger.String("content_type", contentType),
		)
		return nil, fmt.Errorf("invalid Content-Type: %s", contentType)
	}

	// 讀取回應
	var reader io.Reader = resp.Body
	if l.maxSize > 0 {
		reader = io.LimitReader(resp.Body, l.maxSize+1) // +1 用於偵測超過限制
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		logger.Debug("failed to read response", logger.Err(err))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if l.maxSize > 0 && int64(len(data)) > l.maxSize {
		return nil, fmt.Errorf("file too large: exceeds %d bytes limit", l.maxSize)
	}

	logger.Debug("HTTP load successful",
		logger.String("url", source),
		logger.Int("size", len(data)),
		logger.String("content_type", contentType),
	)

	return data, nil
}

// isValidImageContentType 檢查是否為有效的圖片 Content-Type
func isValidImageContentType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/avif",
		"image/jxl",
		"image/heic",
		"image/heif",
		"image/svg+xml",
		"image/bmp",
		"image/tiff",
	}

	// 取得主要的 Content-Type（忽略參數如 charset）
	parts := strings.Split(contentType, ";")
	mainType := strings.TrimSpace(parts[0])

	for _, valid := range validTypes {
		if mainType == valid {
			return true
		}
	}

	// 允許通用的 image/* 類型
	return strings.HasPrefix(mainType, "image/")
}
