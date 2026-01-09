package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusMetrics Prometheus 指標實作
type PrometheusMetrics struct {
	registry *prometheus.Registry

	// ======== HTTP 入口層指標 ========
	httpRequestsTotal   *prometheus.CounterVec   // 請求總數
	httpRequestDuration *prometheus.HistogramVec // 請求處理時間
	httpInflightReqs    prometheus.Gauge         // 進行中請求數
	httpRequestSize     *prometheus.HistogramVec // 請求大小
	httpResponseSize    *prometheus.HistogramVec // 回應大小

	// ======== 圖片處理核心指標 ========
	imageProcessedTotal   *prometheus.CounterVec   // 處理圖片總數（按類型）
	imageSizeBytes        *prometheus.HistogramVec // 圖片大小分佈
	processingDuration    *prometheus.HistogramVec // 處理階段耗時
	processingOperations  *prometheus.CounterVec   // 處理操作計數
	processingErrors      *prometheus.CounterVec   // 處理錯誤分類
	inputImageDimensions  *prometheus.HistogramVec // 輸入圖片尺寸
	outputImageDimensions *prometheus.HistogramVec // 輸出圖片尺寸

	// ======== 錯誤相關指標 ========
	errorsTotal *prometheus.CounterVec // 錯誤總數

	// ======== 快取（Cache）指標 ========
	cacheHitsTotal   *prometheus.CounterVec   // 快取命中
	cacheMissesTotal *prometheus.CounterVec   // 快取未命中
	cacheLatency     *prometheus.HistogramVec // 快取延遲
	cacheEvictions   *prometheus.CounterVec   // 快取淘汰

	// ======== 儲存後端指標 ========
	storageOperations *prometheus.CounterVec   // 儲存操作計數
	storageLatency    *prometheus.HistogramVec // 儲存延遲
	storageErrors     *prometheus.CounterVec   // 儲存錯誤
	storageRetries    *prometheus.CounterVec   // 儲存重試

	// ======== 安全與風控指標 ========
	signatureValidations *prometheus.CounterVec // 簽名驗證
	rejectedRequests     *prometheus.CounterVec // 被拒絕請求
	rateLimitHits        prometheus.Counter     // 流量限制觸發

	// ======== 系統與效能 ========
	uptimeSeconds prometheus.Gauge // 服務運行時間
}

