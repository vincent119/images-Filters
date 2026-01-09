package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vincent119/images-filters/internal/config"
)

func TestRunS3Test_InvalidConfig(t *testing.T) {
	// Just test that it fails cleanly with invalid credentials/endpoint
	cfg := &config.Config{}
	cfg.Storage.Type = "s3"
	cfg.Storage.S3.Endpoint = "http://invalid-endpoint"

	// This will likely fail at connection attempt or first operation
	err := runS3Test(context.Background(), cfg, os.Stdout)
	assert.Error(t, err)
}
