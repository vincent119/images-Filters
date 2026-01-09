package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vincent119/images-filters/internal/storage/types"
)

func TestNoStorage(t *testing.T) {
	s := NewNoStorage()
	ctx := context.Background()
	key := "test.jpg"
	data := []byte("fake")

	// Get
	val, err := s.Get(ctx, key)
	assert.Equal(t, types.ErrNotFound, err)
	assert.Nil(t, val)

	// Put
	err = s.Put(ctx, key, data)
	assert.NoError(t, err)

	// Exists
	exists, err := s.Exists(ctx, key)
	assert.NoError(t, err)
	assert.False(t, exists)

	// Delete
	err = s.Delete(ctx, key)
	assert.NoError(t, err)

	// GetStream
	stream, err := s.GetStream(ctx, key)
	assert.Equal(t, types.ErrNotFound, err)
	assert.Nil(t, stream)

	// PutStream
	reader := strings.NewReader("fake stream")
	err = s.PutStream(ctx, key, io.NopCloser(reader))
	assert.NoError(t, err)

	// PutStream large read ensure discard
	largeReader := bytes.NewReader(make([]byte, 1024))
	done := make(chan bool)
	go func() {
		s.PutStream(ctx, key, io.NopCloser(largeReader))
		done <- true
	}()
	select {
	case <-done:
		// success
	case <-time.After(time.Second):
		t.Fatal("PutStream blocked too long")
	}
}
