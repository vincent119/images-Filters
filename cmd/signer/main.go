// Package main provides URL Signer CLI tool
// for generating and verifying HMAC-signed image processing URLs
package main

import (
	"flag"
	"fmt"
	"io"
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
	os.Exit(run(os.Args, os.Stdout))
}

func run(args []string, out io.Writer) int {
	if len(args) < 2 {
		fmt.Fprint(out, usage)
		return 1
	}

	command := args[1]

	switch command {
	case "sign":
		return signCmd(args[2:], out)
	case "verify":
		return verifyCmd(args[2:], out)
	case "help", "-h", "--help":
		fmt.Fprint(out, usage)
		return 0
	default:
		fmt.Fprintf(out, "Unknown command: %s\n\n", command)
		fmt.Fprint(out, usage)
		return 1
	}
}

func signCmd(args []string, out io.Writer) int {
	signFlags := flag.NewFlagSet("sign", flag.ContinueOnError)
	signFlags.SetOutput(out)
	keyPtr := signFlags.String("key", "", "Security key (or set IMG_SECURITY_KEY env)")
	pathPtr := signFlags.String("path", "", "URL path to sign (e.g., 300x200/test.jpg)")
	baseURLPtr := signFlags.String("base", "", "Base URL (optional, e.g., http://localhost:8080)")
	quietPtr := signFlags.Bool("quiet", false, "Output only the signed path")

	if err := signFlags.Parse(args); err != nil {
		return 1
	}

	// Get key
	key := *keyPtr
	if key == "" {
		key = os.Getenv("IMG_SECURITY_KEY")
	}
	if key == "" {
		fmt.Fprintln(out, "Error: Security key required (-key or IMG_SECURITY_KEY)")
		return 1
	}

	// Get path
	path := *pathPtr
	if path == "" {
		fmt.Fprintln(out, "Error: URL path required (-path)")
		return 1
	}

	// Generate signature
	signer := security.NewSigner(key)
	signedPath := signer.SignURL(path)

	// Output results
	if *quietPtr {
		fmt.Fprint(out, signedPath)
		return 0
	}

	fmt.Fprintln(out, "Original path:", path)
	fmt.Fprintln(out, "Signed path:  ", signedPath)

	if *baseURLPtr != "" {
		fmt.Fprintln(out, "Full URL:     ", *baseURLPtr+signedPath)
	}
	return 0
}

func verifyCmd(args []string, out io.Writer) int {
	verifyFlags := flag.NewFlagSet("verify", flag.ContinueOnError)
	verifyFlags.SetOutput(out)
	keyPtr := verifyFlags.String("key", "", "Security key (or set IMG_SECURITY_KEY env)")
	urlPtr := verifyFlags.String("url", "", "Full signed URL path to verify")

	if err := verifyFlags.Parse(args); err != nil {
		return 1
	}

	// Get key
	key := *keyPtr
	if key == "" {
		key = os.Getenv("IMG_SECURITY_KEY")
	}
	if key == "" {
		fmt.Fprintln(out, "Error: Security key required (-key or IMG_SECURITY_KEY)")
		return 1
	}

	// Get URL
	url := *urlPtr
	if url == "" {
		fmt.Fprintln(out, "Error: URL required (-url)")
		return 1
	}

	// Extract signature and path
	signature, path, ok := security.ExtractSignatureAndPath(url)
	if !ok {
		fmt.Fprintln(out, "❌ Invalid URL format")
		return 1
	}

	// Verify signature
	signer := security.NewSigner(key)
	if signer.Verify(signature, path) {
		fmt.Fprintln(out, "✅ Signature valid")
		fmt.Fprintln(out, "   Path:", path)
		return 0
	} else {
		fmt.Fprintln(out, "❌ Signature invalid")
		return 1
	}
}
