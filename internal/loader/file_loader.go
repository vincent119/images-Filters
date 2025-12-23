package loader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileLoader 本地檔案載入器
type FileLoader struct {
	rootPath string // 根目錄路徑
	maxSize  int64  // 最大檔案大小
}

// FileLoaderOption 檔案載入器選項
type FileLoaderOption func(*FileLoader)

// WithRootPath 設定根目錄路徑
func WithRootPath(rootPath string) FileLoaderOption {
	return func(l *FileLoader) {
		l.rootPath = rootPath
	}
}

// WithFileMaxSize 設定最大檔案大小
func WithFileMaxSize(maxSize int64) FileLoaderOption {
	return func(l *FileLoader) {
		l.maxSize = maxSize
	}
}

// NewFileLoader 建立檔案載入器
func NewFileLoader(opts ...FileLoaderOption) *FileLoader {
	loader := &FileLoader{
		rootPath: "./data/images",
		maxSize:  10 * 1024 * 1024, // 預設 10MB
	}

	for _, opt := range opts {
		opt(loader)
	}

	return loader
}

// CanLoad 檢查是否可以載入指定來源
// 檔案載入器處理不是 HTTP URL 的路徑
func (l *FileLoader) CanLoad(source string) bool {
	return !strings.HasPrefix(source, "http://") && !strings.HasPrefix(source, "https://")
}

// Load 從本地檔案系統載入圖片
func (l *FileLoader) Load(ctx context.Context, source string) ([]byte, error) {
	// 建立完整路徑
	fullPath := l.resolvePath(source)

	// 安全檢查：防止路徑遍歷攻擊
	if err := l.validatePath(fullPath); err != nil {
		return nil, err
	}

	// 檢查檔案是否存在
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", source)
		}
		return nil, fmt.Errorf("cannot access file: %w", err)
	}

	// 檢查是否為目錄
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", source)
	}

	// 檢查檔案大小
	if l.maxSize > 0 && info.Size() > l.maxSize {
		return nil, fmt.Errorf("file too large: %d bytes (limit: %d)", info.Size(), l.maxSize)
	}

	// 讀取檔案
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// resolvePath 解析完整路徑
func (l *FileLoader) resolvePath(source string) string {
	// 如果是絕對路徑，直接返回
	if filepath.IsAbs(source) {
		return source
	}

	// 否則組合根目錄
	return filepath.Join(l.rootPath, source)
}

// validatePath 驗證路徑安全性
func (l *FileLoader) validatePath(fullPath string) error {
	// 取得絕對路徑
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// 取得根目錄絕對路徑
	absRoot, err := filepath.Abs(l.rootPath)
	if err != nil {
		return fmt.Errorf("failed to resolve root path: %w", err)
	}

	// 檢查路徑是否在根目錄內（防止路徑遍歷）
	relPath, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return fmt.Errorf("invalid relative path: %w", err)
	}

	// 如果相對路徑以 ".." 開頭，表示嘗試遍歷到根目錄之外
	if strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("illegal path traversal: %s", fullPath)
	}

	return nil
}

// isValidImageExtension 檢查是否為有效的圖片副檔名
func isValidImageExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".avif": true,
		".jxl":  true,
		".heic": true,
		".heif": true,
		".svg":  true,
		".bmp":  true,
		".tiff": true,
		".tif":  true,
	}
	return validExts[ext]
}
