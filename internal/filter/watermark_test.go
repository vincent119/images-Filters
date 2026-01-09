package filter

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWatermarkFilter_Name(t *testing.T) {
	f := NewWatermarkFilter()
	assert.Equal(t, "watermark", f.Name())
}

func TestParsePosition(t *testing.T) {
	tests := []struct {
		input    string
		expected WatermarkPosition
	}{
		{"center", PositionCenter},
		{"c", PositionCenter},
		{"top-left", PositionTopLeft},
		{"tl", PositionTopLeft},
		{"topleft", PositionTopLeft},
		{"top-right", PositionTopRight},
		{"tr", PositionTopRight},
		{"topright", PositionTopRight},
		{"bottom-left", PositionBottomLeft},
		{"bl", PositionBottomLeft},
		{"bottomleft", PositionBottomLeft},
		{"bottom-right", PositionBottomRight},
		{"br", PositionBottomRight},
		{"bottomright", PositionBottomRight},
		{"top", PositionTop},
		{"t", PositionTop},
		{"bottom", PositionBottom},
		{"b", PositionBottom},
		{"left", PositionLeft},
		{"l", PositionLeft},
		{"right", PositionRight},
		{"r", PositionRight},
		{"unknown", PositionBottomRight},
		{"", PositionBottomRight},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, parsePosition(tt.input))
		})
	}
}

func TestCalculatePosition(t *testing.T) {
	imgBounds := image.Rect(0, 0, 1000, 1000)
	wmBounds := image.Rect(0, 0, 100, 100)
	xOffset, yOffset := 10, 20

	tests := []struct {
		pos   WatermarkPosition
		wantX int
		wantY int
	}{
		{PositionCenter, 450, 450},       // (1000-100)/2, (1000-100)/2
		{PositionTopLeft, 10, 20},        // x, y
		{PositionTopRight, 890, 20},      // 1000-100-10, y
		{PositionBottomLeft, 10, 880},    // x, 1000-100-20
		{PositionBottomRight, 890, 880},  // 1000-100-10, 1000-100-20
		{PositionTop, 450, 20},           // center, y
		{PositionBottom, 450, 880},       // center, bottom
		{PositionLeft, 10, 450},          // x, center
		{PositionRight, 890, 450},        // right, center
		{WatermarkPosition(999), 890, 880}, // default to BottomRight
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Pos-%d", tt.pos), func(t *testing.T) {
			x, y := calculatePosition(imgBounds, wmBounds, tt.pos, xOffset, yOffset)
			assert.Equal(t, tt.wantX, x)
			assert.Equal(t, tt.wantY, y)
		})
	}
}

func TestAdjustAlpha(t *testing.T) {
	// Create a red image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, red)
		}
	}

	// 50% transparency
	result := adjustAlpha(img, 0.5)
	c := result.At(0, 0).(color.NRGBA)

	assert.Equal(t, uint8(255), c.R)
	assert.Equal(t, uint8(0), c.G)
	assert.Equal(t, uint8(0), c.B)
	// Alpha should be around 127/128 depending on rounding
	assert.InDelta(t, 127, c.A, 1)
}

func TestWatermarkFilter_LoadFromFile(t *testing.T) {
	// Create a temporary image file
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test_watermark.png")

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	f, err := os.Create(imgPath)
	assert.NoError(t, err)
	err = png.Encode(f, img)
	f.Close()
	assert.NoError(t, err)

	filter := NewWatermarkFilter()
	loadedImg, err := filter.loadFromFile(imgPath)
	assert.NoError(t, err)
	assert.NotNil(t, loadedImg)
	assert.Equal(t, 10, loadedImg.Bounds().Dx())

	// Test non-existent file
	_, err = filter.loadFromFile(filepath.Join(tmpDir, "non_existent.png"))
	assert.Error(t, err)
}

func TestWatermarkFilter_LoadFromURL(t *testing.T) {
	// Create a test server serving an image
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		png.Encode(w, img)
	}))
	defer server.Close()

	filter := NewWatermarkFilter()
	loadedImg, err := filter.loadFromURL(server.URL)
	assert.NoError(t, err)
	assert.NotNil(t, loadedImg)
	assert.Equal(t, 10, loadedImg.Bounds().Dx())

	// Test invalid URL (should timeout or fail dns)
	// Using a closed port on localhost usually fails fast
	_, err = filter.loadFromURL("http://localhost:54321/invalid")
	assert.Error(t, err)

	// Test 404
	server404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server404.Close()

	_, err = filter.loadFromURL(server404.URL)
	assert.Error(t, err)
}

func TestWatermarkFilter_Apply(t *testing.T) {
	// Create a temporary watermark file
	tmpDir := t.TempDir()
	wmPath := filepath.Join(tmpDir, "wm.png")
	wmImg := image.NewRGBA(image.Rect(0, 0, 50, 50))
	// Fill watermark with red
	draw.Draw(wmImg, wmImg.Bounds(), &image.Uniform{color.RGBA{255, 0, 0, 255}}, image.Point{}, draw.Src)

	f, _ := os.Create(wmPath)
	png.Encode(f, wmImg)
	f.Close()

	filter := NewWatermarkFilter()
	bgImg := image.NewRGBA(image.Rect(0, 0, 200, 200))
	// Fill background with white
	draw.Draw(bgImg, bgImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Test case 1: No params
	res, err := filter.Apply(bgImg, nil)
	assert.NoError(t, err)
	assert.Equal(t, bgImg, res)

	// Test case 2: Basic apply (bottom-right, default)
	res, err = filter.Apply(bgImg, []string{wmPath})
	assert.NoError(t, err)
	assert.NotNil(t, res)

	// Test case 3: With Position and Alpha
	// Pos: Center (br), Alpha: 50
	res, err = filter.Apply(bgImg, []string{wmPath, "center", "50"})
	assert.NoError(t, err)

	// Test case 4: Full params
	// Path, Pos, Alpha, X, Y, Scale
	res, err = filter.Apply(bgImg, []string{wmPath, "tl", "80", "5", "5", "0.5"})
	assert.NoError(t, err)
	// Watermark is 50x50, scale 0.5 -> 25x25. Pos TL, offset 5,5.
	// We can check a pixel at 10,10. It should be red-ish mixed with white.
	// BG is white (255,255,255), WM is Red (255,0,0) at 80% alpha.
	// Result R should be mixed.

	// Test case 5: Invalid watermark path
	res, err = filter.Apply(bgImg, []string{"/invalid/path/to/wm.png"})
	assert.NoError(t, err)
	assert.Equal(t, bgImg, res) // Should return original image on failure
}
