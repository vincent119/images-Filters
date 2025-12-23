package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewPrometheusMetrics(t *testing.T) {
	m := NewPrometheusMetrics("test")

	if m == nil {
		t.Fatal("NewPrometheusMetrics returned nil")
	}

	if m.Registry() == nil {
		t.Fatal("Registry returned nil")
	}
}

func TestRecordRequest(t *testing.T) {
	m := NewPrometheusMetrics("test")

	// 記錄請求
	m.RecordRequest("GET", "/test", 200, 0.5)
	m.RecordRequest("GET", "/test", 200, 0.3)
	m.RecordRequest("POST", "/api", 500, 1.0)

	// 驗證請求總數
	expected := `
		# HELP test_http_requests_total HTTP 請求總數
		# TYPE test_http_requests_total counter
		test_http_requests_total{method="GET",path="/test",status_code="2xx"} 2
		test_http_requests_total{method="POST",path="/api",status_code="5xx"} 1
	`

	if err := testutil.CollectAndCompare(m.httpRequestsTotal, strings.NewReader(expected)); err != nil {
		t.Errorf("Unexpected collecting result: %v", err)
	}
}

func TestRecordImageProcessed(t *testing.T) {
	m := NewPrometheusMetrics("test")

	// 記錄圖片處理
	m.RecordImageProcessed("jpeg", 1024)
	m.RecordImageProcessed("jpeg", 2048)
	m.RecordImageProcessed("png", 4096)

	// 驗證圖片處理計數
	expected := `
		# HELP test_images_processed_total 處理圖片總數（按類型）
		# TYPE test_images_processed_total counter
		test_images_processed_total{image_type="jpeg"} 2
		test_images_processed_total{image_type="png"} 1
	`

	if err := testutil.CollectAndCompare(m.imageProcessedTotal, strings.NewReader(expected)); err != nil {
		t.Errorf("Unexpected collecting result: %v", err)
	}
}

func TestRecordError(t *testing.T) {
	m := NewPrometheusMetrics("test")

	// 記錄錯誤
	m.RecordError("load_error")
	m.RecordError("load_error")
	m.RecordError("process_error")

	// 驗證錯誤計數
	expected := `
		# HELP test_errors_total 錯誤總數
		# TYPE test_errors_total counter
		test_errors_total{error_type="load_error"} 2
		test_errors_total{error_type="process_error"} 1
	`

	if err := testutil.CollectAndCompare(m.errorsTotal, strings.NewReader(expected)); err != nil {
		t.Errorf("Unexpected collecting result: %v", err)
	}
}

func TestStatusCodeToString(t *testing.T) {
	tests := []struct {
		code     int
		expected string
	}{
		{200, "2xx"},
		{201, "2xx"},
		{301, "3xx"},
		{400, "4xx"},
		{404, "4xx"},
		{500, "5xx"},
		{503, "5xx"},
		{100, "unknown"},
	}

	for _, tt := range tests {
		result := statusCodeToString(tt.code)
		if result != tt.expected {
			t.Errorf("statusCodeToString(%d) = %s; want %s", tt.code, result, tt.expected)
		}
	}
}

func TestGinMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := NewPrometheusMetrics("test")

	// 建立測試路由
	router := gin.New()
	router.Use(GinMiddleware(m))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "internal error"})
	})

	// 測試正常請求
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 測試錯誤請求
	req = httptest.NewRequest(http.MethodGet, "/error", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestGinMiddlewareSkipPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := NewPrometheusMetrics("test")

	// 建立測試路由
	router := gin.New()
	router.Use(GinMiddleware(m))
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/metrics", func(c *gin.Context) {
		c.String(200, "metrics")
	})

	// /healthz 應該被跳過（不記錄指標）
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// /metrics 應該被跳過
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMetricsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := NewPrometheusMetrics("test")

	// 建立測試路由
	router := gin.New()
	router.GET("/metrics", MetricsHandler(m, "", ""))

	// 測試無認證
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// 驗證回應包含 Prometheus 格式
	body := w.Body.String()
	if !strings.Contains(body, "go_goroutines") {
		t.Error("Expected metrics output to contain go_goroutines")
	}
}

func TestMetricsHandlerWithAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := NewPrometheusMetrics("test")

	// 建立測試路由（需要認證）
	router := gin.New()
	router.GET("/metrics", MetricsHandler(m, "admin", "password"))

	// 測試無認證（應該失敗）
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("Expected status 401 without auth, got %d", w.Code)
	}

	// 測試錯誤認證（應該失敗）
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.SetBasicAuth("wrong", "wrong")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("Expected status 401 with wrong auth, got %d", w.Code)
	}

	// 測試正確認證
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.SetBasicAuth("admin", "password")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200 with correct auth, got %d", w.Code)
	}
}
