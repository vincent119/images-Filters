package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vincent119/images-filters/internal/cache"
	"github.com/vincent119/images-filters/internal/config"
)

func main() {
	fmt.Println("Starting Redis Configuration Test...")

	// 1. Load Config
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := runRedisTest(context.Background(), cfg, os.Stdout); err != nil {
		fmt.Printf("❌ Redis test failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("🎉 Redis configuration verification passed!")
}

func runRedisTest(ctx context.Context, cfg *config.Config, w io.Writer) error {
	if !cfg.Cache.Enabled {
		return fmt.Errorf("cache is NOT enabled in config")
	}

	if cfg.Cache.Type != "redis" {
		return fmt.Errorf("cache type is '%s', not 'redis'", cfg.Cache.Type)
	}

	fmt.Fprintf(w, "Configuration:\nHost: %s\nPort: %d\nDB: %d\n",
		cfg.Cache.Redis.Host,
		cfg.Cache.Redis.Port,
		cfg.Cache.Redis.DB,
	)

	// 2. Initialize Redis Cache
	rCache, err := cache.NewRedisCache(cfg.Cache.Redis)
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	fmt.Fprintf(w, "✅ Connected to Redis successfully\n")

	// 3. Test Set
	testKey := fmt.Sprintf("test-key-%d", time.Now().Unix())
	testValue := []byte("hello-redis")

	fmt.Fprintf(w, "Setting key %s...\n", testKey)
	if err := rCache.Set(ctx, testKey, testValue, 10*time.Second); err != nil {
		return fmt.Errorf("failed to set key: %w", err)
	}
	fmt.Fprintf(w, "✅ Set key successful\n")

	// 4. Test Get
	fmt.Fprintf(w, "Getting key %s...\n", testKey)
	val, err := rCache.Get(ctx, testKey)
	if err != nil {
		return fmt.Errorf("failed to get key: %w", err)
	}
	if string(val) != string(testValue) {
		return fmt.Errorf("value mismatch. Expected %s, got %s", string(testValue), string(val))
	}
	fmt.Fprintf(w, "✅ Get key successful\n")

	// 5. Test Delete
	fmt.Fprintf(w, "Deleting key %s...\n", testKey)
	if err := rCache.Delete(ctx, testKey); err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	// Verify Delete
	exists, err := rCache.Exists(ctx, testKey)
	if err != nil {
		return fmt.Errorf("failed to check existence: %w", err)
	}
	if exists {
		return fmt.Errorf("key should not exist after delete")
	}
	fmt.Fprintf(w, "✅ Delete key successful\n")

	return nil
}
