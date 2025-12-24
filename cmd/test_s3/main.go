package main

import (
	"context"
	"fmt"
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

	// Force S3 type for this test if not set (or we can rely on env vars)
	// We will assume the user sets IMG_STORAGE_TYPE=s3 if needed, or we check it here.
	if cfg.Storage.Type != "s3" {
		fmt.Printf("⚠️  Configured storage type is '%s', not 's3'. verify configuration.\n", cfg.Storage.Type)
		// We can force it for the test structure if we want to test specifically S3 logic
		// assuming S3 config is populated.
	}

	fmt.Printf("Configuration:\nType: %s\nBucket: %s\nRegion: %s\nEndpoint: %s\n",
		cfg.Storage.Type,
		cfg.Storage.S3.Bucket,
		cfg.Storage.S3.Region,
		cfg.Storage.S3.Endpoint,
	)

	// 2. Initialize S3 Storage directly
	ctx := context.Background()
	s3Store, err := storage.NewS3Storage(ctx, cfg.Storage.S3)
	if err != nil {
		fmt.Printf("❌ Failed to create S3 storage: %v\n", err)
		os.Exit(1)
	}

	// 3. Test Put with text file
	testKey := fmt.Sprintf("test-upload-%d.txt", time.Now().Unix())
	testContent := []byte("This is a test upload from images-filters S3 verification tool.")

	fmt.Printf("Uploading %s...\n", testKey)
	err = s3Store.Put(ctx, testKey, testContent)
	if err != nil {
		fmt.Printf("❌ Upload failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Upload successful")

	// 4. Test Exists
	fmt.Printf("Checking existence of %s...\n", testKey)
	exists, err := s3Store.Exists(ctx, testKey)
	if err != nil {
		fmt.Printf("❌ Check existence failed: %v\n", err)
		os.Exit(1)
	}
	if !exists {
		fmt.Printf("❌ File should exist but Exists() returned false\n")
		os.Exit(1)
	}
	fmt.Println("✅ File exists")

	// 5. Test Get
	fmt.Printf("Downloading %s...\n", testKey)
	content, err := s3Store.Get(ctx, testKey)
	if err != nil {
		fmt.Printf("❌ Download failed: %v\n", err)
		os.Exit(1)
	}
	if string(content) != string(testContent) {
		fmt.Printf("❌ Content mismatch.\nExpected: %s\nGot: %s\n", string(testContent), string(content))
		os.Exit(1)
	}
	fmt.Println("✅ Download and content verification successful")

	// 6. Test Delete - SKIPPED for user verification
	// fmt.Printf("Deleting %s...\n", testKey)
	// err = s3Store.Delete(ctx, testKey)
	// if err != nil {
	// 	fmt.Printf("❌ Delete failed: %v\n", err)
	// 	os.Exit(1)
	// }

	// // Verify deletion
	// exists, err = s3Store.Exists(ctx, testKey)
	// if err != nil {
	// 	fmt.Printf("❌ Check existence after delete failed: %v\n", err)
	// 	os.Exit(1)
	// }
	// if exists {
	// 	fmt.Printf("❌ File should not exist after delete\n")
	// 	os.Exit(1)
	// }
	// fmt.Println("✅ Delete successful")

	fmt.Printf("⚠️  SKIPPING DELETE. File '%s' should remain in the bucket for verification.\n", testKey)
	fmt.Println("🎉 All S3 tests passed!")
}
