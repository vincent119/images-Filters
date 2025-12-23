package filter

import (
	"fmt"
	"image"
	"image/color"
	"testing"
)

// 測試用濾鏡：反轉顏色
type invertFilter struct{}

func (f *invertFilter) Name() string { return "invert" }

func (f *invertFilter) Apply(img image.Image, params []string) (image.Image, error) {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			result.Set(x, y, color.RGBA{
				R: uint8(255 - r>>8),
				G: uint8(255 - g>>8),
				B: uint8(255 - b>>8),
				A: uint8(a >> 8),
			})
		}
	}

	return result, nil
}

// 測試用濾鏡：帶參數的亮度調整
type brightnessFilter struct{}

func (f *brightnessFilter) Name() string { return "brightness" }

func (f *brightnessFilter) Apply(img image.Image, params []string) (image.Image, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("brightness filter requires a parameter")
	}
	// 簡單返回原圖（測試用）
	return img, nil
}

// 測試用濾鏡：失敗
type failFilter struct{}

func (f *failFilter) Name() string { return "fail" }

func (f *failFilter) Apply(img image.Image, params []string) (image.Image, error) {
	return nil, fmt.Errorf("intentional failure")
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	// 註冊成功
	err := r.Register(&invertFilter{})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 重複註冊應失敗
	err = r.Register(&invertFilter{})
	if err == nil {
		t.Fatal("Expected error for duplicate registration")
	}
}

func TestRegistry_Get(t *testing.T) {
	r := NewRegistry()
	r.Register(&invertFilter{})

	// 取得存在的濾鏡
	filter, exists := r.Get("invert")
	if !exists {
		t.Fatal("Expected filter to exist")
	}
	if filter.Name() != "invert" {
		t.Fatalf("Expected filter name 'invert', got '%s'", filter.Name())
	}

	// 取得不存在的濾鏡
	_, exists = r.Get("nonexistent")
	if exists {
		t.Fatal("Expected filter not to exist")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.Register(&invertFilter{})
	r.Register(&brightnessFilter{})

	names := r.List()
	if len(names) != 2 {
		t.Fatalf("Expected 2 filters, got %d", len(names))
	}
}

func TestRegistry_Count(t *testing.T) {
	r := NewRegistry()
	if r.Count() != 0 {
		t.Fatal("Expected 0 filters")
	}

	r.Register(&invertFilter{})
	if r.Count() != 1 {
		t.Fatal("Expected 1 filter")
	}
}

func TestPipeline_Apply(t *testing.T) {
	r := NewRegistry()
	r.Register(&invertFilter{})
	r.Register(&brightnessFilter{})

	// 建立測試圖片
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// 空管線
	p := NewPipeline(r)
	result, err := p.Apply(img)
	if err != nil {
		t.Fatalf("Empty pipeline failed: %v", err)
	}
	if result != img {
		t.Fatal("Empty pipeline should return original image")
	}

	// 單一濾鏡
	p.Add(FilterSpec{Name: "invert", Params: nil})
	result, err = p.Apply(img)
	if err != nil {
		t.Fatalf("Single filter pipeline failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result image")
	}
}

func TestPipeline_ApplyWithUnknownFilter(t *testing.T) {
	r := NewRegistry()
	r.Register(&invertFilter{})

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// 包含未知濾鏡（應該被跳過）
	p := NewPipeline(r)
	p.Add(FilterSpec{Name: "unknown", Params: nil})
	p.Add(FilterSpec{Name: "invert", Params: nil})

	result, err := p.Apply(img)
	if err != nil {
		t.Fatalf("Pipeline with unknown filter failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected result image")
	}
}

func TestPipeline_ApplyWithFailingFilter(t *testing.T) {
	r := NewRegistry()
	r.Register(&failFilter{})

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	p := NewPipeline(r)
	p.Add(FilterSpec{Name: "fail", Params: nil})

	_, err := p.Apply(img)
	if err == nil {
		t.Fatal("Expected error from failing filter")
	}
}

func TestPipeline_Count(t *testing.T) {
	p := NewPipeline(nil)
	if p.Count() != 0 {
		t.Fatal("Expected 0 filters")
	}

	p.Add(FilterSpec{Name: "a", Params: nil})
	p.Add(FilterSpec{Name: "b", Params: nil})
	if p.Count() != 2 {
		t.Fatal("Expected 2 filters")
	}

	p.Clear()
	if p.Count() != 0 {
		t.Fatal("Expected 0 filters after clear")
	}
}

func TestFilterFunc(t *testing.T) {
	ff := NewFilterFunc("test", func(img image.Image, params []string) (image.Image, error) {
		return img, nil
	})

	if ff.Name() != "test" {
		t.Fatalf("Expected name 'test', got '%s'", ff.Name())
	}

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	result, err := ff.Apply(img, nil)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if result != img {
		t.Fatal("Expected same image")
	}
}
