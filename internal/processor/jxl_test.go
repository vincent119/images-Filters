package processor

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestProcessor_Encode_JXL(t *testing.T) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 255, 0, 255}}, image.Point{}, draw.Src)

	p := NewProcessor(80, 1000, 1000)

	// Encode to JXL
	data, err := p.Encode(img, "jxl", 50)
	if err != nil {
		t.Fatalf("Failed to encode JXL: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Encoded JXL data is empty")
	}

	// Verify Content Type
	ct := GetContentType("jxl")
	if ct != "image/jxl" {
		t.Errorf("Expected content type image/jxl, got %s", ct)
	}

	// Decode back to verify it's a valid image
	// gen2brain/jpegxl registers format in init()
	decodedImg, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to decode JXL: %v", err)
	}

	// format string returned by gen2brain/jpegxl might be "jxl" or "jpegxl"
	if format != "jxl" && format != "jpegxl" {
		t.Errorf("Expected format jxl or jpegxl, got %s", format)
	}

	bounds := decodedImg.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("Expected dimensions 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}
