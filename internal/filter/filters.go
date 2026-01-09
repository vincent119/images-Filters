package filter

import (
	"image"
	"strconv"

	"github.com/disintegration/imaging"
)

// BlurFilter 模糊濾鏡
type BlurFilter struct{}

// NewBlurFilter 建立模糊濾鏡
func NewBlurFilter() *BlurFilter {
	return &BlurFilter{}
}

// Name 返回濾鏡名稱
func (f *BlurFilter) Name() string {
	return "blur"
}

// Apply 應用模糊濾鏡
// params[0]: sigma (模糊程度，預設 1.0)
func (f *BlurFilter) Apply(img image.Image, params []string) (image.Image, error) {
	sigma := 1.0
	if len(params) > 0 {
		if s, err := strconv.ParseFloat(params[0], 64); err == nil && s > 0 {
			sigma = s
		}
	}

	return imaging.Blur(img, sigma), nil
}

// GrayscaleFilter 灰階濾鏡
type GrayscaleFilter struct{}

// NewGrayscaleFilter 建立灰階濾鏡
func NewGrayscaleFilter() *GrayscaleFilter {
	return &GrayscaleFilter{}
}

// Name 返回濾鏡名稱
func (f *GrayscaleFilter) Name() string {
	return "grayscale"
}

// Apply 應用灰階濾鏡
func (f *GrayscaleFilter) Apply(img image.Image, params []string) (image.Image, error) {
	return imaging.Grayscale(img), nil
}

// BrightnessFilter 亮度調整濾鏡
type BrightnessFilter struct{}

// NewBrightnessFilter 建立亮度濾鏡
func NewBrightnessFilter() *BrightnessFilter {
	return &BrightnessFilter{}
}

// Name 返回濾鏡名稱
func (f *BrightnessFilter) Name() string {
	return "brightness"
}

// Apply 應用亮度調整
// params[0]: percentage (-100 ~ 100)
func (f *BrightnessFilter) Apply(img image.Image, params []string) (image.Image, error) {
	percentage := 0.0
	if len(params) > 0 {
		if p, err := strconv.ParseFloat(params[0], 64); err == nil {
			percentage = clamp(p, -100, 100)
		}
	}

	return imaging.AdjustBrightness(img, percentage), nil
}

// ContrastFilter 對比度調整濾鏡
type ContrastFilter struct{}

// NewContrastFilter 建立對比度濾鏡
func NewContrastFilter() *ContrastFilter {
	return &ContrastFilter{}
}

// Name 返回濾鏡名稱
func (f *ContrastFilter) Name() string {
	return "contrast"
}

// Apply 應用對比度調整
// params[0]: percentage (-100 ~ 100)
func (f *ContrastFilter) Apply(img image.Image, params []string) (image.Image, error) {
	percentage := 0.0
	if len(params) > 0 {
		if p, err := strconv.ParseFloat(params[0], 64); err == nil {
			percentage = clamp(p, -100, 100)
		}
	}

	return imaging.AdjustContrast(img, percentage), nil
}

// SaturationFilter 飽和度調整濾鏡
type SaturationFilter struct{}

// NewSaturationFilter 建立飽和度濾鏡
func NewSaturationFilter() *SaturationFilter {
	return &SaturationFilter{}
}

// Name 返回濾鏡名稱
func (f *SaturationFilter) Name() string {
	return "saturation"
}

// Apply 應用飽和度調整
// params[0]: percentage (-100 ~ 100)
func (f *SaturationFilter) Apply(img image.Image, params []string) (image.Image, error) {
	percentage := 0.0
	if len(params) > 0 {
		if p, err := strconv.ParseFloat(params[0], 64); err == nil {
			percentage = clamp(p, -100, 100)
		}
	}

	return imaging.AdjustSaturation(img, percentage), nil
}

// SharpenFilter 銳化濾鏡
type SharpenFilter struct{}

// NewSharpenFilter 建立銳化濾鏡
func NewSharpenFilter() *SharpenFilter {
	return &SharpenFilter{}
}

// Name 返回濾鏡名稱
func (f *SharpenFilter) Name() string {
	return "sharpen"
}

// Apply 應用銳化濾鏡
// params[0]: sigma (銳化程度，預設 1.0)
func (f *SharpenFilter) Apply(img image.Image, params []string) (image.Image, error) {
	sigma := 1.0
	if len(params) > 0 {
		if s, err := strconv.ParseFloat(params[0], 64); err == nil && s > 0 {
			sigma = s
		}
	}

	return imaging.Sharpen(img, sigma), nil
}

// InvertFilter 反色濾鏡
type InvertFilter struct{}

// NewInvertFilter 建立反色濾鏡
func NewInvertFilter() *InvertFilter {
	return &InvertFilter{}
}

// Name 返回濾鏡名稱
func (f *InvertFilter) Name() string {
	return "invert"
}

// Apply 應用反色濾鏡
func (f *InvertFilter) Apply(img image.Image, params []string) (image.Image, error) {
	return imaging.Invert(img), nil
}

// NoOpFilter 無操作濾鏡（用於測試）
type NoOpFilter struct{}

// NewNoOpFilter 建立無操作濾鏡
func NewNoOpFilter() *NoOpFilter {
	return &NoOpFilter{}
}

// Name 返回濾鏡名稱
func (f *NoOpFilter) Name() string {
	return "noop"
}

// Apply 返回原圖
func (f *NoOpFilter) Apply(img image.Image, params []string) (image.Image, error) {
	return img, nil
}

// clamp 限制值在範圍內
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// clampInt 限制整數值在範圍內
func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// clampUint8 限制 uint8 值
func clampUint8(value int) uint8 {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return uint8(value)
}
