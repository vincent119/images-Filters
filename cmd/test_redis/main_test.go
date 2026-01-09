package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vincent119/images-filters/internal/config"
)

func TestRunRedisTest_Disabled(t *testing.T) {
	cfg := &config.Config{}
	cfg.Cache.Enabled = false

	err := runRedisTest(context.Background(), cfg, os.Stdout)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "NOT enabled")
}

func TestRunRedisTest_WrongType(t *testing.T) {
	cfg := &config.Config{}
	cfg.Cache.Enabled = true
	cfg.Cache.Type = "memory"

	err := runRedisTest(context.Background(), cfg, os.Stdout)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not 'redis'")
}

// Note: Running a full success test requires a real Redis instance.
// We can mock cache.NewRedisCache if we refactor injection, but that might be overkill for this tool.
// For now, these tests cover the configuration checks which accounts for some code paths.
