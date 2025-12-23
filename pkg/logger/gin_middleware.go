package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vincent119/zlogger"
)

// GinMiddleware 建立 Gin Logger 中介層
// 使用 zlogger 記錄每個 HTTP 請求
func GinMiddleware() gin.HandlerFunc {
	// 不記錄日誌的路徑
	skipPaths := map[string]bool{
		"/healthz": true,
		"/metrics": true,
	}

	// 不記錄日誌的路徑前綴
	skipPathPrefixes := []string{
		"/swagger",
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 跳過特定路徑的日誌記錄
		if skipPaths[path] {
			c.Next()
			return
		}

		// 檢查路徑前綴
		for _, prefix := range skipPathPrefixes {
			if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
				c.Next()
				return
			}
		}

		// 記錄開始時間
		startTime := time.Now()
		query := c.Request.URL.RawQuery

		// 處理請求
		c.Next()

		// 計算處理時間
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 建立日誌欄位
		fields := []zlogger.Field{
			zlogger.Int("status", statusCode),
			zlogger.String("method", method),
			zlogger.String("path", path),
			zlogger.String("ip", clientIP),
			zlogger.String("latency", latency.String()),
			zlogger.Int64("latency_ms", latency.Milliseconds()),
			zlogger.Int("size", c.Writer.Size()),
		}

		// 加入查詢參數（如果有）
		if query != "" {
			fields = append(fields, zlogger.String("query", query))
		}

		// 加入錯誤訊息（如果有）
		if errorMessage != "" {
			fields = append(fields, zlogger.String("error", errorMessage))
		}

		// 根據狀態碼選擇日誌等級
		switch {
		case statusCode >= 500:
			Error("HTTP request", fields...)
		case statusCode >= 400:
			Warn("HTTP request", fields...)
		default:
			Info("", fields...)
		}
	}
}

// GinRecovery 建立 Gin Recovery 中介層
// 使用 zlogger 記錄 panic 錯誤
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 記錄 panic 錯誤
				Error("HTTP request panic",
					zlogger.Any("error", err),
					zlogger.String("method", c.Request.Method),
					zlogger.String("path", c.Request.URL.Path),
					zlogger.String("ip", c.ClientIP()),
				)

				// 返回 500 錯誤
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
