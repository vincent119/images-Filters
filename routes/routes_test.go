package routes

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/parser"
	"github.com/vincent119/images-filters/internal/service"
)

// MockImageService to satisfy interface
type MockImageService struct {
	ProcessImageFunc func(ctx context.Context, parsedURL *parser.ParsedURL) ([]byte, string, error)
	UploadImageFunc  func(ctx context.Context, filename string, contentType string, reader io.Reader) (*service.UploadResult, error)
}

func (m *MockImageService) ProcessImage(ctx context.Context, parsedURL *parser.ParsedURL) ([]byte, string, error) {
	if m.ProcessImageFunc != nil {
		return m.ProcessImageFunc(ctx, parsedURL)
	}
	return nil, "", nil
}

func (m *MockImageService) UploadImage(ctx context.Context, filename string, contentType string, reader io.Reader) (*service.UploadResult, error) {
	if m.UploadImageFunc != nil {
		return m.UploadImageFunc(ctx, filename, contentType, reader)
	}
	return nil, nil
}

// MockWatermarkService
type MockWatermarkService struct {
	DetectWatermarkFunc         func(ctx context.Context, file io.Reader) (*service.DetectionResult, error)
	DetectWatermarkFromPathFunc func(ctx context.Context, path string) (*service.DetectionResult, error)
}

func (m *MockWatermarkService) DetectWatermark(ctx context.Context, r io.Reader) (*service.DetectionResult, error) {
	if m.DetectWatermarkFunc != nil {
		return m.DetectWatermarkFunc(ctx, r)
	}
	return nil, nil
}

func (m *MockWatermarkService) DetectWatermarkFromPath(ctx context.Context, path string) (*service.DetectionResult, error) {
	if m.DetectWatermarkFromPathFunc != nil {
		return m.DetectWatermarkFromPathFunc(ctx, path)
	}
	return nil, nil
}

func TestSetup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup Config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
		Security: config.SecurityConfig{
			Enabled:        true,
			SecurityKey:    "secret",
			AllowedSources: []string{"*"}, // Allow all for test
		},
		Metrics: config.MetricsConfig{
			Enabled: true,
		},
		Swagger: config.SwaggerConfig{
			Enabled:  true,
			Username: "admin",
			Password: "password",
		},
	}

	// Mock Services
	mockImageService := &MockImageService{}
	mockWatermarkService := &MockWatermarkService{}

	// Setup Router
	router := gin.New()
	Setup(router, mockImageService, mockWatermarkService, cfg, nil)

	// Test Healthz
	req, _ := http.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test Swagger Auth (Unauthorized)
	req, _ = http.NewRequest("GET", "/swagger/index.html", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Test Swagger Auth (Authorized)
	req, _ = http.NewRequest("GET", "/swagger/index.html", nil)
	req.SetBasicAuth("admin", "password")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// Expect 301/200/404 depending on how ginSwagger wraps it, but definitively NOT 401
	// Usually redirects to index.html if asking for /swagger/
	// If asking for specific file that doesn't exist in mock fs, might be 404
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)

	// Test Upload Route (Unauthorized)
	req, _ = http.NewRequest("POST", "/upload", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSetup_NoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Security: config.SecurityConfig{
			Enabled: false,
		},
		Swagger: config.SwaggerConfig{
			Enabled: true, // No auth needed
		},
	}

	mockImageService := &MockImageService{}
	mockWatermarkService := &MockWatermarkService{}

	router := gin.New()
	Setup(router, mockImageService, mockWatermarkService, cfg, nil)

	// Test Upload Route (No Auth)
	req, _ := http.NewRequest("POST", "/upload", nil)
	w := httptest.NewRecorder()
	// Handler might fail due to bad request/no file, but NOT 401
	router.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)

	// Test Swagger No Auth
	req, _ = http.NewRequest("GET", "/swagger/index.html", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}
