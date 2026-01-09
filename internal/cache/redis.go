package cache

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/pkg/logger"
)

// RedisCache 實作 Cache 介面
type RedisCache struct {
	client     *redis.Client
	defaultTTL time.Duration
}

// NewRedisCache 建立新的 Redis 快取實例
// NewRedisCache 建立新的 Redis 快取實例
func NewRedisCache(cfg config.RedisCacheConfig) (*RedisCache, error) {
	opts, err := getRedisOptions(cfg)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

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
		logger.String("addr", opts.Addr),
		logger.Int("db", opts.DB),
		logger.Int("pool_size", opts.PoolSize),
		logger.Int("min_idle_conns", opts.MinIdleConns),
		logger.Int("max_idle_conns", opts.MaxIdleConns),
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

func getRedisOptions(cfg config.RedisCacheConfig) (*redis.Options, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// 設定連線池參數
	poolSize := cfg.Pool.Size
	if poolSize == 0 {
		poolSize = 10 // 預設 10 個連線
	}

	minIdleConns := cfg.Pool.MinIdleConns
	if minIdleConns == 0 {
		minIdleConns = 2 // 預設最小 2 個閒置連線
	}

	maxIdleConns := cfg.Pool.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 5 // 預設最大 5 個閒置連線
	}

	poolTimeout := time.Duration(cfg.Pool.Timeout) * time.Second
	if poolTimeout == 0 {
		poolTimeout = 4 * time.Second
	}

	connTimeout := time.Duration(cfg.Pool.ConnTimeout) * time.Second
	if connTimeout == 0 {
		connTimeout = 5 * time.Second
	}

	readTimeout := time.Duration(cfg.Pool.ReadTimeout) * time.Second
	if readTimeout == 0 {
		readTimeout = 3 * time.Second
	}

	writeTimeout := time.Duration(cfg.Pool.WriteTimeout) * time.Second
	if writeTimeout == 0 {
		writeTimeout = 3 * time.Second
	}

	tlsConfig, err := getTLSConfig(cfg.TLS)
	if err != nil {
		return nil, err
	}

	return &redis.Options{
		Addr:         addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxIdleConns: maxIdleConns,
		PoolTimeout:  poolTimeout,
		DialTimeout:  connTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		TLSConfig:    tlsConfig,
	}, nil
}

func getTLSConfig(cfg config.RedisTLSConfig) (*tls.Config, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.Insecure,
	}

	// 載入 CA 證書
	if cfg.CAFile != "" {
		caCert, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert file: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA cert")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// 載入客戶端證書（如果有提供）
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert/key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	logger.Info("redis TLS enabled",
		logger.Bool("insecure", cfg.Insecure),
		logger.String("ca_cert", cfg.CAFile),
	)

	return tlsConfig, nil
}
