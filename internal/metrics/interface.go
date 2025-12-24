// Package metrics 提供 Prometheus 指標收集功能
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics 指標介面
type Metrics interface {
	// ======== HTTP 入口層指標 ========
	// RecordRequest 記錄 HTTP 請求（總數與延遲）
	RecordRequest(method, path string, statusCode int, duration float64)
	// RecordInflightRequest 記錄進行中請求數變化（delta: +1 開始, -1 結束）
	RecordInflightRequest(delta int)
	// RecordRequestSize 記錄請求大小（bytes）
	RecordRequestSize(method, path string, bytes int64)
	// RecordResponseSize 記錄回應大小（bytes）
	RecordResponseSize(method, path string, bytes int64)

	// ======== 圖片處理核心指標 ========
	// RecordImageProcessed 記錄處理完成的圖片
	RecordImageProcessed(imageType string, sizeBytes int64)
	// RecordProcessingDuration 記錄處理階段耗時（phase: decode/transform/encode）
	RecordProcessingDuration(phase string, seconds float64)
	// RecordProcessingOperation 記錄處理操作類型計數（resize/crop/flip/watermark/filter）
	RecordProcessingOperation(operation string)
	// RecordProcessingError 記錄處理錯誤分類（decode_failed/unsupported/timeout/oom）
	RecordProcessingError(errorType string)
	// RecordInputImageSize 記錄輸入圖片尺寸
	RecordInputImageSize(width, height int)
	// RecordOutputImageSize 記錄輸出圖片尺寸
	RecordOutputImageSize(width, height int)
	// RecordError 記錄一般錯誤
	RecordError(errorType string)

	// ======== 快取（Cache）指標 ========
	// RecordCacheHit 記錄快取命中（cacheType: redis/memory）
	RecordCacheHit(cacheType string)
	// RecordCacheMiss 記錄快取未命中
	RecordCacheMiss(cacheType string)
	// RecordCacheLatency 記錄快取操作延遲（operation: get/set）
	RecordCacheLatency(operation, cacheType string, seconds float64)
	// RecordCacheEviction 記錄快取淘汰
	RecordCacheEviction(cacheType string)

	// ======== 儲存後端指標 ========
	// RecordStorageOperation 記錄儲存操作計數（backend: s3/local, operation: get/put）
	RecordStorageOperation(backend, operation string)
	// RecordStorageLatency 記錄儲存操作延遲
	RecordStorageLatency(backend, operation string, seconds float64)
	// RecordStorageError 記錄儲存錯誤（errorType: timeout/not_found/permission）
	RecordStorageError(backend, errorType string)
	// RecordStorageRetry 記錄儲存重試次數
	RecordStorageRetry(backend string)

	// ======== 安全與風控指標 ========
	// RecordSignatureValidation 記錄簽名驗證結果
	RecordSignatureValidation(success bool)
	// RecordRejectedRequest 記錄被拒絕請求（reason: bad_signature/expired/rate_limited）
	RecordRejectedRequest(reason string)
	// RecordRateLimitHit 記錄觸發流量限制次數
	RecordRateLimitHit()

	// ======== 系統與效能 ========
	// RecordUptimeSeconds 記錄服務運行時間
	RecordUptimeSeconds(seconds float64)

	// ======== Registry ========
	// Registry 取得 Prometheus Registry
	Registry() *prometheus.Registry
}
