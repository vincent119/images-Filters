package storage

import (
	"context"
	"fmt"
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
		return nil, fmt.Errorf("建立儲存目錄失敗: %w", err)
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
			return nil, fmt.Errorf("檔案不存在: %s", key)
		}
		return nil, fmt.Errorf("讀取檔案失敗: %w", err)
	}

	return data, nil
}

// Put 儲存圖片資料
func (s *LocalStorage) Put(ctx context.Context, key string, data []byte) error {
	path := s.resolvePath(key)

	// 確保目錄存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("建立目錄失敗: %w", err)
	}

	// 寫入檔案
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("寫入檔案失敗: %w", err)
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
		return false, fmt.Errorf("檢查檔案失敗: %w", err)
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
		return fmt.Errorf("刪除檔案失敗: %w", err)
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
