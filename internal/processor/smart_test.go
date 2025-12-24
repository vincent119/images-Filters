package processor

import (
	"image"
	"image/color"
	"testing"
)

// createTestImageWithFeature creates a generated image with a specific feature
// (a red square) in a specific location to test smart crop.
// The background is black.
func createTestImageWithFeature(w, h int, featureX, featureY, featureSize int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill background with black
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.Black)
		}
	}

	// Draw a red square as the "feature" (high saturation/contrast)
	red := color.RGBA{255, 0, 0, 255}
	for y := featureY; y < featureY+featureSize && y < h; y++ {
		for x := featureX; x < featureX+featureSize && x < w; x++ {
			img.Set(x, y, red)
		}
	}

	return img
}

func TestSmartCrop(t *testing.T) {
	// Create a 1000x500 image with a feature on the right side
	// Feature is at x=800, y=100, size=100
	width := 1000
	height := 500
	img := createTestImageWithFeature(width, height, 800, 100, 100)

	// Initialize Processor
	proc := NewProcessor(80, 2000, 2000)

	// Request smart crop to 200x200
	// We expect the crop to include the red square area (around 800, 100)
	// Standard center crop would take 400-600 width range

	cropped, err := proc.smartCrop(img, 200, 200)
	if err != nil {
		t.Fatalf("Smart crop failed: %v", err)
	}

	// Check aspect ratio
	// Expected 1:1
	bounds := cropped.Bounds()
	ratio := float64(bounds.Dx()) / float64(bounds.Dy())
	if ratio < 0.99 || ratio > 1.01 {
		t.Errorf("Expected aspect ratio 1:1, got %.2f (%dx%d)", ratio, bounds.Dx(), bounds.Dy())
	}

	// Verify that the crop was actually performed (not just returning original)
	if bounds.Dx() == width && bounds.Dy() == height {
		t.Error("Smart crop return original image size, expected a crop")
	}

	// Verify that the crop actually contains the feature (Red pixel)
	// We check the center of the cropped image
	// The feature is all red. If smart crop found it, the cropped image should contain red pixels.

	// A strictly better check is to see if the CropRect (if we could access it) is correct.
	// Since we return an image, we can just scan it.

	hasRed := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := cropped.At(x, y)
			r, _, _, _ := c.RGBA()
			if r > 0 {
				hasRed = true
				break
			}
		}
		if hasRed {
			break
		}
	}

	if !hasRed {
		t.Error("Smart crop did not capture the significant feature (red square). Result might be just black.")
	}
}
