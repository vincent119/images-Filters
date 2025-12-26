package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/security"
)

func main() {
	key := flag.String("key", "", "Security Key (override config)")
	path := flag.String("path", "", "Path to sign (e.g. '500x100/uploads/...')")
	configFile := flag.String("config", "config/config.yaml", "Path to config file")
	flag.Parse()

	if *path == "" {
		fmt.Println("Usage: go run cmd/sign_url/main.go -path \"500x100/uploads/2025/...\"")
		flag.PrintDefaults()
		os.Exit(1)
	}

	secretKey := *key
	if secretKey == "" {
		// Try to load from config
		cfg, err := config.Load(*configFile)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
		secretKey = cfg.Security.SecurityKey
	}

	signer := security.NewSigner(secretKey)
	signedURL := signer.SignURL(*path)

	fmt.Printf("\nOriginal Path: %s\n", *path)
	fmt.Printf("Security Key:  %s\n", secretKey)
	fmt.Printf("Signed URL:    %s\n", signedURL)
	fmt.Printf("Full Example:  http://localhost:8080%s\n\n", signedURL)
}
