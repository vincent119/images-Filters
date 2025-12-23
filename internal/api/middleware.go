package api

import (
	"net/http"
	"strings"
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

// SourceValidatorMiddleware 來源白名單驗證中介層
// 檢查圖片來源是否在允許的白名單中
func SourceValidatorMiddleware(allowedSources []string) gin.HandlerFunc {
	validator := security.NewSourceValidator(allowedSources)

	return func(c *gin.Context) {
		// 白名單未啟用時，允許所有請求
		if !validator.IsEnabled() {
			c.Next()
			return
		}

		// 跳過非圖片處理路徑
		path := c.Request.URL.Path
		if isSkippedPath(path) {
			c.Next()
			return
		}

		// 從 URL 中提取圖片來源
		// 圖片路徑可能是 HTTP URL 或本地路徑
		imagePath := extractImagePath(path)
		if imagePath == "" {
			c.Next()
			return
		}

		// 只對 HTTP URL 進行來源驗證
		if !isHTTPURL(imagePath) {
			c.Next()
			return
		}

		// 驗證來源
		if !validator.IsAllowed(imagePath) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "SOURCE_NOT_ALLOWED",
				"message": "Image source is not in the allowed list",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractImagePath 從 URL 路徑中提取圖片路徑
func extractImagePath(path string) string {
	// 移除開頭斜線
	path = strings.TrimPrefix(path, "/")

	// 分割路徑段
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return ""
	}

	// 跳過 unsafe 或簽名
	startIdx := 1
	if parts[0] == "unsafe" {
		startIdx = 1
	}

	// 找到圖片路徑（通常是最後的部分）
	for i := startIdx; i < len(parts); i++ {
		part := parts[i]
		// 檢查是否為 HTTP URL（URL 編碼或明文）
		if strings.HasPrefix(part, "http") || strings.Contains(part, "%3A%2F%2F") {
			return strings.Join(parts[i:], "/")
		}
	}

	// 返回最後一個部分
	return parts[len(parts)-1]
}

// isHTTPURL 檢查是否為 HTTP URL
func isHTTPURL(s string) bool {
	s = strings.ToLower(s)
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "http%3a%2f%2f") || strings.HasPrefix(s, "https%3a%2f%2f")
}
