package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/metrics"
)

type MockMetrics struct {
	metrics.Metrics // Embed interface if needed, or implement methods stub
	rejected map[string]int
	signatures map[bool]int
}

// We need to implement all methods of Metrics interface or sufficient ones.
// Assuming Metrics is an interface with many methods.
// Using a simpler approach: define method stubs we need.

func (m *MockMetrics) RecordRejectedRequest(reason string) {
	if m.rejected == nil {
		m.rejected = make(map[string]int)
	}
	m.rejected[reason]++
}

func (m *MockMetrics) RecordSignatureValidation(valid bool) {
	if m.signatures == nil {
		m.signatures = make(map[bool]int)
	}
	m.signatures[valid]++
}

// Implement other methods to satisfy interface if passed directly.
// But middleware takes metrics.Metrics interface.
// We need to see `internal/metrics/metrics.go` to know full interface if we want to mock it.
// Or we can pass nil to middleware if acceptable (logic checks `if m != nil`).
// Let's pass nil first for simplicity unless we specifically test metrics recording.
// Or we can do a partial mock if interface allows or use `mock` package.
// Let's rely on nil check logic in middleware first.

// Stub other methods to satisfy interface
func (m *MockMetrics) RecordRequest(method, endpoint, status string) {}
func (m *MockMetrics) RecordDuration(method, endpoint string, duration float64) {}
func (m *MockMetrics) RecordError(errorType string) {}
func (m *MockMetrics) RecordCacheHit(cacheType string) {}
func (m *MockMetrics) RecordCacheMiss(cacheType string) {}
func (m *MockMetrics) RecordCacheLatency(op, cacheType string, duration float64) {}
func (m *MockMetrics) RecordStorageOperation(storageType, operation string) {}
func (m *MockMetrics) RecordStorageLatency(storageType, operation string, duration float64) {}
func (m *MockMetrics) RecordProcessingDuration(stage string, duration float64) {}
func (m *MockMetrics) RecordProcessingError(errorType string) {}
func (m *MockMetrics) RecordImageProcessed(format string, size int64) {}
func (m *MockMetrics) RecordProcessingOperation(operation string) {}
func (m *MockMetrics) RecordOutputImageSize(width, height int) {}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/test-cors", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	t.Run("GET Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test-cors", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("OPTIONS Request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/test-cors", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Capture log output
	buf := new(bytes.Buffer)
	gin.DefaultWriter = buf
	defer func() { gin.DefaultWriter = nil }() // Restore? Actually gin.DefaultWriter is stdout usually.

	r := gin.New()
	r.Use(LoggingMiddleware())
	r.GET("/log-test", func(c *gin.Context) {
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/log-test", nil)
	r.ServeHTTP(w, req)

	assert.Contains(t, buf.String(), "/log-test")
	assert.Contains(t, buf.String(), "200")
}

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Suppress panic output
	gin.DefaultErrorWriter = io.Discard

	r := gin.New()
	r.Use(RecoveryMiddleware())
	r.GET("/panic", func(c *gin.Context) {
		panic("oops")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSecurityMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Disabled", func(t *testing.T) {
		cfg := &config.SecurityConfig{Enabled: false}
		r := gin.New()
		r.Use(SecurityMiddleware(cfg, nil)) // nil metrics
		r.GET("/unsafe/test.jpg", func(c *gin.Context) { c.Status(200) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/unsafe/test.jpg", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Enabled - Unsafe Allowed", func(t *testing.T) {
		cfg := &config.SecurityConfig{Enabled: true, SecurityKey: "secret", AllowUnsafe: true}
		r := gin.New()
		r.Use(SecurityMiddleware(cfg, nil))
		r.GET("/unsafe/test.jpg", func(c *gin.Context) { c.Status(200) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/unsafe/test.jpg", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Enabled - Unsafe Forbidden", func(t *testing.T) {
		cfg := &config.SecurityConfig{Enabled: true, SecurityKey: "secret", AllowUnsafe: false}
		r := gin.New()
		r.Use(SecurityMiddleware(cfg, nil))
		r.GET("/unsafe/test.jpg", func(c *gin.Context) { c.Status(200) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/unsafe/test.jpg", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Enabled - Invalid Signature", func(t *testing.T) {
		cfg := &config.SecurityConfig{Enabled: true, SecurityKey: "secret"}
		r := gin.New()
		r.Use(SecurityMiddleware(cfg, nil))
		r.GET("/*path", func(c *gin.Context) { c.Status(200) })

		// Bad Signature
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/bad_sig/test.jpg", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "INVALID_SIGNATURE")
	})

	t.Run("Enabled - Skipped Path", func(t *testing.T) {
		cfg := &config.SecurityConfig{Enabled: true, SecurityKey: "secret"}
		r := gin.New()
		r.Use(SecurityMiddleware(cfg, nil))
		r.GET("/healthz", func(c *gin.Context) { c.Status(200) })

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/healthz", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Valid signature test would require generating signature.
	// We can use internal/security if needed or rely on integration tests.
	// But let's skip for now or duplicate logic.
}

func TestUploadAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key := "mysecret"

	r := gin.New()
	r.Use(UploadAuthMiddleware(key, nil))
	r.POST("/upload", func(c *gin.Context) { c.Status(200) })

	t.Run("Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/upload", nil)
		req.Header.Set("Authorization", "Bearer mysecret")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Missing Header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/upload", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Format", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/upload", nil)
		req.Header.Set("Authorization", "mysecret")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/upload", nil)
		req.Header.Set("Authorization", "Bearer wrong")
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestSourceValidatorMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	allowed := []string{"example.com"}

	r := gin.New()
	r.Use(SourceValidatorMiddleware(allowed))
	r.GET("/*path", func(c *gin.Context) { c.Status(200) })

	t.Run("Allowed Source", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/unsafe/http://example.com/image.jpg", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Denied Source", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/unsafe/http://evil.com/image.jpg", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Local File (Ignored)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/local/image.jpg", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
