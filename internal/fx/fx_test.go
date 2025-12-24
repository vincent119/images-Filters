package fx

import (
	"testing"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/vincent119/images-filters/internal/cache"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/service"
	"github.com/vincent119/images-filters/internal/storage"
)

// TestServiceModule verifies that the ServiceModule and its dependencies can be provided.
func TestServiceModule(t *testing.T) {
	cfg := &config.Config{
		Processing: config.ProcessingConfig{
			DefaultQuality: 80,
			MaxWidth:       1000,
			MaxHeight:      1000,
			DefaultFormat:  "jpeg",
		},
		Storage: config.StorageConfig{
			Type: "local",
			Local: config.LocalStorageConfig{
				RootPath: "/tmp/test-storage",
			},
		},
		Cache: config.CacheConfig{
			Enabled: false, // Use NoOp cache
			Type:    "memory",
		},
	}

	app := fxtest.New(t,
		// Provide the test config
		fx.Supply(cfg),
		// Provide all modules involved in Service construction
		StorageModule,
		CacheModule,
		ServiceModule,
		// Invoke the service to ensure it can be constructed
		fx.Invoke(func(s service.ImageService) {
			if s == nil {
				t.Error("ImageService is nil")
			}
		}),
	)

	// ValidateApp checks if the dependency graph is valid
	if err := app.Err(); err != nil {
		t.Fatalf("FX app validation failed: %v", err)
	}
}

// TestStorageModule verifies StorageModule independently
func TestStorageModule(t *testing.T) {
	cfg := &config.Config{
		Storage: config.StorageConfig{
			Type: "local",
			Local: config.LocalStorageConfig{
				RootPath: "/tmp/test-storage",
			},
		},
	}

	app := fxtest.New(t,
		fx.Supply(cfg),
		StorageModule,
		fx.Invoke(func(s storage.Storage) {
			if s == nil {
				t.Error("Storage is nil")
			}
		}),
	)
	if err := app.Err(); err != nil {
		t.Fatalf("StorageModule validation failed: %v", err)
	}
}

// TestCacheModule verifies CacheModule independently
func TestCacheModule(t *testing.T) {
	cfg := &config.Config{
		Cache: config.CacheConfig{
			Enabled: true,
			Type:    "memory",
			Memory: config.MemoryCacheConfig{
				MaxSize: 10 * 1024 * 1024, // 10MB
				TTL:     60,
			},
		},
	}

	app := fxtest.New(t,
		fx.Supply(cfg),
		CacheModule,
		fx.Invoke(func(c cache.Cache) {
			if c == nil {
				t.Error("Cache is nil")
			}
		}),
	)
	if err := app.Err(); err != nil {
		t.Fatalf("CacheModule validation failed: %v", err)
	}
}
