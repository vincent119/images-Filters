package cache

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/vincent119/images-filters/internal/config"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *RedisCache) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub redis connection", err)
	}

	// Parse host and port from miniredis addr
	parts := strings.Split(mr.Addr(), ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	cfg := config.RedisCacheConfig{
		Host: host,
		Port: port,
		DB:   0,
		Pool: config.RedisPoolConfig{
			Size: 1,
		},
		TTL: 3600,
	}

	cache, err := NewRedisCache(cfg)
	if err != nil {
		t.Fatalf("failed to create redis cache: %v", err)
	}

	return mr, cache
}

func TestRedisCache_Basic(t *testing.T) {
	mr, c := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	key := "test-key"
	value := []byte("hello-redis")

	// Test Set
	err := c.Set(ctx, key, value, 0)
	assert.NoError(t, err)

	// Verify with miniredis directly
	gotVal, err := mr.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, string(value), gotVal)

	// Test Get
	got, err := c.Get(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, value, got)

	// Test Exists
	exists, err := c.Exists(ctx, key)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test Delete
	err = c.Delete(ctx, key)
	assert.NoError(t, err)

	// Verify deletion
	exists, err = c.Exists(ctx, key)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestRedisCache_TTL(t *testing.T) {
	mr, c := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	key := "ttl-key"
	value := []byte("expire-soon")
	ttl := time.Second

	// Set with TTL
	err := c.Set(ctx, key, value, ttl)
	assert.NoError(t, err)

	// Fast forward time in miniredis
	mr.FastForward(2 * time.Second)

	// Should be gone
	_, err = c.Get(ctx, key)
	assert.Equal(t, ErrCacheMiss, err)
}

func TestRedisCache_CacheMiss(t *testing.T) {
	mr, c := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	_, err := c.Get(ctx, "non-existent")
	assert.Equal(t, ErrCacheMiss, err)
}
