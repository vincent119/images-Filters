// Package fx provides Uber FX dependency injection modules
package fx

import (
	"go.uber.org/fx"

	"github.com/vincent119/images-filters/internal/config"
)

// ConfigModule provides configuration dependency
var ConfigModule = fx.Module("config",
	fx.Provide(NewConfig),
)

// ConfigParams config module parameters
type ConfigParams struct {
	fx.In

	ConfigPath string `name:"config_path" optional:"true"`
}

// NewConfig creates a new config instance
func NewConfig(params ConfigParams) (*config.Config, error) {
	return config.Load(params.ConfigPath)
}
