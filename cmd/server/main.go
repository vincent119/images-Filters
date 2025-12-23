// Package main 圖片處理服務器入口
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/metrics"
	"github.com/vincent119/images-filters/internal/service"
	"github.com/vincent119/images-filters/pkg/logger"
	"github.com/vincent119/images-filters/routes"
)

func main() {
	// 載入設定
	cfg, err := config.Load("")
	if err != nil {
		os.Stderr.WriteString("載入設定失敗: " + err.Error() + "\n")
		os.Exit(1)
	}

	// 初始化日誌系統
	logger.Init(&cfg.Logging)
	defer logger.Sync()

	logger.Info("",
		logger.String("msg", "config load success"),
		logger.String("host", cfg.Server.Host),
		logger.Int("port", cfg.Server.Port),
		logger.String("log_level", cfg.Logging.Level),
		logger.String("storage_type", cfg.Storage.Type),
		logger.Bool("metrics_enabled", cfg.Metrics.Enabled),
	)

	// 設定 Gin 模式
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 建立 Gin 引擎（不使用預設中介層）
	engine := gin.New()

	// 使用 zlogger 中介層
	engine.Use(logger.GinMiddleware())
	engine.Use(logger.GinRecovery())

	// 建立 Metrics（如果啟用）
	var m metrics.Metrics
	if cfg.Metrics.Enabled {
		namespace := cfg.Metrics.Namespace
		if namespace == "" {
			namespace = "imgfilter"
		}
		m = metrics.NewPrometheusMetrics(namespace)
		logger.Info("",
			logger.String("msg", "metrics enabled"),
			logger.String("namespace", namespace),
			logger.String("path", cfg.Metrics.Path),
		)
	}

	// 建立服務（傳入 metrics）
	var imageService service.ImageService
	if m != nil {
		imageService = service.NewImageService(cfg, service.WithMetrics(m))
	} else {
		imageService = service.NewImageService(cfg)
	}

	// 設定路由
	routes.Setup(engine, imageService, cfg, m)

	// 建立 HTTP 服務器
	server := &http.Server{
		Addr:         cfg.GetAddress(),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 啟動服務器（非阻塞）
	go func() {
		logger.Info("",
			logger.String("msg", "server start success"),
			logger.String("address", cfg.GetAddress()),
			logger.String("health_endpoint", "/healthz"),
			logger.String("metrics_endpoint", cfg.Metrics.Path),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("",
				logger.String("msg", "server start failed"),
				logger.Err(err))
		}
	}()

	// 等待中斷信號
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("",
		logger.String("msg", "server closing"),
		logger.String("signal", sig.String()),
	)

	// 優雅關閉
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("",
			logger.String("msg", "server closing failed"),
			logger.Err(err))
		os.Exit(1)
	}

	logger.Info("", logger.String("msg", "server closed"))
}
