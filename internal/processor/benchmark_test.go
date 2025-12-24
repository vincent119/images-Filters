package processor

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"testing"
)

// generateTestImage generates a random noise image of given dimensions
func generateTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{100, 150, 200, 255}}, image.Point{}, draw.Src)
	return img
}

// generateTestImageBytes generates image bytes for testing
func generateTestImageBytes(width, height int) []byte {
	img := generateTestImage(width, height)
	var buf bytes.Buffer
	p := NewProcessor(80, 2000, 2000)
	p.Encode(img, "jpeg", 80)
	buf.WriteTo(&buf)
	// Wait, Encode writes to buffer? No, Encode returns []byte.
	// Let's check Processor.Encode implementation.
	// It returns []byte.
	data, _ := p.Encode(img, "jpeg", 80)
	return data
}

func BenchmarkProcessor_Resize(b *testing.B) {
	// Setup
	p := NewProcessor(80, 2000, 2000)
	inputData := generateTestImageBytes(1920, 1080)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(inputData)
		opts := ProcessOptions{
			Width:  800,
			Height: 600,
			Format: "jpeg",
		}
		// Assuming Process takes io.Reader now
		_, err := p.Process(r, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProcessor_Crop(b *testing.B) {
	p := NewProcessor(80, 2000, 2000)
	inputData := generateTestImageBytes(1920, 1080)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(inputData)
		opts := ProcessOptions{
			CropLeft:   100,
			CropTop:    100,
			CropRight:  1700,
			CropBottom: 900,
			Format:     "jpeg",
		}
		_, err := p.Process(r, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProcessor_Encode_JPEG(b *testing.B) {
	p := NewProcessor(80, 2000, 2000)
	img := generateTestImage(1920, 1080)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Encode(img, "jpeg", 80)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProcessor_Encode_PNG(b *testing.B) {
	p := NewProcessor(80, 2000, 2000)
	img := generateTestImage(1920, 1080)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Encode(img, "png", 0)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProcessor_StreamProcess(b *testing.B) {
	// Simulate streaming by using a slow reader or just normal reader
	p := NewProcessor(80, 2000, 2000)
	inputData := generateTestImageBytes(4096, 2160) // 4K Image

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(inputData)
		opts := ProcessOptions{
			Width:  1920,
			Height: 1080,
			Format: "jpeg",
		}
		_, err := p.Process(r, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}
