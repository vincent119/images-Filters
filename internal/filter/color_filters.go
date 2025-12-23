package filter

import (
	"image"
	"image/color"
	"strconv"

	"github.com/disintegration/imaging"
)

// RGBFilter RGB 調整濾鏡
type RGBFilter struct{}

// NewRGBFilter 建立 RGB 濾鏡
func NewRGBFilter() *RGBFilter {
	return &RGBFilter{}
}

// Name 返回濾鏡名稱
func (f *RGBFilter) Name() string {
	return "rgb"
}

// Apply 應用 RGB 調整
// params[0]: r adjustment (-100 ~ 100)
// params[1]: g adjustment (-100 ~ 100)
// params[2]: b adjustment (-100 ~ 100)
func (f *RGBFilter) Apply(img image.Image, params []string) (image.Image, error) {
	rAdj, gAdj, bAdj := 0.0, 0.0, 0.0

	if len(params) > 0 {
		if v, err := strconv.ParseFloat(params[0], 64); err == nil {
			rAdj = clamp(v, -100, 100) / 100
		}
	}
	if len(params) > 1 {
		if v, err := strconv.ParseFloat(params[1], 64); err == nil {
			gAdj = clamp(v, -100, 100) / 100
		}
	}
	if len(params) > 2 {
		if v, err := strconv.ParseFloat(params[2], 64); err == nil {
			bAdj = clamp(v, -100, 100) / 100
		}
	}

	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// 調整 RGB 值
			newR := float64(r>>8) * (1 + rAdj)
			newG := float64(g>>8) * (1 + gAdj)
			newB := float64(b>>8) * (1 + bAdj)

			result.Set(x, y, color.RGBA{
				R: clampUint8(int(newR)),
				G: clampUint8(int(newG)),
				B: clampUint8(int(newB)),
				A: uint8(a >> 8),
			})
		}
	}

	return result, nil
}

// SepiaFilter 復古色調濾鏡
type SepiaFilter struct{}

// NewSepiaFilter 建立復古濾鏡
func NewSepiaFilter() *SepiaFilter {
	return &SepiaFilter{}
}

// Name 返回濾鏡名稱
func (f *SepiaFilter) Name() string {
	return "sepia"
}

// Apply 應用復古色調
// params[0]: intensity (0 ~ 100, 預設 100)
func (f *SepiaFilter) Apply(img image.Image, params []string) (image.Image, error) {
	intensity := 100.0
	if len(params) > 0 {
		if v, err := strconv.ParseFloat(params[0], 64); err == nil {
			intensity = clamp(v, 0, 100)
		}
	}

	factor := intensity / 100.0
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()

			// 原始值
			origR := float64(r >> 8)
			origG := float64(g >> 8)
			origB := float64(b >> 8)

			// Sepia 轉換公式
			sepiaR := 0.393*origR + 0.769*origG + 0.189*origB
			sepiaG := 0.349*origR + 0.686*origG + 0.168*origB
			sepiaB := 0.272*origR + 0.534*origG + 0.131*origB

			// 混合原始和 sepia
			newR := origR*(1-factor) + sepiaR*factor
			newG := origG*(1-factor) + sepiaG*factor
			newB := origB*(1-factor) + sepiaB*factor

			result.Set(x, y, color.RGBA{
				R: clampUint8(int(newR)),
				G: clampUint8(int(newG)),
				B: clampUint8(int(newB)),
				A: uint8(a >> 8),
			})
		}
	}

	return result, nil
}

// EqualizeFilter 均衡化濾鏡（使用 Gamma 校正近似）
type EqualizeFilter struct{}

// NewEqualizeFilter 建立均衡化濾鏡
func NewEqualizeFilter() *EqualizeFilter {
	return &EqualizeFilter{}
}

// Name 返回濾鏡名稱
func (f *EqualizeFilter) Name() string {
	return "equalize"
}

// Apply 應用均衡化（使用 imaging 的 Gamma 校正）
func (f *EqualizeFilter) Apply(img image.Image, params []string) (image.Image, error) {
	// 使用 auto-contrast 效果
	return imaging.AdjustContrast(imaging.AdjustGamma(img, 1.2), 10), nil
}

// GammaFilter Gamma 校正濾鏡
type GammaFilter struct{}

// NewGammaFilter 建立 Gamma 濾鏡
func NewGammaFilter() *GammaFilter {
	return &GammaFilter{}
}

// Name 返回濾鏡名稱
func (f *GammaFilter) Name() string {
	return "gamma"
}

// Apply 應用 Gamma 校正
// params[0]: gamma value (0.1 ~ 10, 預設 1.0)
func (f *GammaFilter) Apply(img image.Image, params []string) (image.Image, error) {
	gamma := 1.0
	if len(params) > 0 {
		if v, err := strconv.ParseFloat(params[0], 64); err == nil && v > 0 {
			gamma = clamp(v, 0.1, 10)
		}
	}

	return imaging.AdjustGamma(img, gamma), nil
}

// HueFilter 色相調整濾鏡
type HueFilter struct{}

// NewHueFilter 建立色相濾鏡
func NewHueFilter() *HueFilter {
	return &HueFilter{}
}

// Name 返回濾鏡名稱
func (f *HueFilter) Name() string {
	return "hue"
}

// Apply 應用色相調整（簡化版，使用色調旋轉）
// params[0]: degree (-180 ~ 180)
func (f *HueFilter) Apply(img image.Image, params []string) (image.Image, error) {
	degree := 0.0
	if len(params) > 0 {
		if v, err := strconv.ParseFloat(params[0], 64); err == nil {
			degree = clamp(v, -180, 180)
		}
	}

	if degree == 0 {
		return img, nil
	}

	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// 簡化的色相旋轉
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			h, s, l := rgbToHSL(uint8(r>>8), uint8(g>>8), uint8(b>>8))

			// 調整色相
			h += degree / 360.0
			if h > 1 {
				h -= 1
			} else if h < 0 {
				h += 1
			}

			newR, newG, newB := hslToRGB(h, s, l)
			result.Set(x, y, color.RGBA{
				R: newR,
				G: newG,
				B: newB,
				A: uint8(a >> 8),
			})
		}
	}

	return result, nil
}

// rgbToHSL 將 RGB 轉換為 HSL
func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	rf := float64(r) / 255
	gf := float64(g) / 255
	bf := float64(b) / 255

	max := maxFloat(rf, maxFloat(gf, bf))
	min := minFloat(rf, minFloat(gf, bf))
	l = (max + min) / 2

	if max == min {
		h = 0
		s = 0
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case rf:
			h = (gf - bf) / d
			if gf < bf {
				h += 6
			}
		case gf:
			h = (bf-rf)/d + 2
		case bf:
			h = (rf-gf)/d + 4
		}
		h /= 6
	}
	return
}

// hslToRGB 將 HSL 轉換為 RGB
func hslToRGB(h, s, l float64) (r, g, b uint8) {
	var rf, gf, bf float64

	if s == 0 {
		rf = l
		gf = l
		bf = l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q
		rf = hueToRGB(p, q, h+1.0/3.0)
		gf = hueToRGB(p, q, h)
		bf = hueToRGB(p, q, h-1.0/3.0)
	}

	return uint8(rf * 255), uint8(gf * 255), uint8(bf * 255)
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
