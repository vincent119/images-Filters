package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vincent119/images-filters/internal/config"
)

func setupTestRouter() *gin.Engine {
	// Initialize logger for testing without output to keep clean
	cfg := &config.LoggingConfig{
		Level:  "info",
		Format: "console",
		Output: "console", // Ideally noop or buffer, but for now console
	}
	Init(cfg)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestGinMiddleware(t *testing.T) {
	r := setupTestRouter()
	r.Use(GinMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.GET("/error", func(c *gin.Context) {
		c.AbortWithError(http.StatusBadRequest, assert.AnError)
	})
	r.GET("/server_error", func(c *gin.Context) {
		c.Status(http.StatusInternalServerError)
	})
	r.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.GET("/swagger/doc.json", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test Normal Request
	req, _ := http.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test 4xx Error
	req, _ = http.NewRequest("GET", "/error", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test 5xx Error
	req, _ = http.NewRequest("GET", "/server_error", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Test Skipped Paths
	req, _ = http.NewRequest("GET", "/healthz", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	req, _ = http.NewRequest("GET", "/swagger/doc.json", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGinRecovery(t *testing.T) {
	r := setupTestRouter()
	r.Use(GinRecovery())

	r.GET("/panic", func(c *gin.Context) {
		panic("oops")
	})

	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// Should allow panic to be recovered and return 500
	assert.NotPanics(t, func() {
		r.ServeHTTP(w, req)
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
