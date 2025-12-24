package fx

import (
	"go.uber.org/fx"

	"github.com/vincent119/images-filters/internal/cache"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/pkg/logger"
)

// CacheModule provides cache dependency
var CacheModule = fx.Module("cache",
	fx.Provide(NewCache),
)

// NewCache creates a new cache instance based on config
func NewCache(cfg *config.Config) (cache.Cache, error) {
	if !cfg.Cache.Enabled {
		logger.Info("cache disabled, using no-op cache")
		return cache.NewNoOpCache(), nil
	}

	switch cfg.Cache.Type {
	case "redis":
		return cache.NewRedisCache(cfg.Cache.Redis)
	case "memory":
		return cache.NewMemoryCache(cfg.Cache.Memory)
	default:
		logger.Warn("unknown cache type, using no-op cache", logger.String("type", cfg.Cache.Type))
		return cache.NewNoOpCache(), nil
	}
}
