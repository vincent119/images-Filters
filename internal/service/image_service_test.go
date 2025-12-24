package service

import (
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
	return nil, errors.New("not implemented")
}

func (m *MockStorage) PutStream(ctx context.Context, key string, r io.Reader) error {
	return errors.New("not implemented")
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

	// Generate key manually (or expose generateKey for testing, but let's rely on behavior)
	// Actually we need to force a hit.
	// Since generateKey is private, we can't easily predict the key hash without running the private method.
	// However, we can use the private method if we are in the same package `service`.
	// Yes, we are in `package service`.

	impl := svc.(*imageService)
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

	impl := svc.(*imageService)
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
			expected: "png",
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
