package fx

import (
	"context"

	"go.uber.org/fx"

	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/storage"
)

// StorageModule provides storage dependencies
var StorageModule = fx.Module("storage",
	fx.Provide(NewStorage),
)

// NewStorage creates a new storage instance from config
func NewStorage(lc fx.Lifecycle, cfg *config.Config) (storage.Storage, error) {
	// Context for storage creation (e.g. AWS session)
	// We use background context as storage initialization might not need timeout control yet
	// or we can use a timeout context if needed.
	ctx := context.Background()

	s, err := storage.NewStorage(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// If needed, we can add lifecycle hooks for proper shutdown if storage supports it
	// lc.Append(fx.Hook{
	// 	OnStop: func(ctx context.Context) error {
	// 		return s.Close()
	// 	},
	// })

	return s, nil
}
