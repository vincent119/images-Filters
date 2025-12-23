package filter

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"strconv"

	"github.com/disintegration/imaging"
)

// RotateFilter 旋轉濾鏡
type RotateFilter struct{}

// NewRotateFilter 建立旋轉濾鏡
func NewRotateFilter() *RotateFilter {
	return &RotateFilter{}
}

// Name 返回濾鏡名稱
func (f *RotateFilter) Name() string {
	return "rotate"
}

// Apply 應用旋轉
// params[0]: degree (角度，可為負數)
func (f *RotateFilter) Apply(img image.Image, params []string) (image.Image, error) {
	degree := 0.0
	if len(params) > 0 {
		if d, err := strconv.ParseFloat(params[0], 64); err == nil {
			degree = d
		}
	}

	if degree == 0 {
		return img, nil
	}

	return imaging.Rotate(img, degree, color.Transparent), nil
}

// RoundCornersFilter 圓角濾鏡
type RoundCornersFilter struct{}

// NewRoundCornersFilter 建立圓角濾鏡
func NewRoundCornersFilter() *RoundCornersFilter {
	return &RoundCornersFilter{}
}

// Name 返回濾鏡名稱
func (f *RoundCornersFilter) Name() string {
	return "round_corner"
}

// Apply 應用圓角效果
// params[0]: radius (圓角半徑)
func (f *RoundCornersFilter) Apply(img image.Image, params []string) (image.Image, error) {
	radius := 10
	if len(params) > 0 {
		if r, err := strconv.Atoi(params[0]); err == nil && r > 0 {
			radius = r
		}
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	result := image.NewRGBA(bounds)

	// 確保半徑不超過圖片尺寸的一半
	maxRadius := minInt(width, height) / 2
	if radius > maxRadius {
		radius = maxRadius
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// 判斷是否在圓角區域外
			if isOutsideRoundedRect(x, y, width, height, radius) {
				result.Set(x+bounds.Min.X, y+bounds.Min.Y, color.Transparent)
			} else {
				result.Set(x+bounds.Min.X, y+bounds.Min.Y, img.At(x+bounds.Min.X, y+bounds.Min.Y))
			}
		}
	}

	return result, nil
}

// isOutsideRoundedRect 判斷點是否在圓角矩形外
func isOutsideRoundedRect(x, y, width, height, radius int) bool {
	// 左上角
	if x < radius && y < radius {
		return !isInsideCircle(x, y, radius, radius, radius)
	}
	// 右上角
	if x >= width-radius && y < radius {
		return !isInsideCircle(x, y, width-radius-1, radius, radius)
	}
	// 左下角
	if x < radius && y >= height-radius {
		return !isInsideCircle(x, y, radius, height-radius-1, radius)
	}
	// 右下角
	if x >= width-radius && y >= height-radius {
		return !isInsideCircle(x, y, width-radius-1, height-radius-1, radius)
	}
	return false
}

// isInsideCircle 判斷點是否在圓內
func isInsideCircle(x, y, cx, cy, r int) bool {
	dx := x - cx
	dy := y - cy
	return dx*dx+dy*dy <= r*r
}

// NoiseFilter 雜訊濾鏡
type NoiseFilter struct{}

// NewNoiseFilter 建立雜訊濾鏡
func NewNoiseFilter() *NoiseFilter {
	return &NoiseFilter{}
}

// Name 返回濾鏡名稱
func (f *NoiseFilter) Name() string {
	return "noise"
}

// Apply 應用雜訊效果
// params[0]: amount (雜訊強度 0~100，預設 20)
func (f *NoiseFilter) Apply(img image.Image, params []string) (image.Image, error) {
	amount := 20.0
	if len(params) > 0 {
		if a, err := strconv.ParseFloat(params[0], 64); err == nil {
			amount = clamp(a, 0, 100)
		}
	}

	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	noiseRange := int(amount * 2.55) // 0-255 的範圍

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// 添加隨機雜訊
			noise := rand.Intn(noiseRange*2+1) - noiseRange

			result.Set(x, y, color.RGBA{
				R: clampUint8(int(r>>8) + noise),
				G: clampUint8(int(g>>8) + noise),
				B: clampUint8(int(b>>8) + noise),
				A: uint8(a >> 8),
			})
		}
	}

	return result, nil
}

// FlipHFilter 水平翻轉濾鏡
type FlipHFilter struct{}

// NewFlipHFilter 建立水平翻轉濾鏡
func NewFlipHFilter() *FlipHFilter {
	return &FlipHFilter{}
}

// Name 返回濾鏡名稱
func (f *FlipHFilter) Name() string {
	return "fliph"
}

// Apply 應用水平翻轉
func (f *FlipHFilter) Apply(img image.Image, params []string) (image.Image, error) {
	return imaging.FlipH(img), nil
}

// FlipVFilter 垂直翻轉濾鏡
type FlipVFilter struct{}

// NewFlipVFilter 建立垂直翻轉濾鏡
func NewFlipVFilter() *FlipVFilter {
	return &FlipVFilter{}
}

// Name 返回濾鏡名稱
func (f *FlipVFilter) Name() string {
	return "flipv"
}

// Apply 應用垂直翻轉
func (f *FlipVFilter) Apply(img image.Image, params []string) (image.Image, error) {
	return imaging.FlipV(img), nil
}

// PixelateFilter 像素化濾鏡
type PixelateFilter struct{}

// NewPixelateFilter 建立像素化濾鏡
func NewPixelateFilter() *PixelateFilter {
	return &PixelateFilter{}
}

// Name 返回濾鏡名稱
func (f *PixelateFilter) Name() string {
	return "pixelate"
}

// Apply 應用像素化效果
// params[0]: size (像素塊大小，預設 10)
func (f *PixelateFilter) Apply(img image.Image, params []string) (image.Image, error) {
	size := 10
	if len(params) > 0 {
		if s, err := strconv.Atoi(params[0]); err == nil && s > 0 {
			size = s
		}
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 縮小再放大實現像素化效果
	newWidth := int(math.Ceil(float64(width) / float64(size)))
	newHeight := int(math.Ceil(float64(height) / float64(size)))

	small := imaging.Resize(img, newWidth, newHeight, imaging.NearestNeighbor)
	return imaging.Resize(small, width, height, imaging.NearestNeighbor), nil
}

// minInt 返回較小的整數
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
