package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/security"
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

// SecurityMiddleware 安全驗證中介層
// 驗證 HMAC 簽名或允許 unsafe 路徑
func SecurityMiddleware(cfg *config.SecurityConfig) gin.HandlerFunc {
	var signer *security.Signer
	if cfg.Enabled && cfg.SecurityKey != "" {
		signer = security.NewSigner(cfg.SecurityKey)
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 跳過非圖片處理路徑
		if isSkippedPath(path) {
			c.Next()
			return
		}

		// 安全機制未啟用時，允許所有請求
		if !cfg.Enabled {
			c.Next()
			return
		}

		// 處理 unsafe 路徑
		if security.IsUnsafePath(path) {
			if cfg.AllowUnsafe {
				c.Next()
				return
			}
			// 禁止 unsafe 路徑
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "FORBIDDEN",
				"message": "Unsafe access is disabled",
			})
			c.Abort()
			return
		}

		// 驗證 HMAC 簽名
		if signer == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "CONFIG_ERROR",
				"message": "Security key not configured",
			})
			c.Abort()
			return
		}

		signature, imagePath, ok := security.ExtractSignatureAndPath(path)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "INVALID_SIGNATURE",
				"message": "Invalid URL format",
			})
			c.Abort()
			return
		}

		if !signer.Verify(signature, imagePath) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "INVALID_SIGNATURE",
				"message": "Signature verification failed",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isSkippedPath 檢查是否為不需要安全驗證的路徑
func isSkippedPath(path string) bool {
	skippedPaths := []string{
		"/healthz",
		"/metrics",
		"/swagger",
	}

	for _, p := range skippedPaths {
		if path == p || len(path) > len(p) && path[:len(p)+1] == p+"/" {
			return true
		}
	}

	return false
}

