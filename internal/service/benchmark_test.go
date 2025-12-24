package service

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"testing"

	"github.com/vincent119/images-filters/internal/processor"
)

// generateBenchmarkImage creates a JPEG image in memory for benchmarking
func generateBenchmarkImage(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// Fill with some noise/gradient to make it realistic for compression
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x + y) % 255),
				G: uint8((x * y) % 255),
				B: uint8(x % 255),
				A: 255,
			})
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75}); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

var benchmarkData []byte

func init() {
	// Generate a 1000x1000 image once
	benchmarkData = generateBenchmarkImage(1000, 1000)
}

// Benchmark_Buffering simulates the old behavior:
// 1. Read the entire stream into a byte slice (Load)
// 2. Create a reader from that slice
// 3. Process
func Benchmark_Buffering(b *testing.B) {
	proc := processor.NewProcessor(80, 2000, 2000)
	// ctx := context.Background() // Not used in Process anymore
	opts := processor.ProcessOptions{
		Width:  500,
		Height: 500,
		Format: "jpeg",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate the input stream
		inputStream := bytes.NewReader(benchmarkData)

		// Step 1: Buffering - Read entire content into memory
		// This simulates the old loader.Load() returning []byte
		data, err := io.ReadAll(inputStream)
		if err != nil {
			b.Fatal(err)
		}

		// Step 2: Process (requires wrapping data back into a reader for current API)
		_, err = proc.Process(bytes.NewReader(data), opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark_Streaming simulates the new behavior:
// 1. Pass the stream directly to Process
func Benchmark_Streaming(b *testing.B) {
	proc := processor.NewProcessor(80, 2000, 2000)
	// ctx := context.Background() // Not used in Process anymore
	opts := processor.ProcessOptions{
		Width:  500,
		Height: 500,
		Format: "jpeg",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate the input stream
		inputStream := bytes.NewReader(benchmarkData)

		// Step 1: Streaming - internal pipeline reads directly from input stream
		_, err := proc.Process(inputStream, opts)
		if err != nil {
			b.Fatal(err)
		}
	}
}
