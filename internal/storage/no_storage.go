package storage

import (
	"context"

	"github.com/vincent119/images-filters/internal/storage/types"
)

// NoStorage implements a no-op storage backend
// Used for testing, benchmarking, or when persistence is not required
type NoStorage struct{}

// NewNoStorage creates a new NoStorage instance
func NewNoStorage() *NoStorage {
	return &NoStorage{}
}

// Get always returns ErrNotFound as it stores nothing
func (s *NoStorage) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, types.ErrNotFound
}

// Put does nothing and returns nil (simulating success)
func (s *NoStorage) Put(ctx context.Context, key string, data []byte) error {
	return nil
}

// Exists always returns false
func (s *NoStorage) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

// Delete does nothing and returns nil
func (s *NoStorage) Delete(ctx context.Context, key string) error {
	return nil
}
