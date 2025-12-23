package types

import "errors"

// ErrNotFound records that the key was not found in storage
var ErrNotFound = errors.New("not found")
