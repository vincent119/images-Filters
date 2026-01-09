package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/vincent119/images-filters/internal/cache"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/parser"
)

// MockCache 模擬快取
type MockCache struct {
	data map[string][]byte
	GetFunc func(ctx context.Context, key string) ([]byte, error)
	SetFunc func(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string][]byte),
	}
}

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, cache.ErrCacheMiss
}

func (m *MockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value, ttl)
	}
	m.data[key] = value
	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

// MockStorage 模擬儲存
type MockStorage struct {
	data map[string][]byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string][]byte),
	}
}

func (m *MockStorage) Get(ctx context.Context, key string) ([]byte, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, errors.New("file not found")
}

func (m *MockStorage) Put(ctx context.Context, key string, data []byte) error {
	m.data[key] = data
	return nil
}

func (m *MockStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockStorage) GetStream(ctx context.Context, key string) (io.ReadCloser, error) {
	if val, ok := m.data[key]; ok {
		// Create a ReadCloser
		return io.NopCloser(bytes.NewReader(val)), nil
	}
	return nil, errors.New("file not found")
}

func (m *MockStorage) PutStream(ctx context.Context, key string, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	m.data[key] = data
	return nil
}

func TestProcessImage_CacheHit(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Processing: config.ProcessingConfig{
			DefaultQuality: 80,
			MaxWidth: 1000,
			MaxHeight: 1000,
			DefaultFormat: "jpeg",
		},
		Server: config.ServerConfig{MaxRequestSize: 1024*1024},
		Storage: config.StorageConfig{
			Type: "local",
			Local: config.LocalStorageConfig{RootPath: "/tmp"},
		},
	}

	mockStore := NewMockStorage()
	mockCache := NewMockCache()
	svc := NewImageService(cfg, mockStore, mockCache)

	// Pre-populate cache
	parsedURL := &parser.ParsedURL{
		ImagePath: "test/image.jpg",
		Width: 100,
		Height: 100,
	}

	// Access internal generateKey
	type keyGenerator interface {
		generateKey(p *parser.ParsedURL) string
	}

	// We need to use reflection or copy logic if we can't access it.
	// But tests in same package can access private methods.
	// svc is ImageService interface, we need to assert to *imageService,
	// but *imageService is private in service.go.
	// Wait, internal/service/image_service.go:28 type imageService struct (lowercase)
	// So we cannot cast it here in a separate package?
	// Oh, `package service` in `image_service_test.go` means we ARE in the same package.
	// But `NewImageService` returns interface.

	// Reflection or accessing private fields/methods only works if the test file is in `package service`.
	// Line 1 says `package service`. So we can access `imageService` struct.

	// However, `svc` is declared as `ImageService` interface. We need to Type Assert.
	// But `imageService` struct is unexported. Can we type assert to unexported struct?
	// Yes, within the same package.

	// EXCEPT: if the previous file content view showed `imageService` is unexported `type imageService struct`.
	// Yes it is.

	// Wait, I cannot type assert to unexported type if I am in `package service_test` (if it was separate).
	// But here it is `package service`. So it should be fine.

	// Let's try to find how to do it.
	// Actually, `NewImageService` returns `ImageService` interface.
	// In Go, you can type assert to a struct defined in the same package.

	// BUT, note that in `image_service_test.go` provided before:
	// `impl := svc.(*imageService)`
	// This works if `TestProcessImage_CacheHit` is in `package service`.

	// Let's check `image_service_test.go` first line.
	// `package service`

	impl, ok := svc.(*imageService)
	if !ok {
		t.Fatal("Failed to cast service to *imageService")
	}

	key := impl.generateKey(parsedURL)
	cachedData := []byte("cached-image-data")
	mockCache.data[key] = cachedData

	// Execute
	data, contentType, err := svc.ProcessImage(context.Background(), parsedURL)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if string(data) != string(cachedData) {
		t.Errorf("Expected cached data, got %s", string(data))
	}
	if contentType != "image/jpeg" { // Default/Inferred format
		t.Errorf("Expected image/jpeg, got %s", contentType)
	}
}

func TestProcessImage_StorageHit_CacheMiss(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Processing: config.ProcessingConfig{
			DefaultQuality: 80,
			MaxWidth: 1000,
			MaxHeight: 1000,
			DefaultFormat: "jpeg",
		},
		Server: config.ServerConfig{MaxRequestSize: 1024*1024},
		Storage: config.StorageConfig{
			Type: "local",
			Local: config.LocalStorageConfig{RootPath: "/tmp"},
		},
	}

	mockStore := NewMockStorage()
	mockCache := NewMockCache()
	svc := NewImageService(cfg, mockStore, mockCache)

	parsedURL := &parser.ParsedURL{
		ImagePath: "test/image2.jpg",
		Width: 200,
		Height: 200,
	}

	impl, _ := svc.(*imageService)
	key := impl.generateKey(parsedURL)
	storedData := []byte("stored-processed-data")

	// Populate Storage but not Cache
	mockStore.data[key] = storedData

	// Execute
	data, _, err := svc.ProcessImage(context.Background(), parsedURL)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if string(data) != string(storedData) {
		t.Errorf("Expected stored data, got %s", string(data))
	}

	// Check if Cache was populated
	if val, ok := mockCache.data[key]; !ok {
		t.Error("Expected cache to be populated after storage hit")
	} else if string(val) != string(storedData) {
		t.Errorf("Expected cache data match, got %s", string(val))
	}
}

