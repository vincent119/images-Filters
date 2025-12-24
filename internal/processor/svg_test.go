package processor

import (
	"strings"
	"testing"
)

func TestProcessor_Process_SVG(t *testing.T) {
	// Simple SVG string
	svgContent := `<?xml version="1.0" encoding="UTF-8"?>
<svg width="100" height="100" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
    <rect x="10" y="10" width="80" height="80" fill="red"/>
</svg>`

	p := NewProcessor(80, 1000, 1000)

	// Test case 1: Default size
	img, err := p.Process(strings.NewReader(svgContent), ProcessOptions{})
	if err != nil {
		t.Fatalf("Failed to process SVG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Expected default size 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Test case 2: Resized
	imgResized, err := p.Process(strings.NewReader(svgContent), ProcessOptions{Width: 200, Height: 200})
	if err != nil {
		t.Fatalf("Failed to process SVG with resize: %v", err)
	}

	boundsResized := imgResized.Bounds()
	if boundsResized.Dx() != 200 || boundsResized.Dy() != 200 {
		t.Errorf("Expected resized 200x200, got %dx%d", boundsResized.Dx(), boundsResized.Dy())
	}
}
