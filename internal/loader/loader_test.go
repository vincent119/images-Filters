package loader

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLoader for testing factory delegation
type MockLoader struct {
	mock.Mock
}

func (m *MockLoader) Load(ctx context.Context, source string) ([]byte, error) {
	args := m.Called(ctx, source)
	var data []byte
	if args.Get(0) != nil {
		data = args.Get(0).([]byte)
	}
	return data, args.Error(1)
}

func (m *MockLoader) CanLoad(source string) bool {
	args := m.Called(source)
	return args.Bool(0)
}

func (m *MockLoader) LoadStream(ctx context.Context, source string) (io.ReadCloser, error) {
	args := m.Called(ctx, source)
	var stream io.ReadCloser
	if args.Get(0) != nil {
		stream = args.Get(0).(io.ReadCloser)
	}
	return stream, args.Error(1)
}

func TestNewLoaderFactory(t *testing.T) {
	l1 := new(MockLoader)
	l2 := new(MockLoader)
	f := NewLoaderFactory(l1, l2)
	assert.NotNil(t, f)
	assert.Len(t, f.loaders, 2)
}

func TestLoaderFactory_GetLoader(t *testing.T) {
	l1 := new(MockLoader)
	l2 := new(MockLoader)
	f := NewLoaderFactory(l1, l2)

	// Scenario 1: First loader matches
	l1.On("CanLoad", "source1").Return(true)
	loader, err := f.GetLoader("source1")
	assert.NoError(t, err)
	assert.Equal(t, l1, loader)

	// Scenario 2: Second loader matches
	l1.On("CanLoad", "source2").Return(false)
	l2.On("CanLoad", "source2").Return(true)
	loader, err = f.GetLoader("source2")
	assert.NoError(t, err)
	assert.Equal(t, l2, loader)

	// Scenario 3: No loader matches
	l1.On("CanLoad", "unknown").Return(false)
	l2.On("CanLoad", "unknown").Return(false)
	loader, err = f.GetLoader("unknown")
	assert.Error(t, err)
	assert.Nil(t, loader)
	assert.Contains(t, err.Error(), "no suitable loader found")
}

func TestLoaderFactory_Load(t *testing.T) {
	l1 := new(MockLoader)
	f := NewLoaderFactory(l1)
	ctx := context.Background()

	// Success
	l1.On("CanLoad", "valid").Return(true)
	l1.On("Load", ctx, "valid").Return([]byte("data"), nil)
	data, err := f.Load(ctx, "valid")
	assert.NoError(t, err)
	assert.Equal(t, []byte("data"), data)

	// Loader Error
	l1.On("CanLoad", "error_source").Return(true)
	l1.On("Load", ctx, "error_source").Return(([]byte)(nil), errors.New("load error"))
	data, err = f.Load(ctx, "error_source")
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Equal(t, "load error", err.Error())

	// No Loader Found
	l1.On("CanLoad", "unknown").Return(false)
	data, err = f.Load(ctx, "unknown")
	assert.Error(t, err)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "no suitable loader found")
}

func TestLoaderFactory_LoadStream(t *testing.T) {
	l1 := new(MockLoader)
	f := NewLoaderFactory(l1)
	ctx := context.Background()

	// Success
	mockStream := io.NopCloser(strings.NewReader("stream data"))
	l1.On("CanLoad", "valid_stream").Return(true)
	l1.On("LoadStream", ctx, "valid_stream").Return(mockStream, nil)
	stream, err := f.LoadStream(ctx, "valid_stream")
	assert.NoError(t, err)
	data, _ := io.ReadAll(stream)
	assert.Equal(t, "stream data", string(data))

	// Stream Error
	l1.On("CanLoad", "error_stream").Return(true)
	l1.On("LoadStream", ctx, "error_stream").Return((io.ReadCloser)(nil), errors.New("stream error"))
	stream, err = f.LoadStream(ctx, "error_stream")
	assert.Error(t, err)
	assert.Nil(t, stream)

	// No Loader
	l1.On("CanLoad", "unknown").Return(false)
	stream, err = f.LoadStream(ctx, "unknown")
	assert.Error(t, err)
	assert.Nil(t, stream)
}

func TestReadAll(t *testing.T) {
	// Test unlimited read
	r := strings.NewReader("hello world")
	data, err := readAll(r, 0)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(data))

	// Test limited read (limit > content)
	r = strings.NewReader("short")
	data, err = readAll(r, 100)
	assert.NoError(t, err)
	assert.Equal(t, "short", string(data))

	// Test limited read (limit < content)
	r = strings.NewReader("longer than limit")
	data, err = readAll(r, 4)
	assert.NoError(t, err)
	assert.Equal(t, "long", string(data))

	// Test read error (hard to simulate with strings.Reader, use iotest if needed)
	// But readAll mainly delegates to io.ReadAll or io.LimitReader.
	// We trust stdlib, but ensuring args passed correctly.

	// Test error reader
	errReader := iotest.ErrReader(errors.New("read error"))
	data, err = readAll(errReader, 0)
	assert.Error(t, err)
	assert.Equal(t, "read error", err.Error())
}
