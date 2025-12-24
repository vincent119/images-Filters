package processor

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestProcessor_Encode_AVIF(t *testing.T) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{255, 0, 0, 255}}, image.Point{}, draw.Src)

	p := NewProcessor(80, 1000, 1000)

	// Encode to AVIF
	data, err := p.Encode(img, "avif", 50)
	if err != nil {
		t.Fatalf("Failed to encode AVIF: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Encoded AVIF data is empty")
	}

	// Verify Content Type
	ct := GetContentType("avif")
	if ct != "image/avif" {
		t.Errorf("Expected content type image/avif, got %s", ct)
	}

	// Decode back to verify it's a valid image
	// Note: gen2brain/avif registers format in init(), so image.Decode shoud work if we import it.
	// We imported it in processor.go, so it should be registered.
	decodedImg, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to decode AVIF: %v", err)
	}

	if format != "avif" {
		t.Errorf("Expected format avif, got %s", format)
	}

	bounds := decodedImg.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Expected dimensions 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}
