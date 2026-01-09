package fx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/vincent119/images-filters/internal/cache"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/metrics"
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

func TestLoggerModule(t *testing.T) {
	cfg := &config.Config{
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "json",
		},
	}

	app := fxtest.New(t,
		fx.Supply(cfg),
		LoggerModule,
	)
	app.RequireStart()
	app.RequireStop()
}

func TestMetricsModule(t *testing.T) {
	t.Run("Enabled", func(t *testing.T) {
		cfg := &config.Config{
			Metrics: config.MetricsConfig{
				Enabled:   true,
				Namespace: "test",
			},
		}
		app := fxtest.New(t,
			fx.Supply(cfg),
			MetricsModule,
			fx.Invoke(func(m metrics.Metrics) {
				assert.NotNil(t, m)
			}),
		)
		app.RequireStart()
		app.RequireStop()
	})

	t.Run("Disabled", func(t *testing.T) {
		cfg := &config.Config{
			Metrics: config.MetricsConfig{
				Enabled: false,
			},
		}
		app := fxtest.New(t,
			fx.Supply(cfg),
			MetricsModule,
			// When disabled, Provide returns MetricsResult with nil Metrics.
			// fx.Invoke(func(m metrics.Metrics)) might receive nil if it's strictly typed?
			// Actually NewMetrics returns MetricsResult{Metrics: m} where m is nil/valid.
			// Fx handles struct embedding.
			fx.Invoke(func(m metrics.Metrics) {
				// If m is interface, it might be nil
				assert.Nil(t, m)
			}),
		)
		if err := app.Err(); err != nil {
			t.Fatalf("App error: %v", err)
		}
	})
}

func TestNewWatermarkService(t *testing.T) {
	// Testing the provider function wrapper
	cfg := &config.Config{}
	store := storage.NewNoStorage()

	svc := NewWatermarkService(cfg, store)
	assert.NotNil(t, svc)
}

// Mock services for Fx test
type mockImgSvc struct {
	service.ImageService
}
type mockWmSvc struct {
	service.WatermarkService
}

func TestServerModule(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: 0, // Random port
		},
		Logging: config.LoggingConfig{
			Level: "info",
		},
	}

	app := fxtest.New(t,
		fx.Supply(cfg),
		fx.Provide(func() service.ImageService { return &mockImgSvc{} }),
		fx.Provide(func() service.WatermarkService { return &mockWmSvc{} }),
		fx.Provide(func() metrics.Metrics { return nil }), // Optional in ServerModule?
		ServerModule,
	)

	// Verify it starts (binds port, registers routes)
	app.RequireStart()

	// Verify OnStop works
	app.RequireStop()
}
