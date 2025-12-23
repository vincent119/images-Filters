package storage

import (
	"context"
)

// MixedStorage implements hybrid storage strategy
// usually used for separating source images (Read-only/Source) and processed images (Read-Write/Result)
type MixedStorage struct {
	source Storage // Source storage (e.g. S3 for original images)
	result Storage // Result storage (e.g. Local or S3 for processed/cached images)
}

// NewMixedStorage creates a new mixed storage instance
func NewMixedStorage(source Storage, result Storage) *MixedStorage {
	return &MixedStorage{
		source: source,
		result: result,
	}
}

// Get tries to get from result storage first (cache hit), then source storage
// But usually mixed storage logic depends on the use case.
// If it's "Source vs Result", usually we want to get the requested key.
// If the key represents a processed image, it should be in Result.
// If the key represents a source image, it might be in Source.
// However, the `Storage` interface is generic.
// Strategy:
// 1. Try Result storage.
// 2. If not found, try Source storage.
// This acts like a transparent overlay.
func (s *MixedStorage) Get(ctx context.Context, key string) ([]byte, error) {
	// Try result storage first
	data, err := s.result.Get(ctx, key)
	if err == nil {
		return data, nil
	}

	// Internal error or not found in result, try source
	// We might want to distinguish between "not found" and "error"
	// But simply falling back is a robust simple strategy for "Read"
	return s.source.Get(ctx, key)
}

// Put saves to Result storage only. Source is considered read-only in this pattern.
func (s *MixedStorage) Put(ctx context.Context, key string, data []byte) error {
	return s.result.Put(ctx, key, data)
}

// Exists checks Result then Source
func (s *MixedStorage) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := s.result.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	return s.source.Exists(ctx, key)
}

// Delete removes from Result storage.
// Optionally, we could decide if we want to delete from Source too,
// but for "Source/Result" separation, usually Source is immutable or managed separately.
// Let's safe delete from Result only for now to protect Source.
func (s *MixedStorage) Delete(ctx context.Context, key string) error {
	// Only delete from result storage to protect source
	return s.result.Delete(ctx, key)
}
