// Package routes 定義 HTTP 路由
package routes

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/vincent119/images-filters/internal/api"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/metrics"
	"github.com/vincent119/images-filters/internal/service"

	// Swagger docs
	_ "github.com/vincent119/images-filters/docs/swagger"
)

// Setup 設定路由
func Setup(engine *gin.Engine, imageService service.ImageService, watermarkService service.WatermarkService, cfg *config.Config, m metrics.Metrics) {
	// 建立處理器
	handler := api.NewHandler(imageService)
	watermarkHandler := api.NewWatermarkHandler(watermarkService)

	// 套用全域中介層
	engine.Use(api.CORSMiddleware())
	engine.Use(api.RecoveryMiddleware())

	// 安全驗證中介層
	engine.Use(api.SecurityMiddleware(&cfg.Security, m))

	// 來源白名單中介層
	engine.Use(api.SourceValidatorMiddleware(cfg.Security.AllowedSources))

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

	// Swagger 端點（如果啟用）
	if cfg.Swagger.Enabled {
		swaggerPath := cfg.Swagger.Path
		if swaggerPath == "" {
			swaggerPath = "/swagger"
		}

		// 如果設定了認證，加入 Basic Auth 中介層
		if cfg.Swagger.Username != "" && cfg.Swagger.Password != "" {
			swaggerGroup := engine.Group(swaggerPath)
			swaggerGroup.Use(swaggerBasicAuth(cfg.Swagger.Username, cfg.Swagger.Password))
			swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		} else {
			engine.GET(swaggerPath+"/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		}
	}

	// 圖片上傳端點（需要 Bearer Auth）
	if cfg.Security.Enabled && cfg.Security.SecurityKey != "" {
		uploadGroup := engine.Group("/upload")
		uploadGroup.Use(api.UploadAuthMiddleware(cfg.Security.SecurityKey, m))
		uploadGroup.POST("", handler.HandleUpload)

		// 浮水印檢測端點（共用 Upload Auth）
		detectGroup := engine.Group("/detect")
		detectGroup.Use(api.UploadAuthMiddleware(cfg.Security.SecurityKey, m))
		detectGroup.POST("", watermarkHandler.HandleDetect)
	} else {
		// 安全機制未啟用時，允許直接上傳（僅開發環境）
		engine.POST("/upload", handler.HandleUpload)
		engine.POST("/detect", watermarkHandler.HandleDetect)
	}

	// 使用 NoRoute 處理所有其他請求（圖片處理）
	// 這樣可以避免萬用字元路由與其他路由衝突
	engine.NoRoute(handler.HandleImage)
}

// swaggerBasicAuth Swagger Basic Auth 中介層
func swaggerBasicAuth(username, password string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, pass, ok := c.Request.BasicAuth()
		if !ok || !secureCompare(user, username) || !secureCompare(pass, password) {
			c.Header("WWW-Authenticate", `Basic realm="Swagger"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

// secureCompare 安全比較字串（防止時序攻擊）
func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
