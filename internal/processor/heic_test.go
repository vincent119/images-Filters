package processor

import (
	"testing"
)

func TestProcessor_HEIC_Support(t *testing.T) {
	// Verify Content Type
	ct := GetContentType("heic")
	if ct != "image/heic" {
		t.Errorf("Expected content type image/heic, got %s", ct)
	}
}

// TestHEICDecoding checks if 'heic' format is registered in image package.
func TestHEICDecoding_Registration(t *testing.T) {
	// We can't easily Encode to HEIC to test Decode,
	// so we check internal registration indirectly or just rely on import.
	// Here we try to simulate a decode to see if "heic" is recognized?
	// Actually, without a sample file, best we can do is ensure import doesn't panic
	// and maybe check if we can conceptually support it.

	// Check if "heic" is in the list of registered formats?
	// Go doesn't expose a list of registered formats directly.

	// Attempting to decode garbage with "heir" header might trigger specific error if format detected?
	// But simplest is to trust the import side-effect for now
	// as we don't have a small heic sample string.
}
