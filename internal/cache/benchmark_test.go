package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/vincent119/images-filters/internal/config"
)

func BenchmarkMemoryCache_Set(b *testing.B) {
	c, err := NewMemoryCache(config.MemoryCacheConfig{MaxSize: 100 * 1024 * 1024, TTL: 60})
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	data := []byte("test-data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		err := c.Set(ctx, key, data, time.Minute)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemoryCache_Get(b *testing.B) {
	c, err := NewMemoryCache(config.MemoryCacheConfig{MaxSize: 100 * 1024 * 1024, TTL: 60})
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	data := []byte("test-data")
	key := "test-key"
	c.Set(ctx, key, data, time.Minute)
	time.Sleep(100 * time.Millisecond) // Wait for Ristretto admission

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.Get(ctx, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemoryCache_SetGet_Parallel(b *testing.B) {
	c, err := NewMemoryCache(config.MemoryCacheConfig{MaxSize: 100 * 1024 * 1024, TTL: 60})
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	data := []byte("test-data")

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			key := fmt.Sprintf("key-%d", i%1000) // Reuse keys
			if i%2 == 0 {
				c.Set(ctx, key, data, time.Minute)
			} else {
				c.Get(ctx, key)
			}
		}
	})
}