// NewPrometheusMetrics 建立 Prometheus 指標收集器
func NewPrometheusMetrics(namespace string) *PrometheusMetrics {
	registry := prometheus.NewRegistry()

	// 註冊預設的 Go 收集器
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	m := &PrometheusMetrics{
		registry: registry,

		// ======== HTTP 入口層指標 ========
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "HTTP 請求總數",
			},
			[]string{"method", "path", "status_code"},
		),

		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP 請求處理時間（秒）",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"method", "path", "status_code"},
		),

		httpInflightReqs: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_inflight_requests",
				Help:      "進行中的 HTTP 請求數",
			},
		),

		httpRequestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_size_bytes",
				Help:      "HTTP 請求大小（位元組）",
				Buckets:   []float64{100, 1024, 10240, 102400, 1048576, 10485760}, // 100B ~ 10MB
			},
			[]string{"method", "path"},
		),

		httpResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP 回應大小（位元組）",
				Buckets:   []float64{1024, 10240, 102400, 512000, 1048576, 5242880, 10485760}, // 1KB ~ 10MB
			},
			[]string{"method", "path"},
		),

		// ======== 圖片處理核心指標 ========
		imageProcessedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "images_processed_total",
				Help:      "處理圖片總數（按類型）",
			},
			[]string{"image_type"},
		),

		imageSizeBytes: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "image_size_bytes",
				Help:      "圖片大小分佈（位元組）",
				Buckets:   []float64{1024, 10240, 102400, 512000, 1048576, 5242880, 10485760}, // 1KB ~ 10MB
			},
			[]string{"image_type"},
		),

		processingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "image_processing_duration_seconds",
				Help:      "圖片處理階段耗時（秒）",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
			},
			[]string{"phase"},
		),

		processingOperations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "image_processing_operations_total",
				Help:      "圖片處理操作計數",
			},
			[]string{"operation"},
		),

		processingErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "image_processing_errors_total",
				Help:      "圖片處理錯誤分類計數",
			},
			[]string{"error_type"},
		),

		inputImageDimensions: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "image_input_dimensions",
				Help:      "輸入圖片尺寸分佈",
				Buckets:   []float64{100, 500, 1000, 2000, 4000, 8000}, // pixels
			},
			[]string{"dimension"}, // width / height
		),

		outputImageDimensions: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "image_output_dimensions",
				Help:      "輸出圖片尺寸分佈",
				Buckets:   []float64{100, 500, 1000, 2000, 4000, 8000}, // pixels
			},
			[]string{"dimension"}, // width / height
		),

		// ======== 錯誤相關指標 ========
		errorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "errors_total",
				Help:      "錯誤總數",
			},
			[]string{"error_type"},
		),

		// ======== 快取指標 ========
		cacheHitsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_hits_total",
				Help:      "快取命中總數",
			},
			[]string{"cache_type"},
		),

		cacheMissesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_misses_total",
				Help:      "快取未命中總數",
			},
			[]string{"cache_type"},
		),

		cacheLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "cache_latency_seconds",
				Help:      "快取操作延遲（秒）",
				Buckets:   []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5},
			},
			[]string{"operation", "cache_type"},
		),

		cacheEvictions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_evictions_total",
				Help:      "快取淘汰總數",
			},
			[]string{"cache_type"},
		),

		// ======== 儲存後端指標 ========
		storageOperations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "storage_operations_total",
				Help:      "儲存操作總數",
			},
			[]string{"backend", "operation"},
		),

		storageLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "storage_latency_seconds",
				Help:      "儲存操作延遲（秒）",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
			},
			[]string{"backend", "operation"},
		),

		storageErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "storage_errors_total",
				Help:      "儲存錯誤總數",
			},
			[]string{"backend", "error_type"},
		),

		storageRetries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "storage_retries_total",
				Help:      "儲存重試總數",
			},
			[]string{"backend"},
		),

		// ======== 安全與風控指標 ========
		signatureValidations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "signature_validations_total",
				Help:      "簽名驗證總數",
			},
			[]string{"success"},
		),

		rejectedRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "rejected_requests_total",
				Help:      "被拒絕請求總數",
			},
			[]string{"reason"},
		),

		rateLimitHits: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "rate_limit_hits_total",
				Help:      "流量限制觸發總數",
			},
		),

		// ======== 系統與效能 ========
		uptimeSeconds: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "uptime_seconds",
				Help:      "服務運行時間（秒）",
			},
		),
	}

	// 註冊所有指標
	registry.MustRegister(
		// HTTP
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.httpInflightReqs,
		m.httpRequestSize,
		m.httpResponseSize,
		// 圖片處理
		m.imageProcessedTotal,
		m.imageSizeBytes,
		m.processingDuration,
		m.processingOperations,
		m.processingErrors,
		m.inputImageDimensions,
		m.outputImageDimensions,
		// 錯誤
		m.errorsTotal,
		// 快取
		m.cacheHitsTotal,
		m.cacheMissesTotal,
		m.cacheLatency,
		m.cacheEvictions,
		// 儲存
		m.storageOperations,
		m.storageLatency,
		m.storageErrors,
		m.storageRetries,
		// 安全
		m.signatureValidations,
		m.rejectedRequests,
		m.rateLimitHits,
		// 系統
		m.uptimeSeconds,
	)

	return m
}

// ======== HTTP 入口層指標方法 ========

// RecordRequest 記錄 HTTP 請求
func (m *PrometheusMetrics) RecordRequest(method, path string, statusCode int, duration float64) {
	statusCodeStr := statusCodeToString(statusCode)
	m.httpRequestsTotal.WithLabelValues(method, path, statusCodeStr).Inc()
	m.httpRequestDuration.WithLabelValues(method, path, statusCodeStr).Observe(duration)
}

// RecordInflightRequest 記錄進行中請求數變化
func (m *PrometheusMetrics) RecordInflightRequest(delta int) {
	m.httpInflightReqs.Add(float64(delta))
}

// RecordRequestSize 記錄請求大小
func (m *PrometheusMetrics) RecordRequestSize(method, path string, bytes int64) {
	m.httpRequestSize.WithLabelValues(method, path).Observe(float64(bytes))
}

// RecordResponseSize 記錄回應大小
func (m *PrometheusMetrics) RecordResponseSize(method, path string, bytes int64) {
	m.httpResponseSize.WithLabelValues(method, path).Observe(float64(bytes))
}

// ======== 圖片處理核心指標方法 ========

