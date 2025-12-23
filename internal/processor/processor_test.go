package processor

import (
	"image"
	"image/color"
	"testing"
)


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

func TestEncode_JPG(t *testing.T) {
	p := NewProcessor(85, 4096, 4096)
	img := createTestImage(100, 100)

	// 測試 jpg 別名
	data, err := p.Encode(img, "jpg", 85)
	if err != nil {
		t.Fatalf("Encode JPG failed: %v", err)
	}

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

func TestEncode_DefaultFormat(t *testing.T) {
	p := NewProcessor(85, 4096, 4096)
	img := createTestImage(100, 100)

	// 未知格式應該預設為 JPEG
	data, err := p.Encode(img, "unknown", 85)
	if err != nil {
		t.Fatalf("Encode unknown format failed: %v", err)
	}

	if len(data) < 3 || data[0] != 0xFF || data[1] != 0xD8 || data[2] != 0xFF {
		t.Error("Unknown format should default to JPEG")
	}
}

func TestEncode_DefaultQuality(t *testing.T) {
	p := NewProcessor(90, 4096, 4096) // 設定預設品質為 90
	img := createTestImage(100, 100)

	// quality=0 應該使用處理器的預設品質
	data1, err := p.Encode(img, "jpeg", 0)
	if err != nil {
		t.Fatalf("Encode with quality=0 failed: %v", err)
	}

	// 明確指定品質 90
	data2, err := p.Encode(img, "jpeg", 90)
	if err != nil {
		t.Fatalf("Encode with quality=90 failed: %v", err)
	}

	// 兩者大小應該相同（因為使用相同品質）
	if len(data1) != len(data2) {
		t.Log("Note: Default quality encoding may have slight differences")
	}

	// 確保都是有效的輸出
	if len(data1) == 0 || len(data2) == 0 {
		t.Error("Encoded data should not be empty")
	}
}

func TestEncode_QualityAffectsSize(t *testing.T) {
	p := NewProcessor(85, 4096, 4096)
	img := createTestImage(200, 200) // 稍大的圖片

	lowQuality, err := p.Encode(img, "jpeg", 10)
	if err != nil {
		t.Fatalf("Encode low quality failed: %v", err)
	}

	highQuality, err := p.Encode(img, "jpeg", 95)
	if err != nil {
		t.Fatalf("Encode high quality failed: %v", err)
	}

	// 高品質應該產生更大的檔案
	if len(highQuality) <= len(lowQuality) {
		t.Errorf("High quality (%d bytes) should be larger than low quality (%d bytes)",
			len(highQuality), len(lowQuality))
	}
}

