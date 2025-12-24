package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/pkg/logger"
)

// RedisCache 實作 Cache 介面
type RedisCache struct {
	client *redis.Client
	defaultTTL time.Duration
}

// NewRedisCache 建立新的 Redis 快取實例
func NewRedisCache(cfg config.RedisCacheConfig) (*RedisCache, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 測試連線
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	ttl := time.Duration(cfg.TTL) * time.Second
	if ttl == 0 {
		ttl = time.Hour // 預設 1 小時
	}

	logger.Info("redis cache initialized",
		logger.String("addr", addr),
		logger.Int("db", cfg.DB),
	)

	return &RedisCache{
		client:     client,
		defaultTTL: ttl,
	}, nil
}

// Get 取得快取值
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("failed to get from redis: %w", err)
	}
	return val, nil
}

// Set 設定快取值
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = r.defaultTTL
	}

	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set to redis: %w", err)
	}
	return nil
}

// Delete 刪除快取值
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from redis: %w", err)
	}
	return nil
}

// Exists 檢查 Key 是否存在
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence in redis: %w", err)
	}
	return n > 0, nil
}