// RecordImageProcessed 記錄處理的圖片
func (m *PrometheusMetrics) RecordImageProcessed(imageType string, sizeBytes int64) {
	m.imageProcessedTotal.WithLabelValues(imageType).Inc()
	m.imageSizeBytes.WithLabelValues(imageType).Observe(float64(sizeBytes))
}

// RecordProcessingDuration 記錄處理階段耗時
func (m *PrometheusMetrics) RecordProcessingDuration(phase string, seconds float64) {
	m.processingDuration.WithLabelValues(phase).Observe(seconds)
}

// RecordProcessingOperation 記錄處理操作類型
func (m *PrometheusMetrics) RecordProcessingOperation(operation string) {
	m.processingOperations.WithLabelValues(operation).Inc()
}

// RecordProcessingError 記錄處理錯誤分類
func (m *PrometheusMetrics) RecordProcessingError(errorType string) {
	m.processingErrors.WithLabelValues(errorType).Inc()
}

// RecordInputImageSize 記錄輸入圖片尺寸
func (m *PrometheusMetrics) RecordInputImageSize(width, height int) {
	m.inputImageDimensions.WithLabelValues("width").Observe(float64(width))
	m.inputImageDimensions.WithLabelValues("height").Observe(float64(height))
}

// RecordOutputImageSize 記錄輸出圖片尺寸
func (m *PrometheusMetrics) RecordOutputImageSize(width, height int) {
	m.outputImageDimensions.WithLabelValues("width").Observe(float64(width))
	m.outputImageDimensions.WithLabelValues("height").Observe(float64(height))
}

// RecordError 記錄錯誤
func (m *PrometheusMetrics) RecordError(errorType string) {
	m.errorsTotal.WithLabelValues(errorType).Inc()
}

// ======== 快取指標方法 ========

// RecordCacheHit 記錄快取命中
func (m *PrometheusMetrics) RecordCacheHit(cacheType string) {
	m.cacheHitsTotal.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss 記錄快取未命中
func (m *PrometheusMetrics) RecordCacheMiss(cacheType string) {
	m.cacheMissesTotal.WithLabelValues(cacheType).Inc()
}

// RecordCacheLatency 記錄快取操作延遲
func (m *PrometheusMetrics) RecordCacheLatency(operation, cacheType string, seconds float64) {
	m.cacheLatency.WithLabelValues(operation, cacheType).Observe(seconds)
}

// RecordCacheEviction 記錄快取淘汰
func (m *PrometheusMetrics) RecordCacheEviction(cacheType string) {
	m.cacheEvictions.WithLabelValues(cacheType).Inc()
}

// ======== 儲存後端指標方法 ========

// RecordStorageOperation 記錄儲存操作
func (m *PrometheusMetrics) RecordStorageOperation(backend, operation string) {
	m.storageOperations.WithLabelValues(backend, operation).Inc()
}

// RecordStorageLatency 記錄儲存延遲
func (m *PrometheusMetrics) RecordStorageLatency(backend, operation string, seconds float64) {
	m.storageLatency.WithLabelValues(backend, operation).Observe(seconds)
}

// RecordStorageError 記錄儲存錯誤
func (m *PrometheusMetrics) RecordStorageError(backend, errorType string) {
	m.storageErrors.WithLabelValues(backend, errorType).Inc()
}

// RecordStorageRetry 記錄儲存重試
func (m *PrometheusMetrics) RecordStorageRetry(backend string) {
	m.storageRetries.WithLabelValues(backend).Inc()
}

// ======== 安全與風控指標方法 ========

// RecordSignatureValidation 記錄簽名驗證結果
func (m *PrometheusMetrics) RecordSignatureValidation(success bool) {
	m.signatureValidations.WithLabelValues(fmt.Sprintf("%t", success)).Inc()
}

// RecordRejectedRequest 記錄被拒絕請求
func (m *PrometheusMetrics) RecordRejectedRequest(reason string) {
	m.rejectedRequests.WithLabelValues(reason).Inc()
}

// RecordRateLimitHit 記錄流量限制觸發
func (m *PrometheusMetrics) RecordRateLimitHit() {
	m.rateLimitHits.Inc()
}

// ======== 系統與效能指標方法 ========

// RecordUptimeSeconds 記錄服務運行時間
func (m *PrometheusMetrics) RecordUptimeSeconds(seconds float64) {
	m.uptimeSeconds.Set(seconds)
}

// ======== Registry ========

// Registry 取得 Prometheus Registry
func (m *PrometheusMetrics) Registry() *prometheus.Registry {
	return m.registry
}

// ======== 輔助函式 ========

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
