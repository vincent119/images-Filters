// Package main provides URL Signer CLI tool
// for generating and verifying HMAC-signed image processing URLs
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vincent119/images-filters/internal/security"
)

const usage = `URL Signer - HMAC Signature Tool

Usage:
  signer [command] [options]

Commands:
  sign      Generate signed URL
  verify    Verify signed URL

Examples:
  # Generate signed URL
  signer sign -key "your-secret-key" -path "300x200/test.jpg"

  # Verify signature
  signer verify -key "your-secret-key" -url "/abc123.../300x200/test.jpg"

Environment Variables:
  IMG_SECURITY_KEY    Security key (alternative to -key flag)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "sign":
		signCmd()
	case "verify":
		verifyCmd()
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		fmt.Print(usage)
		os.Exit(1)
	}
}

func signCmd() {
	signFlags := flag.NewFlagSet("sign", flag.ExitOnError)
	keyPtr := signFlags.String("key", "", "Security key (or set IMG_SECURITY_KEY env)")
	pathPtr := signFlags.String("path", "", "URL path to sign (e.g., 300x200/test.jpg)")
	baseURLPtr := signFlags.String("base", "", "Base URL (optional, e.g., http://localhost:8080)")

	if err := signFlags.Parse(os.Args[2:]); err != nil {
		fmt.Println("Error parsing arguments:", err)
		os.Exit(1)
	}

	// Get key
	key := *keyPtr
	if key == "" {
		key = os.Getenv("IMG_SECURITY_KEY")
	}
	if key == "" {
		fmt.Println("Error: Security key required (-key or IMG_SECURITY_KEY)")
		os.Exit(1)
	}

	// Get path
	path := *pathPtr
	if path == "" {
		fmt.Println("Error: URL path required (-path)")
		os.Exit(1)
	}

	// Generate signature
	signer := security.NewSigner(key)
	signedPath := signer.SignURL(path)

	// Output results
	fmt.Println("Original path:", path)
	fmt.Println("Signed path:  ", signedPath)

	if *baseURLPtr != "" {
		fmt.Println("Full URL:     ", *baseURLPtr+signedPath)
	}
}

func verifyCmd() {
	verifyFlags := flag.NewFlagSet("verify", flag.ExitOnError)
	keyPtr := verifyFlags.String("key", "", "Security key (or set IMG_SECURITY_KEY env)")
	urlPtr := verifyFlags.String("url", "", "Full signed URL path to verify")

	if err := verifyFlags.Parse(os.Args[2:]); err != nil {
		fmt.Println("Error parsing arguments:", err)
		os.Exit(1)
	}

	// Get key
	key := *keyPtr
	if key == "" {
		key = os.Getenv("IMG_SECURITY_KEY")
	}
	if key == "" {
		fmt.Println("Error: Security key required (-key or IMG_SECURITY_KEY)")
		os.Exit(1)
	}

	// Get URL
	url := *urlPtr
	if url == "" {
		fmt.Println("Error: URL required (-url)")
		os.Exit(1)
	}

	// Extract signature and path
	signature, path, ok := security.ExtractSignatureAndPath(url)
	if !ok {
		fmt.Println("❌ Invalid URL format")
		os.Exit(1)
	}

	// Verify signature
	signer := security.NewSigner(key)
	if signer.Verify(signature, path) {
		fmt.Println("✅ Signature valid")
		fmt.Println("   Path:", path)
	} else {
		fmt.Println("❌ Signature invalid")
		os.Exit(1)
	}
}

