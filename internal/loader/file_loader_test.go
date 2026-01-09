package loader

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileLoader_CanLoad(t *testing.T) {
	loader := NewFileLoader()
	assert.True(t, loader.CanLoad("image.jpg"))
	assert.True(t, loader.CanLoad("path/to/image.png"))
	assert.True(t, loader.CanLoad("/abs/path/image.webp"))
	assert.False(t, loader.CanLoad("http://example.com/image.jpg"))
	assert.False(t, loader.CanLoad("https://example.com/image.jpg"))
}

func TestFileLoader_Load(t *testing.T) {
	// 建立臨時測試目錄
	tmpDir, err := ioutil.TempDir("", "loader_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 1. 建立正常圖片文件
	validPath := filepath.Join(tmpDir, "test.jpg")
	err = ioutil.WriteFile(validPath, []byte("fake image data"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// 2. 建立過大文件
	largePath := filepath.Join(tmpDir, "large.jpg")
	largeData := make([]byte, 1024*1024+1) // > 1MB
	ioutil.WriteFile(largePath, largeData, 0644)

	// 3. 建立一個子目錄
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)

	loader := NewFileLoader(
		WithRootPath(tmpDir),
		WithFileMaxSize(1024*1024), // 1MB Limit
	)

	tests := []struct {
		name      string
		path      string
		expectErr bool
	}{
		{
			name:      "Success (Relative)",
			path:      "test.jpg", // Relative to root
			expectErr: false,
		},
		{
			name:      "File Too Large",
			path:      "large.jpg",
			expectErr: true,
		},
		{
			name:      "Not Found",
			path:      "notfound.jpg",
			expectErr: true,
		},
		{
			name:      "Path Traversal",
			path:      "../test.jpg", // Trying to go up from root (though here it might be valid if tmpDir is deep, logic prevents it)
			expectErr: true,
		},
		{
			name:      "Is Directory",
			path:      "subdir",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := loader.Load(context.Background(), tt.path)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, data)
				assert.Equal(t, "fake image data", string(data))
			}
		})
	}
}
