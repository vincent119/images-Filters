package fx

import (
	"go.uber.org/fx"

	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/metrics"
	"github.com/vincent119/images-filters/pkg/logger"
)

// MetricsModule provides metrics dependency
var MetricsModule = fx.Module("metrics",
	fx.Provide(NewMetrics),
)

// MetricsResult metrics module result
type MetricsResult struct {
	fx.Out

	Metrics metrics.Metrics `optional:"true"`
}

// NewMetrics creates a new metrics instance (if enabled)
func NewMetrics(cfg *config.Config) MetricsResult {
	if !cfg.Metrics.Enabled {
		return MetricsResult{}
	}

	namespace := cfg.Metrics.Namespace
	if namespace == "" {
		namespace = "imgfilter"
	}

	m := metrics.NewPrometheusMetrics(namespace)

	logger.Info("",
		logger.String("msg", "metrics enabled"),
		logger.String("namespace", namespace),
		logger.String("path", cfg.Metrics.Path),
	)

	return MetricsResult{Metrics: m}
}
