package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/storage"
)

func main() {
	fmt.Println("Starting S3 Storage Test...")

	// 1. Load Config
	cfg, err := config.Load("")
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := runS3Test(context.Background(), cfg, os.Stdout); err != nil {
		fmt.Printf("❌ S3 test failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("🎉 All S3 tests passed!")
}

func runS3Test(ctx context.Context, cfg *config.Config, w io.Writer) error {
	// Force S3 type for this test if not set (or we can rely on env vars)
	// We will assume the user sets IMG_STORAGE_TYPE=s3 if needed, or we check it here.
	if cfg.Storage.Type != "s3" {
		fmt.Fprintf(w, "⚠️  Configured storage type is '%s', not 's3'. verify configuration.\n", cfg.Storage.Type)
		// We can force it for the test structure if we want to test specifically S3 logic
		// assuming S3 config is populated.
	}

	fmt.Fprintf(w, "Configuration:\nType: %s\nBucket: %s\nRegion: %s\nEndpoint: %s\n",
		cfg.Storage.Type,
		cfg.Storage.S3.Bucket,
		cfg.Storage.S3.Region,
		cfg.Storage.S3.Endpoint,
	)

	// 2. Initialize S3 Storage directly
	s3Store, err := storage.NewS3Storage(ctx, cfg.Storage.S3)
	if err != nil {
		return fmt.Errorf("failed to create S3 storage: %w", err)
	}

	// 3. Test Put with text file
	testKey := fmt.Sprintf("test-upload-%d.txt", time.Now().Unix())
	testContent := []byte("This is a test upload from images-filters S3 verification tool.")

	fmt.Fprintf(w, "Uploading %s...\n", testKey)
	err = s3Store.Put(ctx, testKey, testContent)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	fmt.Fprintf(w, "✅ Upload successful\n")

	// 4. Test Exists
	fmt.Fprintf(w, "Checking existence of %s...\n", testKey)
	exists, err := s3Store.Exists(ctx, testKey)
	if err != nil {
		return fmt.Errorf("check existence failed: %w", err)
	}
	if !exists {
		return fmt.Errorf("file should exist but Exists() returned false")
	}
	fmt.Fprintf(w, "✅ File exists\n")

	// 5. Test Get
	fmt.Fprintf(w, "Downloading %s...\n", testKey)
	content, err := s3Store.Get(ctx, testKey)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	if string(content) != string(testContent) {
		return fmt.Errorf("content mismatch. Expected: %s, Got: %s", string(testContent), string(content))
	}
	fmt.Fprintf(w, "✅ Download and content verification successful\n")

	fmt.Fprintf(w, "⚠️  SKIPPING DELETE. File '%s' should remain in the bucket for verification.\n", testKey)

	return nil
}
