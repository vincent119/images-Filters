package metrics

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler 建立 /metrics 端點處理器
// 支援 Basic Auth 認證
func MetricsHandler(m Metrics, username, password string) gin.HandlerFunc {
	// 建立 Prometheus HTTP 處理器
	handler := promhttp.HandlerFor(m.Registry(), promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})

	return func(c *gin.Context) {
		// 如果設定了認證，進行 Basic Auth 驗證
		if username != "" && password != "" {
			user, pass, ok := c.Request.BasicAuth()
			if !ok || !secureCompare(user, username) || !secureCompare(pass, password) {
				c.Header("WWW-Authenticate", `Basic realm="Metrics"`)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}

		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// secureCompare 安全比較字串（防止時序攻擊）
func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
