package main

import (
	"context"
	"fmt"
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

	if !cfg.Cache.Enabled {
		fmt.Println("⚠️  Cache is NOT enabled in config.")
		os.Exit(1)
	}

	if cfg.Cache.Type != "redis" {
		fmt.Printf("⚠️  Cache type is '%s', not 'redis'.\n", cfg.Cache.Type)
		os.Exit(1)
	}

	fmt.Printf("Configuration:\nHost: %s\nPort: %d\nDB: %d\n",
		cfg.Cache.Redis.Host,
		cfg.Cache.Redis.Port,
		cfg.Cache.Redis.DB,
	)

	// 2. Initialize Redis Cache
	rCache, err := cache.NewRedisCache(cfg.Cache.Redis)
	if err != nil {
		fmt.Printf("❌ Failed to connect to Redis: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Connected to Redis successfully")

	// 3. Test Set
	ctx := context.Background()
	testKey := fmt.Sprintf("test-key-%d", time.Now().Unix())
	testValue := []byte("hello-redis")

	fmt.Printf("Setting key %s...\n", testKey)
	if err := rCache.Set(ctx, testKey, testValue, 10*time.Second); err != nil {
		fmt.Printf("❌ Failed to set key: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Set key successful")

	// 4. Test Get
	fmt.Printf("Getting key %s...\n", testKey)
	val, err := rCache.Get(ctx, testKey)
	if err != nil {
		fmt.Printf("❌ Failed to get key: %v\n", err)
		os.Exit(1)
	}
	if string(val) != string(testValue) {
		fmt.Printf("❌ Value mismatch. Expected %s, got %s\n", string(testValue), string(val))
		os.Exit(1)
	}
	fmt.Println("✅ Get key successful")

	// 5. Test Delete
	fmt.Printf("Deleting key %s...\n", testKey)
	if err := rCache.Delete(ctx, testKey); err != nil {
		fmt.Printf("❌ Failed to delete key: %v\n", err)
		os.Exit(1)
	}

	// Verify Delete
	exists, err := rCache.Exists(ctx, testKey)
	if err != nil {
		fmt.Printf("❌ Failed to check existence: %v\n", err)
		os.Exit(1)
	}
	if exists {
		fmt.Println("❌ Key should not exist after delete")
		os.Exit(1)
	}
	fmt.Println("✅ Delete key successful")

	fmt.Println("🎉 Redis configuration verification passed!")
}
