// Package routes 定義 HTTP 路由
package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/vincent119/images-filters/internal/api"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/metrics"
	"github.com/vincent119/images-filters/internal/service"
)

// Setup 設定路由
func Setup(engine *gin.Engine, imageService service.ImageService, cfg *config.Config, m metrics.Metrics) {
	// 建立處理器
	handler := api.NewHandler(imageService)

	// 套用全域中介層
	engine.Use(api.CORSMiddleware())
	engine.Use(api.RecoveryMiddleware())

	// 如果啟用 Metrics，加入 Metrics 中介層
	if cfg.Metrics.Enabled && m != nil {
		engine.Use(metrics.GinMiddleware(m))
	}

	// 健康檢查端點
	engine.GET("/healthz", handler.HealthCheck)

	// Metrics 端點（如果啟用）
	if cfg.Metrics.Enabled && m != nil {
		metricsPath := cfg.Metrics.Path
		if metricsPath == "" {
			metricsPath = "/metrics"
		}
		engine.GET(metricsPath, metrics.MetricsHandler(m, cfg.Metrics.Username, cfg.Metrics.Password))
	}

	// 使用 NoRoute 處理所有其他請求（圖片處理）
	// 這樣可以避免萬用字元路由與其他路由衝突
	engine.NoRoute(handler.HandleImage)
}
