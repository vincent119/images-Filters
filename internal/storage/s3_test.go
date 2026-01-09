package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// MockS3API mocks S3API interface for testing
type MockS3API struct {
	mu      sync.RWMutex
	objects map[string][]byte
}

func NewMockS3API() *MockS3API {
	return &MockS3API{
		objects: make(map[string][]byte),
	}
}

func (m *MockS3API) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := *params.Key
	data, exists := m.objects[key]
	if !exists {
		return nil, &types.NoSuchKey{}
	}

	return &s3.GetObjectOutput{
		Body:          io.NopCloser(bytes.NewReader(data)),
		ContentLength: aws.Int64(int64(len(data))),
		ContentType:   nil, // Optional: verify content type logic if needed
	}, nil
}

func (m *MockS3API) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := *params.Key
	data, err := io.ReadAll(params.Body)
	if err != nil {
		return nil, err
	}
	m.objects[key] = data

	return &s3.PutObjectOutput{}, nil
}

func (m *MockS3API) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := *params.Key
	data, exists := m.objects[key]
	if !exists {
		// StatusCode 404
		return nil, &httpResponseError{statusCode: 404}
	}

	return &s3.HeadObjectOutput{
		ContentLength: aws.Int64(int64(len(data))),
	}, nil
}

type httpResponseError struct {
	statusCode int
}

func (e *httpResponseError) Error() string {
	return fmt.Sprintf("http status %d", e.statusCode)
}

func (e *httpResponseError) HTTPStatusCode() int {
	return e.statusCode
}

func (m *MockS3API) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := *params.Key
	delete(m.objects, key)

	return &s3.DeleteObjectOutput{}, nil
}

func TestS3Storage(t *testing.T) {
	mockAPI := NewMockS3API()
	storage := &S3Storage{
		client: mockAPI,
		bucket: "test-bucket",
	}

	ctx := context.Background()
	testKey := "test/image.jpg"
	testData := []byte("test image data")

	// Test Put
	t.Run("Put", func(t *testing.T) {
		err := storage.Put(ctx, testKey, testData)
		if err != nil {
			t.Errorf("Put() error = %v", err)
		}

		// Verify data in mock
		mockAPI.mu.RLock()
		storedData, exists := mockAPI.objects[testKey]
		mockAPI.mu.RUnlock()

		if !exists {
			t.Error("object should exist in mock")
		}
		if !bytes.Equal(storedData, testData) {
			t.Errorf("stored data mismatch")
		}
	})

	// Test Exists
	t.Run("Exists", func(t *testing.T) {
		exists, err := storage.Exists(ctx, testKey)
		if err != nil {
			t.Errorf("Exists() error = %v", err)
		}
		if !exists {
			t.Error("Exists() = false; want true")
		}

		exists, err = storage.Exists(ctx, "nonexistent.jpg")
		if err != nil {
			t.Errorf("Exists() error = %v", err)
		}
		if exists {
			t.Error("Exists() = true; want false")
		}
	})

	// Test Get
	t.Run("Get", func(t *testing.T) {
		data, err := storage.Get(ctx, testKey)
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}
		if !bytes.Equal(data, testData) {
			t.Errorf("Get() = %s; want %s", data, testData)
		}

		_, err = storage.Get(ctx, "nonexistent.jpg")
		if err == nil {
			t.Error("Get() should return error for nonexistent key")
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		err := storage.Delete(ctx, testKey)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		exists, _ := storage.Exists(ctx, testKey)
		if exists {
			t.Error("object should be deleted")
		}
	})

	// Test Stream
	t.Run("Stream", func(t *testing.T) {
		streamKey := "stream/test.jpg"
		streamData := []byte("stream data")

		// PutStream
		err := storage.PutStream(ctx, streamKey, bytes.NewReader(streamData))
		if err != nil {
			t.Errorf("PutStream() error = %v", err)
		}

		// Verify PutStream result via normal Get
		mockAPI.mu.RLock()
		storedData, exists := mockAPI.objects[streamKey]
		mockAPI.mu.RUnlock()
		if !exists {
			t.Error("PutStream object should exist in mock")
		}
		if !bytes.Equal(storedData, streamData) {
			t.Errorf("PutStream stored data mismatch")
		}

		// GetStream
		rc, err := storage.GetStream(ctx, streamKey)
		if err != nil {
			t.Fatalf("GetStream() error = %v", err)
		}
		defer rc.Close()

		readData, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("ReadAll error = %v", err)
		}
		if !bytes.Equal(readData, streamData) {
			t.Errorf("GetStream data mismatch")
		}

		// GetStream Not Found
		_, err = storage.GetStream(ctx, "nonexistent-stream")
		if err == nil {
			t.Error("GetStream should error on not found")
		}
	})
}
