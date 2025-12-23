package storage

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/vincent119/images-filters/internal/config"
)

// NewStorage creates a storage instance based on configuration
func NewStorage(ctx context.Context, cfg *config.Config) (Storage, error) {
	return NewStorageByType(ctx, cfg, cfg.Storage.Type)
}

// NewStorageByType creates a storage instance by specific type name
// This allows recursive creation for mixed storage
func NewStorageByType(ctx context.Context, cfg *config.Config, storageType string) (Storage, error) {
	switch storageType {
	case "local":
		// Check if config is adequate
		if cfg.Storage.Local.RootPath == "" {
			return nil, fmt.Errorf("local storage root path is required")
		}
		// Ensure absolute path if needed, but NewLocalStorage handles path validation
		// For simplicity, we pass as is, or resolve it.
		// NewLocalStorage in local.go takes rootPath.

		// If using relative path, it is relative to working directory.
		path := cfg.Storage.Local.RootPath
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve local storage path: %w", err)
		}

		return NewLocalStorage(absPath)

	case "s3":
		return NewS3Storage(ctx, cfg.Storage.S3)

	case "mixed":
		// Recursive creation
		sourceType := cfg.Storage.Mixed.SourceStorage
		resultType := cfg.Storage.Mixed.ResultStorage

		if sourceType == "mixed" || resultType == "mixed" {
			return nil, fmt.Errorf("nested mixed storage is not supported")
		}

		source, err := NewStorageByType(ctx, cfg, sourceType)
		if err != nil {
			return nil, fmt.Errorf("failed to create source storage for mixed mode: %w", err)
		}

		result, err := NewStorageByType(ctx, cfg, resultType)
		if err != nil {
			return nil, fmt.Errorf("failed to create result storage for mixed mode: %w", err)
		}

		return NewMixedStorage(source, result), nil

	case "no_storage":
		// Assuming we implement a no-op storage or similar
		// For now, return error or implement minimal dummy
		return nil, fmt.Errorf("no_storage type not implemented yet")

	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}
