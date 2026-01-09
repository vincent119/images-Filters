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
	"github.com/stretchr/testify/mock"
	"github.com/vincent119/images-filters/internal/service"
)

// MockWatermarkService
type MockWatermarkService struct {
	mock.Mock
}

func (m *MockWatermarkService) DetectWatermark(ctx context.Context, file io.Reader) (*service.DetectionResult, error) {
	args := m.Called(ctx, file)
	res := args.Get(0)
	if res == nil {
		return nil, args.Error(1)
	}
	return res.(*service.DetectionResult), args.Error(1)
}

func (m *MockWatermarkService) DetectWatermarkFromPath(ctx context.Context, path string) (*service.DetectionResult, error) {
	args := m.Called(ctx, path)
	res := args.Get(0)
	if res == nil {
		return nil, args.Error(1)
	}
	return res.(*service.DetectionResult), args.Error(1)
}

func setupWatermarkRouter(svc service.WatermarkService) (*gin.Engine, *WatermarkHandler) {
	gin.SetMode(gin.TestMode)
	h := NewWatermarkHandler(svc)
	r := gin.New()
	r.POST("/detect", h.HandleDetect)
	return r, h
}

func TestWatermarkHandler_HandleDetect(t *testing.T) {
	mockSvc := new(MockWatermarkService)
	r, _ := setupWatermarkRouter(mockSvc)

	t.Run("File Upload Success", func(t *testing.T) {
		mockSvc.On("DetectWatermark", mock.Anything, mock.Anything).Return(&service.DetectionResult{
			Detected:   true,
			Text:       "test-watermark",
			Confidence: 0.95,
		}, nil).Once()

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.jpg")
		part.Write([]byte("image data"))
		writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/detect", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"detected":true, "text":"test-watermark", "confidence":0.95}`, w.Body.String())
	})

	t.Run("File Upload Error", func(t *testing.T) {
		mockSvc.On("DetectWatermark", mock.Anything, mock.Anything).Return(nil, errors.New("detection failed")).Once()

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.jpg")
		part.Write([]byte("image data"))
		writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/detect", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "DETECTION_ERROR")
	})

	t.Run("Path Param Success", func(t *testing.T) {
		mockSvc.On("DetectWatermarkFromPath", mock.Anything, "uploads/test.jpg").Return(&service.DetectionResult{
			Detected:   false,
			Text:       "",
			Confidence: 0,
		}, nil).Once()

		// Send POST with form data 'path'
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("path", "uploads/test.jpg")
		writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/detect", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"detected":false, "text":"", "confidence":0}`, w.Body.String())
	})

	t.Run("Path Param Error", func(t *testing.T) {
		mockSvc.On("DetectWatermarkFromPath", mock.Anything, "uploads/missing.jpg").Return(nil, errors.New("file not found")).Once()

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("path", "uploads/missing.jpg")
		writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/detect", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "DETECTION_ERROR")
	})

	t.Run("Missing File and Path", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/detect", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "INVALID_REQUEST")
	})
}
