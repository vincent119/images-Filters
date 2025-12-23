package fx

import (
	"go.uber.org/fx"

	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/metrics"
	"github.com/vincent119/images-filters/internal/service"
)

// ServiceModule provides service dependencies
var ServiceModule = fx.Module("service",
	fx.Provide(NewImageService),
)

// ServiceParams service module parameters
type ServiceParams struct {
	fx.In

	Config  *config.Config
	Metrics metrics.Metrics `optional:"true"`
}

// NewImageService creates a new image service instance
func NewImageService(params ServiceParams) service.ImageService {
	opts := make([]service.ServiceOption, 0)

	if params.Metrics != nil {
		opts = append(opts, service.WithMetrics(params.Metrics))
	}

	return service.NewImageService(params.Config, opts...)
}
