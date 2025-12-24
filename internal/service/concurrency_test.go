package service

import (
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/loader"
	"github.com/vincent119/images-filters/internal/parser"
	"github.com/vincent119/images-filters/internal/processor"
)

// DelayLoader 模擬延遲載入
type DelayLoader struct {
	delay time.Duration
}

func (l *DelayLoader) Load(ctx context.Context, source string) ([]byte, error) {
	time.Sleep(l.delay)
	return []byte("fake-image-data"), nil
}

func (l *DelayLoader) CanLoad(source string) bool {
	return true
}

func (l *DelayLoader) LoadStream(ctx context.Context, source string) (io.ReadCloser, error) {
	time.Sleep(l.delay)
	// Return a dummy reader
	// Need bytes package or strings
	return io.NopCloser(strings.NewReader("fake-image-data")), nil
}

func TestWorkerPool_ConcurrencyLimit(t *testing.T) {
	// 1. Setup: Workers = 1
	cfg := &config.Config{
		Processing: config.ProcessingConfig{
			Workers:        1,
			DefaultQuality: 80,
			MaxWidth:       1000,
			MaxHeight:      1000,
			DefaultFormat:  "jpeg",
		},
		Storage: config.StorageConfig{
			Local: config.LocalStorageConfig{RootPath: "/tmp"},
		},
	}

	// 2. Mock Components
	mockLoader := &DelayLoader{delay: 100 * time.Millisecond}
	lf := loader.NewLoaderFactory(mockLoader)

	proc := processor.NewProcessor(80, 1000, 1000)
	store := &MockStorage{data: make(map[string][]byte)} // Reusing MockStorage from image_service_test.go if in same package
	// Wait, MockStorage is in image_service_test.go. Tests in same package share vars if in same folder.
	// But image_service_test.go defines MockStorage. Correct.

	c := &MockCache{data: make(map[string][]byte)} // Storage/Cache hit should be avoided

	// 3. Construct Service Manually
	svc := &imageService{
		cfg:       cfg,
		loader:    lf,
		processor: proc,
		storage:   store,
		cache:     c,
		sem:       make(chan struct{}, 1), // Capacity 1
	}

	// 4. Execute Concurrent Requests
	// Cache/Storage Miss -> Trigger Load (Delay) -> Process (Fail)

	parsedURL := &parser.ParsedURL{
		ImagePath: "http://example.com/test.jpg",
		Width:     100,
		Height:    100,
	}

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _, _ = svc.ProcessImage(context.Background(), parsedURL)
	}()

	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond) // Ensure first one gets lock first
		_, _, _ = svc.ProcessImage(context.Background(), parsedURL)
	}()

	wg.Wait()
	elapsed := time.Since(start)

	// 5. Verify
	// If parallel: max(100, 100) = ~100ms
	// If serial: 100 + 100 = ~200ms
	// Allow some buffer, say 180ms
	if elapsed < 180*time.Millisecond {
		t.Errorf("Expected serial execution (> 180ms), but got %v. Semaphore might not be working.", elapsed)
	} else {
		t.Logf("Success: Execution took %v, confirming serial processing.", elapsed)
	}
}
