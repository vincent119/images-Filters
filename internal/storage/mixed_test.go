package storage

import (
	"context"
	"errors"
	"io"
	"testing"
)

// MockStorage for testing MixedStorage
type MockStorage struct {
	data map[string][]byte
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string][]byte),
	}
}

func (m *MockStorage) Get(ctx context.Context, key string) ([]byte, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, errors.New("not found")
}

func (m *MockStorage) Put(ctx context.Context, key string, data []byte) error {
	m.data[key] = data
	return nil
}

func (m *MockStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *MockStorage) GetStream(ctx context.Context, key string) (io.ReadCloser, error) {
	// For testing MixedStorage, we might need real implementation if tested?
	// But current tests use []byte Get/Put.
	// So unimplemented is fine for now unless we add stream tests.
	return nil, errors.New("not implemented")
}

func (m *MockStorage) PutStream(ctx context.Context, key string, r io.Reader) error {
	return errors.New("not implemented")
}

func TestMixedStorage(t *testing.T) {
	source := NewMockStorage()
	result := NewMockStorage()
	mixed := NewMixedStorage(source, result)

	ctx := context.Background()
	sourceKey := "source.jpg"
	resultKey := "result.jpg"
	testData := []byte("data")

	// Setup: Put data in source only
	source.Put(ctx, sourceKey, testData)

	// Test Get: Should retrieve from source if not in result
	t.Run("Get from Source", func(t *testing.T) {
		data, err := mixed.Get(ctx, sourceKey)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if string(data) != string(testData) {
			t.Errorf("Got %s, want %s", data, testData)
		}
	})

	// Test Get: Should retrieve from result if present
	t.Run("Get from Result", func(t *testing.T) {
		// Put data in result
		result.Put(ctx, resultKey, testData)

		data, err := mixed.Get(ctx, resultKey)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if string(data) != string(testData) {
			t.Errorf("Got %s, want %s", data, testData)
		}
	})

	// Test Put: Should put to result only
	t.Run("Put", func(t *testing.T) {
		key := "new.jpg"
		err := mixed.Put(ctx, key, testData)
		if err != nil {
			t.Errorf("Put failed: %v", err)
		}

		// Check result
		if _, ok := result.data[key]; !ok {
			t.Error("Data not in result storage")
		}
		// Check source
		if _, ok := source.data[key]; ok {
			t.Error("Data unexpectedly in source storage")
		}
	})

	// Test Exists: Should check both
	t.Run("Exists", func(t *testing.T) {
		// Existing in source
		exists, _ := mixed.Exists(ctx, sourceKey)
		if !exists {
			t.Error("Should find key in source")
		}

		// Existing in result
		exists, _ = mixed.Exists(ctx, resultKey)
		if !exists {
			t.Error("Should find key in result")
		}

		// Non-existent
		exists, _ = mixed.Exists(ctx, "nonexistent")
		if exists {
			t.Error("Should not find non-existent key")
		}
	})

	// Test Delete: Should delete from result only
	t.Run("Delete", func(t *testing.T) {
		// Setup: Key in both
		key := "shared.jpg"
		source.Put(ctx, key, testData)
		result.Put(ctx, key, testData)

		err := mixed.Delete(ctx, key)
		if err != nil {
			t.Errorf("Delete failed: %v", err)
		}

		// Check result: gone
		if _, ok := result.data[key]; ok {
			t.Error("Key should be deleted from result")
		}

		// Check source: remains
		if _, ok := source.data[key]; !ok {
			t.Error("Key should remain in source")
		}
	})
}
