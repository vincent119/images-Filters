package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetrics Prometheus 指標實作
type PrometheusMetrics struct {
	registry *prometheus.Registry

	// HTTP 請求相關指標
	httpRequestsTotal   *prometheus.CounterVec   // 請求總數
	httpRequestDuration *prometheus.HistogramVec // 請求處理時間

	// 圖片處理相關指標
	imageProcessedTotal *prometheus.CounterVec   // 處理圖片總數（按類型）
	imageSizeBytes      *prometheus.HistogramVec // 圖片大小分佈

	// 錯誤相關指標
	errorsTotal *prometheus.CounterVec // 錯誤總數
}

// NewPrometheusMetrics 建立 Prometheus 指標收集器
func NewPrometheusMetrics(namespace string) *PrometheusMetrics {
	registry := prometheus.NewRegistry()

	// 註冊預設的 Go 收集器
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	m := &PrometheusMetrics{
		registry: registry,

		// HTTP 請求總數
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "HTTP 請求總數",
			},
			[]string{"method", "path", "status_code"},
		),

		// HTTP 請求處理時間
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP 請求處理時間（秒）",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path", "status_code"},
		),

		// 圖片處理總數
		imageProcessedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "images_processed_total",
				Help:      "處理圖片總數（按類型）",
			},
			[]string{"image_type"},
		),

		// 圖片大小分佈
		imageSizeBytes: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "image_size_bytes",
				Help:      "圖片大小分佈（位元組）",
				Buckets:   []float64{1024, 10240, 102400, 512000, 1048576, 5242880, 10485760}, // 1KB ~ 10MB
			},
			[]string{"image_type"},
		),

		// 錯誤總數
		errorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_total",
				Help:      "錯誤總數",
			},
			[]string{"error_type"},
		),
	}

	// 註冊所有指標
	registry.MustRegister(
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.imageProcessedTotal,
		m.imageSizeBytes,
		m.errorsTotal,
	)

	return m
}

// RecordRequest 記錄 HTTP 請求
func (m *PrometheusMetrics) RecordRequest(method, path string, statusCode int, duration float64) {
	statusCodeStr := statusCodeToString(statusCode)

	m.httpRequestsTotal.WithLabelValues(method, path, statusCodeStr).Inc()
	m.httpRequestDuration.WithLabelValues(method, path, statusCodeStr).Observe(duration)
}

// RecordImageProcessed 記錄處理的圖片
func (m *PrometheusMetrics) RecordImageProcessed(imageType string, sizeBytes int64) {
	m.imageProcessedTotal.WithLabelValues(imageType).Inc()
	m.imageSizeBytes.WithLabelValues(imageType).Observe(float64(sizeBytes))
}

// RecordError 記錄錯誤
func (m *PrometheusMetrics) RecordError(errorType string) {
	m.errorsTotal.WithLabelValues(errorType).Inc()
}

// Registry 取得 Prometheus Registry
func (m *PrometheusMetrics) Registry() *prometheus.Registry {
	return m.registry
}

// statusCodeToString 將狀態碼轉為字串
func statusCodeToString(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}
