package main

import (
	"bytes"
	"os"
	"strings"
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
			args:     []string{"signer"},
			wantExit: 1,
			wantOut:  "Usage:",
		},
		{
			name:     "Help",
			args:     []string{"signer", "help"},
			wantExit: 0,
			wantOut:  "Usage:",
		},
		{
			name:     "Unknown command",
			args:     []string{"signer", "unknown"},
			wantExit: 1,
			wantOut:  "Unknown command: unknown",
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

func TestSignCmd(t *testing.T) {
	// Set env for test but reset later
	os.Setenv("IMG_SECURITY_KEY", "")
	defer os.Setenv("IMG_SECURITY_KEY", "")

	tests := []struct {
		name     string
		args     []string
		wantExit int
		wantOut  string
	}{
		{
			name:     "Missing Key",
			args:     []string{"sign", "-path", "test.jpg"},
			wantExit: 1,
			wantOut:  "Security key required",
		},
		{
			name:     "Missing Path",
			args:     []string{"sign", "-key", "secret"},
			wantExit: 1,
			wantOut:  "URL path required",
		},
		{
			name:     "Success",
			args:     []string{"sign", "-key", "secret", "-path", "test.jpg", "-quiet"},
			wantExit: 0,
			wantOut:  "", // Quiet mode output checked later, but here we expect no error text
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := new(bytes.Buffer)
			// signCmd expects args starting AFTER "sign"
			// But run() handles splitting.
			// Let's call run() with full args to be simple
			fullArgs := append([]string{"signer"}, tt.args...)

			exitCode := run(fullArgs, out)
			assert.Equal(t, tt.wantExit, exitCode)
			if tt.wantOut != "" {
				assert.Contains(t, out.String(), tt.wantOut)
			}
		})
	}

	// Test Output Content
	t.Run("Output Content", func(t *testing.T) {
		out := new(bytes.Buffer)
		exitCode := run([]string{"signer", "sign", "-key", "test-secret", "-path", "test.jpg", "-quiet"}, out)
		assert.Equal(t, 0, exitCode)
		assert.Contains(t, out.String(), "/test.jpg") // Should contain signature
	})
}

func TestVerifyCmd(t *testing.T) {
	os.Setenv("IMG_SECURITY_KEY", "")

	// Generate a valid signed URL first using run()
	out := new(bytes.Buffer)
	run([]string{"signer", "sign", "-key", "secret", "-path", "test.jpg", "-quiet"}, out)
	validURL := out.String()
	out.Reset()

	tests := []struct {
		name     string
		args     []string
		wantExit int
		wantOut  string
	}{
		{
			name:     "Missing Key",
			args:     []string{"verify", "-url", validURL},
			wantExit: 1,
			wantOut:  "Security key required",
		},
		{
			name:     "Missing URL",
			args:     []string{"verify", "-key", "secret"},
			wantExit: 1,
			wantOut:  "URL required",
		},
		{
			name:     "Invalid Format",
			args:     []string{"verify", "-key", "secret", "-url", "bad-url"},
			wantExit: 1,
			wantOut:  "Invalid URL format",
		},
		{
			name: "Invalid Signature",
			// Signature must be 44 chars long to pass format check
			args:     []string{"verify", "-key", "secret", "-url", "/" + strings.Repeat("a", 44) + "/test.jpg"},
			wantExit: 1,
			wantOut:  "Signature invalid",
		},
		{
			name:     "Success",
			args:     []string{"verify", "-key", "secret", "-url", validURL},
			wantExit: 0,
			wantOut:  "Signature valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out.Reset()
			fullArgs := append([]string{"signer"}, tt.args...)
			exitCode := run(fullArgs, out)
			assert.Equal(t, tt.wantExit, exitCode)
			if tt.wantOut != "" {
				assert.Contains(t, out.String(), tt.wantOut)
			}
		})
	}
}
