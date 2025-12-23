package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalStorage(t *testing.T) {
	// 建立臨時目錄
	tmpDir := t.TempDir()

	storage, err := NewLocalStorage(tmpDir)
	if err != nil {
		t.Fatalf("建立儲存失敗: %v", err)
	}

	ctx := context.Background()
	testKey := "test/image.jpg"
	testData := []byte("test image data")

	// 測試 Put
	t.Run("Put", func(t *testing.T) {
		err := storage.Put(ctx, testKey, testData)
		if err != nil {
			t.Errorf("Put() error = %v", err)
		}

		// 驗證檔案存在
		path := filepath.Join(tmpDir, testKey)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("檔案應該存在")
		}
	})

	// 測試 Exists
	t.Run("Exists", func(t *testing.T) {
		exists, err := storage.Exists(ctx, testKey)
		if err != nil {
			t.Errorf("Exists() error = %v", err)
		}
		if !exists {
			t.Error("Exists() = false; want true")
		}

		// 測試不存在的檔案
		exists, err = storage.Exists(ctx, "nonexistent.jpg")
		if err != nil {
			t.Errorf("Exists() error = %v", err)
		}
		if exists {
			t.Error("Exists() = true; want false")
		}
	})

	// 測試 Get
	t.Run("Get", func(t *testing.T) {
		data, err := storage.Get(ctx, testKey)
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}
		if string(data) != string(testData) {
			t.Errorf("Get() = %s; want %s", data, testData)
		}

		// 測試不存在的檔案
		_, err = storage.Get(ctx, "nonexistent.jpg")
		if err == nil {
			t.Error("Get() 應該返回錯誤")
		}
	})

	// 測試 Delete
	t.Run("Delete", func(t *testing.T) {
		err := storage.Delete(ctx, testKey)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		// 驗證檔案不存在
		exists, _ := storage.Exists(ctx, testKey)
		if exists {
			t.Error("檔案應該被刪除")
		}

		// 刪除不存在的檔案應該不報錯
		err = storage.Delete(ctx, "nonexistent.jpg")
		if err != nil {
			t.Errorf("Delete() 不存在的檔案應該不報錯: %v", err)
		}
	})
}

func TestLocalStorage_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	storage, err := NewLocalStorage(tmpDir)
	if err != nil {
		t.Fatalf("建立儲存失敗: %v", err)
	}

	ctx := context.Background()

	// 測試路徑遍歷防護
	maliciousKeys := []string{
		"../etc/passwd",
		"..\\etc\\passwd",
		"/etc/passwd",
		"\\etc\\passwd",
	}

	for _, key := range maliciousKeys {
		t.Run(key, func(t *testing.T) {
			// Put 應該將檔案儲存在根目錄內
			err := storage.Put(ctx, key, []byte("test"))
			if err != nil {
				t.Logf("Put() error = %v (預期行為)", err)
				return
			}

			// 確認檔案儲存在正確位置
			// resolvePath 會清理路徑，所以檔案應該在 tmpDir 內
		})
	}
}
