package fx

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/metrics"
	"github.com/vincent119/images-filters/internal/service"
	"github.com/vincent119/images-filters/pkg/logger"
	"github.com/vincent119/images-filters/routes"
)

// ServerModule provides HTTP server dependencies
var ServerModule = fx.Module("server",
	fx.Provide(NewGinEngine),
	fx.Provide(NewHTTPServer),
	fx.Invoke(RegisterRoutes),
	fx.Invoke(StartServer),
)

// GinEngineParams gin engine parameters
type GinEngineParams struct {
	fx.In

	Config *config.Config
}

// NewGinEngine creates a new Gin engine
func NewGinEngine(params GinEngineParams) *gin.Engine {
	// Set Gin mode
	if params.Config.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(logger.GinMiddleware())
	engine.Use(logger.GinRecovery())

	return engine
}

// HTTPServerParams HTTP server parameters
type HTTPServerParams struct {
	fx.In

	Config *config.Config
	Engine *gin.Engine
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(params HTTPServerParams) *http.Server {
	return &http.Server{
		Addr:         params.Config.GetAddress(),
		Handler:      params.Engine,
		ReadTimeout:  params.Config.Server.ReadTimeout,
		WriteTimeout: params.Config.Server.WriteTimeout,
	}
}

// RouteParams route setup parameters
type RouteParams struct {
	fx.In

	Engine       *gin.Engine
	ImageService     service.ImageService
	WatermarkService service.WatermarkService
	Config           *config.Config
	Metrics          metrics.Metrics `optional:"true"`
}

// RegisterRoutes registers all routes
func RegisterRoutes(params RouteParams) {
	routes.Setup(params.Engine, params.ImageService, params.WatermarkService, params.Config, params.Metrics)
}

// ServerStartParams server start parameters
type ServerStartParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Server    *http.Server
	Config    *config.Config
}

// StartServer starts the HTTP server with lifecycle hooks
func StartServer(params ServerStartParams) {
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("",
				logger.String("msg", "server starting"),
				logger.String("address", params.Config.GetAddress()),
				logger.String("health_endpoint", "/healthz"),
				logger.String("metrics_endpoint", params.Config.Metrics.Path),
			)

			go func() {
				if err := params.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal("",
						logger.String("msg", "server start failed"),
						logger.Err(err),
					)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("", logger.String("msg", "server stopping"))

			if err := params.Server.Shutdown(ctx); err != nil {
				logger.Error("",
					logger.String("msg", "server shutdown failed"),
					logger.Err(err),
				)
				return err
			}

			logger.Info("", logger.String("msg", "server stopped"))
			return nil
		},
	})
}
