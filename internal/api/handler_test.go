package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vincent119/images-filters/internal/parser"
	"github.com/vincent119/images-filters/internal/service"
)

// mockImageService is a mock implementation of service.ImageService
type mockImageService struct {
	processFunc func(ctx context.Context, parsedURL *parser.ParsedURL) ([]byte, string, error)
	uploadFunc  func(ctx context.Context, filename string, contentType string, reader io.Reader) (*service.UploadResult, error)
}

func (m *mockImageService) ProcessImage(ctx context.Context, parsedURL *parser.ParsedURL) ([]byte, string, error) {
	if m.processFunc != nil {
		return m.processFunc(ctx, parsedURL)
	}
	return nil, "", nil
}

func (m *mockImageService) UploadImage(ctx context.Context, filename string, contentType string, reader io.Reader) (*service.UploadResult, error) {
	if m.uploadFunc != nil {
		return m.uploadFunc(ctx, filename, contentType, reader)
	}
	return nil, nil
}

func setupTestRouter() (*gin.Engine, *Handler, *mockImageService) {
	gin.SetMode(gin.TestMode)
	mockService := &mockImageService{}
	handler := NewHandler(mockService)
	router := gin.New()
	return router, handler, mockService
}

func TestHandler_HealthCheck(t *testing.T) {
	router, handler, _ := setupTestRouter()
	router.GET("/healthz", handler.HealthCheck)

	req, _ := http.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestHandler_HandleImage(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		mockProcess    func(ctx context.Context, parsedURL *parser.ParsedURL) ([]byte, string, error)
		expectedStatus int
		expectedBody   string // for error cases or content type check
	}{
		{
			name: "Success (Unsafe)",
			path: "/unsafe/300x200/http://example.com/image.jpg",
			mockProcess: func(ctx context.Context, parsedURL *parser.ParsedURL) ([]byte, string, error) {
				return []byte("fake_image_data"), "image/jpeg", nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid URL",
			path:           "/invalid",
			mockProcess:    nil, // Should not be called
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Image Not Found",
			path: "/unsafe/http://example.com/notfound.jpg",
			mockProcess: func(ctx context.Context, parsedURL *parser.ParsedURL) ([]byte, string, error) {
				return nil, "", errors.New("image not found")
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Processing Error",
			path: "/unsafe/http://example.com/error.jpg",
			mockProcess: func(ctx context.Context, parsedURL *parser.ParsedURL) ([]byte, string, error) {
				return nil, "", errors.New("internal error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, handler, mockService := setupTestRouter()
			mockService.processFunc = tt.mockProcess

			// We need to register a route that matches the path structure or use NoRoute/MatchAll
			// Since HandleImage parses request.URL.Path, we can just glob match everything
			router.GET("/*path", handler.HandleImage)

			req, _ := http.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "fake_image_data", w.Body.String())
				assert.Equal(t, "image/jpeg", w.Header().Get("Content-Type"))
			}
		})
	}
}

func TestHandler_HandleUpload(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func() (*http.Request, error)
		mockUpload     func(ctx context.Context, filename string, contentType string, reader io.Reader) (*service.UploadResult, error)
		expectedStatus int
	}{
		{
			name: "Success",
			setupRequest: func() (*http.Request, error) {
				body := new(bytes.Buffer)
				writer := multipart.NewWriter(body)
				h := make(map[string][]string)
				h["Content-Disposition"] = []string{`form-data; name="file"; filename="test.jpg"`}
				h["Content-Type"] = []string{"image/jpeg"}
				part, err := writer.CreatePart(h)
				if err != nil {
					return nil, err
				}
				part.Write([]byte("fake image content"))
				writer.Close()
				req, _ := http.NewRequest("POST", "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			mockUpload: func(ctx context.Context, filename string, contentType string, reader io.Reader) (*service.UploadResult, error) {
				return &service.UploadResult{
					Path:      "uploads/test.jpg",
					SignedURL: "/signed/test.jpg",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "No File",
			setupRequest: func() (*http.Request, error) {
				req, _ := http.NewRequest("POST", "/upload", nil) // No body
				return req, nil
			},
			mockUpload:     nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid File Type",
			setupRequest: func() (*http.Request, error) {
				body := new(bytes.Buffer)
				writer := multipart.NewWriter(body)
				h := make(map[string][]string)
				h["Content-Disposition"] = []string{`form-data; name="file"; filename="test.txt"`}
				h["Content-Type"] = []string{"text/plain"}
				part, err := writer.CreatePart(h)
				if err != nil {
					return nil, err
				}
				part.Write([]byte("text content"))
				writer.Close()
				req, _ := http.NewRequest("POST", "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			mockUpload:     nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Upload Error",
			setupRequest: func() (*http.Request, error) {
				body := new(bytes.Buffer)
				writer := multipart.NewWriter(body)
				h := make(map[string][]string)
				h["Content-Disposition"] = []string{`form-data; name="file"; filename="test.jpg"`}
				h["Content-Type"] = []string{"image/jpeg"}
				part, err := writer.CreatePart(h)
				if err != nil {
					return nil, err
				}
				part.Write([]byte("fake image content"))
				writer.Close()
				req, _ := http.NewRequest("POST", "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, nil
			},
			mockUpload: func(ctx context.Context, filename string, contentType string, reader io.Reader) (*service.UploadResult, error) {
				return nil, errors.New("upload failed")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, handler, mockService := setupTestRouter()
			mockService.uploadFunc = tt.mockUpload
			router.POST("/upload", handler.HandleUpload)

			req, err := tt.setupRequest()
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Logf("Response Body: %s", w.Body.String())
			}
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
