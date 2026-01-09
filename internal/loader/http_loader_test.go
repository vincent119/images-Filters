package loader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPLoader_CanLoad(t *testing.T) {
	loader := NewHTTPLoader()
	assert.True(t, loader.CanLoad("http://example.com/image.jpg"))
	assert.True(t, loader.CanLoad("https://example.com/image.jpg"))
	assert.False(t, loader.CanLoad("/local/path/image.jpg"))
	assert.False(t, loader.CanLoad("ftp://example.com/image.jpg"))
}

func TestHTTPLoader_Load(t *testing.T) {
	// 建立 Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/image.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte("fake image data"))
		case "/large.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(make([]byte, 1024*1024*2)) // 2MB
		case "/invalid_type":
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("not an image"))
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
		case "/timeout":
			time.Sleep(100 * time.Millisecond)
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write([]byte("delayed image"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	loader := NewHTTPLoader(
		WithMaxSize(1024*1024), // 1MB Limit
		WithHTTPTimeout(500*time.Millisecond),
	)

	tests := []struct {
		name      string
		url       string
		expectErr bool
	}{
		{
			name:      "Success",
			url:       server.URL + "/image.jpg",
			expectErr: false,
		},
		{
			name:      "File Too Large",
			url:       server.URL + "/large.jpg",
			expectErr: true,
		},
		{
			name:      "Invalid Content Type",
			url:       server.URL + "/invalid_type",
			expectErr: true,
		},
		{
			name:      "HTTP Error",
			url:       server.URL + "/error",
			expectErr: true,
		},
		{
			name:      "Not Found",
			url:       server.URL + "/not_found",
			expectErr: true,
		},
		{
			// Timeout is harder to test reliably in unit tests without mocking Client explicitly,
			// but we can try with short timeout
			name:      "Timeout",
			url:       server.URL + "/timeout",
			expectErr: false, // 100ms < 500ms
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := loader.Load(context.Background(), tt.url)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, data)
				if len(data) > 0 {
					assert.Contains(t, string(data), "image")
				}
			}
		})
	}

	// Test Timeout specifically with a text context
	t.Run("Real Timeout", func(t *testing.T) {
		timeoutLoader := NewHTTPLoader(WithHTTPTimeout(10 * time.Millisecond))
		_, err := timeoutLoader.Load(context.Background(), server.URL+"/timeout")
		assert.Error(t, err)
	})
}
