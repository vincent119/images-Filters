package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantExit int
		wantOut  string
	}{
		{
			name:     "No args",
			args:     []string{"sign_url"},
			wantExit: 1,
			wantOut:  "Usage:",
		},
		{
			name:     "Missing Key and Config",
			// We point to non-existent config so it fails if key is also missing
			args:     []string{"sign_url", "-path", "test.jpg", "-config", "nonexistent.yaml"},
			wantExit: 1,
			wantOut:  "Failed to load config",
		},
		{
			name:     "Success with Key",
			args:     []string{"sign_url", "-path", "test.jpg", "-key", "secret"},
			wantExit: 0,
			wantOut:  "/test.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := new(bytes.Buffer)
			exitCode := run(tt.args, out)
			assert.Equal(t, tt.wantExit, exitCode)
			if tt.wantOut != "" {
				assert.Contains(t, out.String(), tt.wantOut)
			}
		})
	}
}
