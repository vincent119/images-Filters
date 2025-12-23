// Package metrics 提供 Prometheus 指標收集功能
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics 指標介面
type Metrics interface {
	// 請求相關指標
	RecordRequest(method, path string, statusCode int, duration float64)
	RecordImageProcessed(imageType string, sizeBytes int64)
	RecordError(errorType string)

	// 取得 Registry
	Registry() *prometheus.Registry
}
