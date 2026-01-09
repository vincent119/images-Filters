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
	"github.com/stretchr/testify/assert"
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

func setupTestStorage() (*S3Storage, *MockS3API) {
	mockAPI := NewMockS3API()
	storage := &S3Storage{
		client: mockAPI,
		bucket: "test-bucket",
	}
	return storage, mockAPI
}

func TestS3Storage_Put(t *testing.T) {
	storage, mockAPI := setupTestStorage()
	ctx := context.Background()
	testKey := "test/image.jpg"
	testData := []byte("test image data")

	err := storage.Put(ctx, testKey, testData)
	assert.NoError(t, err)

	// Verify data in mock
	mockAPI.mu.RLock()
	storedData, exists := mockAPI.objects[testKey]
	mockAPI.mu.RUnlock()

	assert.True(t, exists, "object should exist in mock")
	assert.Equal(t, testData, storedData, "stored data mismatch")
}

func TestS3Storage_Exists(t *testing.T) {
	storage, _ := setupTestStorage()
	ctx := context.Background()
	testKey := "test/image.jpg"
	testData := []byte("test image data")

	// Pre-populate
	err := storage.Put(ctx, testKey, testData)
	assert.NoError(t, err)

	 exists, err := storage.Exists(ctx, testKey)
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = storage.Exists(ctx, "nonexistent.jpg")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestS3Storage_Get(t *testing.T) {
	storage, _ := setupTestStorage()
	ctx := context.Background()
	testKey := "test/image.jpg"
	testData := []byte("test image data")

	err := storage.Put(ctx, testKey, testData)
	assert.NoError(t, err)

	data, err := storage.Get(ctx, testKey)
	assert.NoError(t, err)
	assert.Equal(t, testData, data)

	_, err = storage.Get(ctx, "nonexistent.jpg")
	assert.Error(t, err)
}

func TestS3Storage_Delete(t *testing.T) {
	storage, _ := setupTestStorage()
	ctx := context.Background()
	testKey := "test/image.jpg"
	testData := []byte("test image data")

	err := storage.Put(ctx, testKey, testData)
	assert.NoError(t, err)

	err = storage.Delete(ctx, testKey)
	assert.NoError(t, err)

	exists, _ := storage.Exists(ctx, testKey)
	assert.False(t, exists, "object should be deleted")
}

func TestS3Storage_Stream(t *testing.T) {
	storage, mockAPI := setupTestStorage()
	ctx := context.Background()
	streamKey := "stream/test.jpg"
	streamData := []byte("stream data")

	// PutStream
	err := storage.PutStream(ctx, streamKey, bytes.NewReader(streamData))
	assert.NoError(t, err)

	// Verify PutStream result via normal Get logic on mock
	mockAPI.mu.RLock()
	storedData, exists := mockAPI.objects[streamKey]
	mockAPI.mu.RUnlock()
	assert.True(t, exists, "PutStream object should exist in mock")
	assert.Equal(t, streamData, storedData, "PutStream stored data mismatch")

	// GetStream
	rc, err := storage.GetStream(ctx, streamKey)
	assert.NoError(t, err)
	defer rc.Close()

	readData, err := io.ReadAll(rc)
	assert.NoError(t, err)
	assert.Equal(t, streamData, readData)

	// GetStream Not Found
	_, err = storage.GetStream(ctx, "nonexistent-stream")
	assert.Error(t, err)
}
