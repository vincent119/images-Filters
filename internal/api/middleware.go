package api

import (
	"time"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware CORS 中介層
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// LoggingMiddleware 日誌中介層
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 開始時間
		startTime := time.Now()

		// 處理請求
		c.Next()

		// 計算處理時間
		latency := time.Since(startTime)

		// 記錄日誌（使用 Gin 預設的日誌格式）
		gin.DefaultWriter.Write([]byte(
			time.Now().Format("2006/01/02 - 15:04:05") +
				" | " + c.ClientIP() +
				" | " + c.Request.Method +
				" | " + c.Request.URL.Path +
				" | " + latency.String() +
				" | " + string(rune(c.Writer.Status())) +
				"\n",
		))
	}
}

// RecoveryMiddleware 錯誤恢復中介層
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.Recovery()
}
