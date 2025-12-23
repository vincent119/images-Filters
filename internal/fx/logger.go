package fx

import (
	"context"

	"go.uber.org/fx"

	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/pkg/logger"
)

// LoggerModule provides logger dependency
var LoggerModule = fx.Module("logger",
	fx.Invoke(InitLogger),
)

// InitLogger initializes the logger with lifecycle hooks
func InitLogger(lc fx.Lifecycle, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Init(&cfg.Logging)
			logger.Info("",
				logger.String("msg", "logger initialized"),
				logger.String("level", cfg.Logging.Level),
				logger.String("format", cfg.Logging.Format),
			)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Sync()
			return nil
		},
	})
}
