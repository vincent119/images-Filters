package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateImage(t *testing.T) {
	tempDir := t.TempDir()
	filename := filepath.Join(tempDir, "test_gen.jpg")

	err := generateImage(filename)
	if err != nil {
		t.Fatalf("generateImage failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("File %s not created", filename)
	}
}
