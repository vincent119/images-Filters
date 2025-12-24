package cache

import (
	"context"
	"testing"
	"time"

	"github.com/vincent119/images-filters/internal/config"
)

func TestMemoryCache_Basic(t *testing.T) {
	cfg := config.MemoryCacheConfig{
		MaxSize: 10 * 1024 * 1024, // 10 MB
		TTL:     3600,
	}

	c, err := NewMemoryCache(cfg)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}
	defer c.Close()

	ctx := context.Background()
	key := "test-key"
	value := []byte("hello-world")

	// Test Set
	if err := c.Set(ctx, key, value, 0); err != nil {
		t.Errorf("Set failed: %v", err)
	}

	// Wait for Ristretto's eventual consistency (Set is async)
	time.Sleep(100 * time.Millisecond)

	// Test Get
	got, err := c.Get(ctx, key)
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("Expected %s, got %s", value, got)
	}

	// Test Exists
	exists, err := c.Exists(ctx, key)
	if err != nil {
		t.Errorf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected key to exist")
	}

	// Test Delete
	if err := c.Delete(ctx, key); err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Wait for Ristretto
	time.Sleep(100 * time.Millisecond)

	exists, err = c.Exists(ctx, key)
	if err != nil {
		t.Errorf("Exists check failed: %v", err)
	}
	if exists {
		t.Error("Expected key to be deleted")
	}
}

func TestMemoryCache_TTL(t *testing.T) {
	cfg := config.MemoryCacheConfig{
		MaxSize: 10 * 1024 * 1024,
		TTL:     1, // 1 second default
	}

	c, err := NewMemoryCache(cfg)
	if err != nil {
		t.Fatalf("Failed to create memory cache: %v", err)
	}
	defer c.Close()

	ctx := context.Background()
	key := "ttl-key"
	value := []byte("expire-soon")

	// Set with 500ms TTL
	if err := c.Set(ctx, key, value, 500*time.Millisecond); err != nil {
		t.Errorf("Set failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Should exist
	got, err := c.Get(ctx, key)
	if err != nil {
		t.Errorf("Expected to get value before expiration, got err: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("Value mismatch")
	}

	// Wait for expiration
	time.Sleep(600 * time.Millisecond)

	// Should not exist
	_, err = c.Get(ctx, key)
	if err != ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss after expiration, got: %v", err)
	}
}
