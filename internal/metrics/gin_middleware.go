package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GinMiddleware 建立 Gin Metrics 中介層
// 收集 HTTP 請求相關指標
func GinMiddleware(m Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳過 /metrics 和 /healthz 路徑
		if c.Request.URL.Path == "/metrics" || c.Request.URL.Path == "/healthz" {
			c.Next()
			return
		}

		// 記錄進行中請求數 (+1)
		m.RecordInflightRequest(1)
		defer m.RecordInflightRequest(-1)

		start := time.Now()

		// 記錄請求大小
		path := normalizePath(c.Request.URL.Path)
		if c.Request.ContentLength > 0 {
			m.RecordRequestSize(c.Request.Method, path, c.Request.ContentLength)
		}

		// 處理請求
		c.Next()

		// 記錄指標
		duration := time.Since(start).Seconds()
		statusCode := c.Writer.Status()

		m.RecordRequest(c.Request.Method, path, statusCode, duration)

		// 記錄回應大小
		responseSize := int64(c.Writer.Size())
		if responseSize > 0 {
			m.RecordResponseSize(c.Request.Method, path, responseSize)
		}

		// 如果是錯誤回應，記錄錯誤
		if statusCode >= 400 {
			errorType := "http_" + strconv.Itoa(statusCode)
			m.RecordError(errorType)
		}
	}
}

// normalizePath 標準化路徑，避免高基數問題
// 將動態路徑參數替換為佔位符
func normalizePath(path string) string {
	// 對於圖片處理路由，統一標記
	if len(path) > 1 && path != "/healthz" && path != "/metrics" {
		// 檢查是否為 unsafe 路由
		if len(path) > 7 && path[:7] == "/unsafe" {
			return "/unsafe/*"
		}
		// 其他圖片處理路由（帶簽名）
		return "/*"
	}
	return path
}
