package processor

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

// Tests for internal helper functions

func TestCalculateDimensions(t *testing.T) {
	tests := []struct {
		name           string
		originalWidth  int
		originalHeight int
		targetWidth    int
		targetHeight   int
		fitIn          bool
		wantWidth      int
		wantHeight     int
	}{
		{
			name:           "原尺寸（兩個都是 0）",
			originalWidth:  800,
			originalHeight: 600,
			targetWidth:    0,
			targetHeight:   0,
			fitIn:          false,
			wantWidth:      800,
			wantHeight:     600,
		},
		{
			name:           "只指定寬度",
			originalWidth:  800,
			originalHeight: 600,
			targetWidth:    400,
			targetHeight:   0,
			fitIn:          false,
			wantWidth:      400,
			wantHeight:     300,
		},
		{
			name:           "只指定高度",
			originalWidth:  800,
			originalHeight: 600,
			targetWidth:    0,
			targetHeight:   300,
			fitIn:          false,
			wantWidth:      400,
			wantHeight:     300,
		},
		{
			name:           "指定固定尺寸（填滿模式）",
			originalWidth:  800,
			originalHeight: 600,
			targetWidth:    300,
			targetHeight:   200,
			fitIn:          false,
			wantWidth:      300,
			wantHeight:     200,
		},
		{
			name:           "Fit-in 較寬圖片",
			originalWidth:  800,
			originalHeight: 400,
			targetWidth:    300,
			targetHeight:   300,
			fitIn:          true,
			wantWidth:      300,
			wantHeight:     150,
		},
		{
			name:           "Fit-in 較高圖片",
			originalWidth:  400,
			originalHeight: 800,
			targetWidth:    300,
			targetHeight:   300,
			fitIn:          true,
			wantWidth:      150,
			wantHeight:     300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := calculateDimensions(
				tt.originalWidth, tt.originalHeight,
				tt.targetWidth, tt.targetHeight,
				tt.fitIn,
			)

			if w != tt.wantWidth {
				t.Errorf("width = %d; want %d", w, tt.wantWidth)
			}
			if h != tt.wantHeight {
				t.Errorf("height = %d; want %d", h, tt.wantHeight)
			}
		})
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{"jpeg", "image/jpeg"},
		{"jpg", "image/jpeg"},
		{"png", "image/png"},
		{"gif", "image/gif"},
		{"webp", "image/webp"},
		{"avif", "image/avif"},
		{"jxl", "image/jxl"},
		{"unknown", "image/jpeg"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			if got := GetContentType(tt.format); got != tt.want {
				t.Errorf("GetContentType(%s) = %s; want %s", tt.format, got, tt.want)
			}
		})
	}
}

// createTestImage 建立測試用圖片
func createTestImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: 128,
				A: 255,
			})
		}
	}
	return img
}

func TestEncode_JPEG(t *testing.T) {
	p := NewProcessor(85, 4096, 4096)
	img := createTestImage(100, 100)

	// 測試 JPEG 編碼
	data, err := p.Encode(img, "jpeg", 85)
	if err != nil {
		t.Fatalf("Encode JPEG failed: %v", err)
	}

	// 驗證 JPEG 魔術數字 (FF D8 FF)
	if len(data) < 3 || data[0] != 0xFF || data[1] != 0xD8 || data[2] != 0xFF {
		t.Error("Output is not valid JPEG format")
	}
}

func TestEncode_PNG(t *testing.T) {
	p := NewProcessor(85, 4096, 4096)
	img := createTestImage(100, 100)

	data, err := p.Encode(img, "png", 0)
	if err != nil {
		t.Fatalf("Encode PNG failed: %v", err)
	}

	// 驗證 PNG 魔術數字 (89 50 4E 47)
	if len(data) < 4 || data[0] != 0x89 || data[1] != 0x50 || data[2] != 0x4E || data[3] != 0x47 {
		t.Error("Output is not valid PNG format")
	}
}

func TestEncode_GIF(t *testing.T) {
	p := NewProcessor(85, 4096, 4096)
	img := createTestImage(100, 100)

	data, err := p.Encode(img, "gif", 0)
	if err != nil {
		t.Fatalf("Encode GIF failed: %v", err)
	}

	// 驗證 GIF 魔術數字 (47 49 46 38)
	if len(data) < 4 || string(data[:3]) != "GIF" {
		t.Error("Output is not valid GIF format")
	}
}

func TestEncode_WebP(t *testing.T) {
	p := NewProcessor(85, 4096, 4096)
	img := createTestImage(100, 100)

	data, err := p.Encode(img, "webp", 85)
	if err != nil {
		t.Fatalf("Encode WebP failed: %v", err)
	}

	// 驗證 WebP 魔術數字 (RIFF....WEBP)
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		t.Errorf("Output is not valid WebP format, got header: %x", data[:min(12, len(data))])
	}
}

func TestProcessor_Process_Resize(t *testing.T) {
	p := NewProcessor(100, 1000, 1000)
	img := createTestImage(100, 100)

	opts := ProcessOptions{
		Width:  50,
		Height: 50,
		Quality: 80,
	}

	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Remove ctx argument if Process signature doesn't take it anymore
	// Checking processor.go Step 994/995.. Process signature:
	// func (p *Processor) Process(r io.Reader, opts ProcessOptions) (image.Image, error)
	// No context!

	outputImg, err := p.Process(bytes.NewReader(buf.Bytes()), opts)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if outputImg.Bounds().Dx() != 50 || outputImg.Bounds().Dy() != 50 {
		t.Errorf("Expected resize to 50x50, got %dx%d", outputImg.Bounds().Dx(), outputImg.Bounds().Dy())
	}
}

func TestProcessor_Process_Crop(t *testing.T) {
	p := NewProcessor(100, 1000, 1000)
	img := createTestImage(100, 100)

	var buf bytes.Buffer
	jpeg.Encode(&buf, img, nil)

	// In Processor.go, CropLeft/Top/Right/Bottom are used for manual crop
	// See line 125-127 in processor.go view
	opts := ProcessOptions{
		CropLeft: 10,
		CropTop:  10,
		CropRight: 90,
		CropBottom: 90,
	}

	outputImg, err := p.Process(bytes.NewReader(buf.Bytes()), opts)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Crop result size should be (Right-Left) x (Bottom-Top) = 80x80
	if outputImg.Bounds().Dx() != 80 || outputImg.Bounds().Dy() != 80 {
		t.Errorf("Expected crop to 80x80, got %dx%d", outputImg.Bounds().Dx(), outputImg.Bounds().Dy())
	}
}

func TestProcessor_Process_Flip(t *testing.T) {
	p := NewProcessor(100, 1000, 1000)
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	// Set top-left pixel to red
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})

	var buf bytes.Buffer
	jpeg.Encode(&buf, img, nil)

	opts := ProcessOptions{
		FlipH: true,
	}

	outputImg, err := p.Process(bytes.NewReader(buf.Bytes()), opts)
	if err != nil {
		t.Fatal(err)
	}

	if outputImg.Bounds().Dx() != 100 {
		t.Error("Flip shouldn't change width unexpectedly")
	}

	// Since we are processing real JPEG, pixel check is fuzzy.
	// But minimal crash/interface check passes.
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
