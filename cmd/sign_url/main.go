package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/security"
)

func main() {
	os.Exit(run(os.Args, os.Stdout))
}

func run(args []string, out io.Writer) int {
	flags := flag.NewFlagSet("sign_url", flag.ContinueOnError)
	flags.SetOutput(out)
	key := flags.String("key", "", "Security Key (override config)")
	path := flags.String("path", "", "Path to sign (e.g. '500x100/uploads/...')")
	configFile := flags.String("config", "config/config.yaml", "Path to config file")

	if err := flags.Parse(args[1:]); err != nil {
		return 1
	}

	if *path == "" {
		fmt.Fprintln(out, "Usage: go run cmd/sign_url/main.go -path \"500x100/uploads/2025/...\"")
		flags.SetOutput(out)
		flags.PrintDefaults()
		return 1
	}

	secretKey := *key
	if secretKey == "" {
		// Try to load from config
		cfg, err := config.Load(*configFile)
		if err != nil {
			// If config fails and key not provided, fail
			// log.Fatalf exits program, we want to return 1.
			// However in test environment we might not have config file.
			// Let's print error.
			fmt.Fprintf(out, "Failed to load config: %v\n", err)
			return 1
		}
		secretKey = cfg.Security.SecurityKey
	}

	signer := security.NewSigner(secretKey)
	signedURL := signer.SignURL(*path)

	fmt.Fprintf(out, "\nOriginal Path: %s\n", *path)
	fmt.Fprintf(out, "Security Key:  %s\n", secretKey)
	fmt.Fprintf(out, "Signed URL:    %s\n", signedURL)
	fmt.Fprintf(out, "Full Example:  http://localhost:8080%s\n\n", signedURL)
	return 0
}
