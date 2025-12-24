package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage 本地檔案儲存
type LocalStorage struct {
	rootPath string
}

// NewLocalStorage 建立本地儲存
func NewLocalStorage(rootPath string) (*LocalStorage, error) {
	// 確保根目錄存在
	if err := os.MkdirAll(rootPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		rootPath: rootPath,
	}, nil
}

// Get 取得圖片資料
func (s *LocalStorage) Get(ctx context.Context, key string) ([]byte, error) {
	path := s.resolvePath(key)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// Put 儲存圖片資料
func (s *LocalStorage) Put(ctx context.Context, key string, data []byte) error {
	path := s.resolvePath(key)

	// 確保目錄存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 寫入檔案
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Exists 檢查圖片是否存在
func (s *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	path := s.resolvePath(key)

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file: %w", err)
	}

	return true, nil
}

// Delete 刪除圖片
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	path := s.resolvePath(key)

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil // 檔案不存在視為成功
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// resolvePath 解析完整路徑
func (s *LocalStorage) resolvePath(key string) string {
	// 清理路徑，防止路徑遍歷
	cleanKey := filepath.Clean(key)
	cleanKey = strings.TrimPrefix(cleanKey, "/")
	cleanKey = strings.TrimPrefix(cleanKey, "\\")

	return filepath.Join(s.rootPath, cleanKey)
}

// GetStream 取得圖片資料串流
func (s *LocalStorage) GetStream(ctx context.Context, key string) (io.ReadCloser, error) {
	path := s.resolvePath(key)

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return f, nil
}

// PutStream 儲存圖片資料串流
func (s *LocalStorage) PutStream(ctx context.Context, key string, r io.Reader) error {
	path := s.resolvePath(key)

	// 確保目錄存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 建立檔案
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	// 寫入資料
	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("failed to write stream to file: %w", err)
	}

	return nil
}
