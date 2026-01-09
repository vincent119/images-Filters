package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNoOpCache(t *testing.T) {
	c := NewNoOpCache()
	ctx := context.Background()

	// Exists should always return false
	exists, err := c.Exists(ctx, "any")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Get should always return ErrCacheMiss
	val, err := c.Get(ctx, "any")
	assert.Equal(t, ErrCacheMiss, err)
	assert.Nil(t, val)

	// Set should not error
	err = c.Set(ctx, "key", []byte("val"), time.Second)
	assert.NoError(t, err)

	// Delete should not error
	err = c.Delete(ctx, "key")
	assert.NoError(t, err)
}