func TestDetermineFormat(t *testing.T) {
	cfg := &config.Config{
		Processing: config.ProcessingConfig{
			DefaultFormat: "jpeg",
		},
	}
	// Note: negotiateFormat is a method of *imageService struct, not interface.
	// But determineFormat is also a method of *imageService.
	// We need to initialize imageService struct directly to test internal method if it was private,
	// but determineFormat is private. So we test it via the struct.
	svc := &imageService{cfg: cfg}

	tests := []struct {
		name         string
		imagePath    string
		acceptHeader string
		filters      []parser.Filter
		expected     string
	}{
		{
			name:         "Default (No Accept, No Filter)",
			imagePath:    "image.jpg",
			acceptHeader: "",
			expected:     "jpeg",
		},
		{
			name:         "Explicit Filter Override",
			imagePath:    "image.jpg",
			acceptHeader: "image/avif,image/webp",
			filters: []parser.Filter{
				{Name: "format", Params: []string{"png"}},
			},
			expected:     "png",
		},
		{
			name:         "Accept AVIF",
			imagePath:    "image.jpg",
			acceptHeader: "image/avif,image/webp",
			expected:     "avif",
		},
		{
			name:         "Accept WebP",
			imagePath:    "image.jpg",
			acceptHeader: "image/webp,image/jpeg",
			expected:     "webp",
		},
		{
			name:         "Accept JXL",
			imagePath:    "image.jpg",
			acceptHeader: "image/jxl",
			expected:     "jxl",
		},
		{
			name:         "Complex Accept String",
			imagePath:    "image.jpg",
			acceptHeader: "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			expected:     "avif",
		},
		{
			name:         "Original Format (Validation)",
			imagePath:    "image.png",
			acceptHeader: "",
			expected:     "png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL := &parser.ParsedURL{
				ImagePath:    tt.imagePath,
				AcceptHeader: tt.acceptHeader,
				Filters:      tt.filters,
			}
			got := svc.determineFormat(parsedURL)
			if got != tt.expected {
				t.Errorf("determineFormat() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestUploadImage(t *testing.T) {
	// Setup
	cfg := &config.Config{
		Processing: config.ProcessingConfig{
			DefaultQuality: 80,
		},
		Security: config.SecurityConfig{
			Enabled:        true,
			SecurityKey:    "test-secret",
		},
		BlindWatermark: config.BlindWatermarkConfig{
			Enabled: false,
		},
	}

	mockStore := NewMockStorage()
	mockCache := NewMockCache()
	svc := NewImageService(cfg, mockStore, mockCache)

	filename := "test.jpg"
	contentType := "image/jpeg"
	content := []byte("fake-image-content")
	reader := bytes.NewReader(content)

	// Execute
	result, err := svc.UploadImage(context.Background(), filename, contentType, reader)

	// Verify
	if err != nil {
		t.Fatalf("UploadImage failed: %v", err)
	}
	if result.Path == "" {
		t.Error("Expected valid path")
	}
	if result.SignedURL == "" {
		t.Error("Expected valid signed URL")
	}

	// Verify storage
	if val, ok := mockStore.data[result.Path]; !ok {
		t.Error("File not saved to storage")
	} else if string(val) != string(content) {
		t.Errorf("Content mismatch, got %s", string(val))
	}
}

func TestGenerateSignedURL(t *testing.T) {
	// 1. Unsafe Mode
	cfgUnsafe := &config.Config{
		Security: config.SecurityConfig{
			Enabled: false,
		},
	}
	svcUnsafe := &imageService{cfg: cfgUnsafe}
	urlUnsafe := svcUnsafe.generateSignedURL("path/to/image.jpg")
	if urlUnsafe != "/unsafe/path/to/image.jpg" {
		t.Errorf("Expected /unsafe/path/to/image.jpg, got %s", urlUnsafe)
	}

	// 2. Secure Mode
	key := "test-secret"
	cfgSecure := &config.Config{
		Security: config.SecurityConfig{
			Enabled:     true,
			SecurityKey: key,
		},
	}
	svcSecure := &imageService{cfg: cfgSecure}
	urlSecure := svcSecure.generateSignedURL("path/to/image.jpg")

	// We don't want to re-implement signing logic here to check exact hash,
	// but we can check if it follows structure /{sig}/{path}
	// or verifies with same key.
	if urlSecure == "" {
		t.Error("Expected generated URL, got empty")
	}
	// Basic format check
	// Should not contain "unsafe"
	if bytes.Contains([]byte(urlSecure), []byte("unsafe")) {
		t.Error("Secure URL should not contain unsafe")
	}
}
