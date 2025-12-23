package filter

import (
	"image"
	"testing"
)

// createTestImage 建立測試用圖片
func createTestImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// 填充為白色
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, image.White)
		}
	}
	return img
}

func TestBlurFilter(t *testing.T) {
	filter := NewBlurFilter()

	if filter.Name() != "blur" {
		t.Fatalf("Expected name 'blur', got '%s'", filter.Name())
	}

	img := createTestImage(100, 100)

	// 無參數
	result, err := filter.Apply(img, nil)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result image")
	}

	// 帶參數
	result, err = filter.Apply(img, []string{"3"})
	if err != nil {
		t.Fatalf("Apply with params failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result image")
	}
}

func TestGrayscaleFilter(t *testing.T) {
	filter := NewGrayscaleFilter()

	if filter.Name() != "grayscale" {
		t.Fatalf("Expected name 'grayscale', got '%s'", filter.Name())
	}

	img := createTestImage(100, 100)
	result, err := filter.Apply(img, nil)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result image")
	}
}

func TestBrightnessFilter(t *testing.T) {
	filter := NewBrightnessFilter()

	if filter.Name() != "brightness" {
		t.Fatalf("Expected name 'brightness', got '%s'", filter.Name())
	}

	img := createTestImage(100, 100)

	// 增加亮度
	result, err := filter.Apply(img, []string{"20"})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result image")
	}

	// 降低亮度
	result, err = filter.Apply(img, []string{"-20"})
	if err != nil {
		t.Fatalf("Apply with negative value failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result image")
	}
}

func TestContrastFilter(t *testing.T) {
	filter := NewContrastFilter()

	if filter.Name() != "contrast" {
		t.Fatalf("Expected name 'contrast', got '%s'", filter.Name())
	}

	img := createTestImage(100, 100)

	result, err := filter.Apply(img, []string{"30"})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result image")
	}
}

func TestSaturationFilter(t *testing.T) {
	filter := NewSaturationFilter()

	if filter.Name() != "saturation" {
		t.Fatalf("Expected name 'saturation', got '%s'", filter.Name())
	}

	img := createTestImage(100, 100)

	result, err := filter.Apply(img, []string{"50"})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result image")
	}
}

func TestSharpenFilter(t *testing.T) {
	filter := NewSharpenFilter()

	if filter.Name() != "sharpen" {
		t.Fatalf("Expected name 'sharpen', got '%s'", filter.Name())
	}

	img := createTestImage(100, 100)

	result, err := filter.Apply(img, []string{"1.5"})
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result image")
	}
}

func TestInvertFilter(t *testing.T) {
	filter := NewInvertFilter()

	if filter.Name() != "invert" {
		t.Fatalf("Expected name 'invert', got '%s'", filter.Name())
	}

	img := createTestImage(100, 100)

	result, err := filter.Apply(img, nil)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result image")
	}
}

func TestNoOpFilter(t *testing.T) {
	filter := NewNoOpFilter()

	if filter.Name() != "noop" {
		t.Fatalf("Expected name 'noop', got '%s'", filter.Name())
	}

	img := createTestImage(100, 100)

	result, err := filter.Apply(img, nil)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if result != img {
		t.Fatal("Expected same image")
	}
}

func TestDefaultFiltersRegistered(t *testing.T) {
	// 檢查預設濾鏡已註冊
	expectedFilters := []string{
		"blur", "grayscale", "brightness",
		"contrast", "saturation", "sharpen",
		"invert", "noop",
	}

	for _, name := range expectedFilters {
		if _, exists := Get(name); !exists {
			t.Errorf("Expected filter '%s' to be registered", name)
		}
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		value, min, max, expected float64
	}{
		{50, 0, 100, 50},
		{-10, 0, 100, 0},
		{150, 0, 100, 100},
		{0, 0, 100, 0},
		{100, 0, 100, 100},
	}

	for _, tt := range tests {
		result := clamp(tt.value, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("clamp(%v, %v, %v) = %v; want %v",
				tt.value, tt.min, tt.max, result, tt.expected)
		}
	}
}
