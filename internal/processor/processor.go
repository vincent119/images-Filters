// Package processor 提供圖片處理功能
// 包含 Resize、Crop、Flip 等核心圖片操作
package processor

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

// Processor 圖片處理器
type Processor struct {
	// 處理品質（1-100）
	Quality int
	// 最大寬度限制
	MaxWidth int
	// 最大高度限制
	MaxHeight int
}

// ProcessOptions 處理選項
type ProcessOptions struct {
	// 尺寸
	Width  int
	Height int

	// 翻轉
	FlipH bool
	FlipV bool

	// Fit-in 模式（不裁切，保持比例）
	FitIn bool

	// 裁切
	CropLeft   int
	CropTop    int
	CropRight  int
	CropBottom int

	// Smart 裁切
	Smart bool

	// 輸出品質
	Quality int

	// 輸出格式
	Format string
}

// NewProcessor 建立新的處理器
func NewProcessor(quality, maxWidth, maxHeight int) *Processor {
	return &Processor{
		Quality:   quality,
		MaxWidth:  maxWidth,
		MaxHeight: maxHeight,
	}
}

// Process 處理圖片
func (p *Processor) Process(data []byte, opts ProcessOptions) (image.Image, error) {
	// 解碼圖片
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// 執行裁切
	if opts.CropLeft > 0 || opts.CropTop > 0 || opts.CropRight > 0 || opts.CropBottom > 0 {
		img = p.crop(img, opts.CropLeft, opts.CropTop, opts.CropRight, opts.CropBottom)
	}

	// 執行縮放
	if opts.Width > 0 || opts.Height > 0 {
		img = p.resize(img, opts.Width, opts.Height, opts.FitIn)
	}

	// 執行翻轉
	if opts.FlipH {
		img = p.flipHorizontal(img)
	}
	if opts.FlipV {
		img = p.flipVertical(img)
	}

	return img, nil
}

// DecodeImage 解碼圖片資料
func (p *Processor) DecodeImage(data []byte) (image.Image, string, error) {
	return image.Decode(bytes.NewReader(data))
}

// GetImageSize 取得圖片尺寸
func (p *Processor) GetImageSize(data []byte) (width, height int, err error) {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image config: %w", err)
	}
	return config.Width, config.Height, nil
}

// resize 縮放圖片
func (p *Processor) resize(img image.Image, width, height int, fitIn bool) image.Image {
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// 驗證尺寸限制
	if p.MaxWidth > 0 && width > p.MaxWidth {
		width = p.MaxWidth
	}
	if p.MaxHeight > 0 && height > p.MaxHeight {
		height = p.MaxHeight
	}

	// 計算目標尺寸
	targetWidth, targetHeight := calculateDimensions(
		originalWidth, originalHeight,
		width, height,
		fitIn,
	)

	// 執行縮放
	return imaging.Resize(img, targetWidth, targetHeight, imaging.Lanczos)
}

// crop 裁切圖片
func (p *Processor) crop(img image.Image, left, top, right, bottom int) image.Image {
	return imaging.Crop(img, image.Rect(left, top, right, bottom))
}

// flipHorizontal 水平翻轉
func (p *Processor) flipHorizontal(img image.Image) image.Image {
	return imaging.FlipH(img)
}

// flipVertical 垂直翻轉
func (p *Processor) flipVertical(img image.Image) image.Image {
	return imaging.FlipV(img)
}

// calculateDimensions 計算目標尺寸
func calculateDimensions(originalWidth, originalHeight, targetWidth, targetHeight int, fitIn bool) (int, int) {
	// 如果兩個尺寸都是 0，返回原尺寸
	if targetWidth == 0 && targetHeight == 0 {
		return originalWidth, originalHeight
	}

	// 計算原始比例
	ratio := float64(originalWidth) / float64(originalHeight)

	// 只指定寬度
	if targetHeight == 0 {
		return targetWidth, int(float64(targetWidth) / ratio)
	}

	// 只指定高度
	if targetWidth == 0 {
		return int(float64(targetHeight) * ratio), targetHeight
	}

	// 兩個尺寸都指定
	if fitIn {
		// Fit-in 模式：縮放到能完全放入指定尺寸內，保持比例
		targetRatio := float64(targetWidth) / float64(targetHeight)
		if ratio > targetRatio {
			// 圖片較寬，以寬度為準
			return targetWidth, int(float64(targetWidth) / ratio)
		}
		// 圖片較高，以高度為準
		return int(float64(targetHeight) * ratio), targetHeight
	}

	// 填滿模式：返回指定尺寸（可能會裁切）
	return targetWidth, targetHeight
}

// Encode 編碼圖片
func (p *Processor) Encode(img image.Image, format string, quality int) ([]byte, error) {
	if quality == 0 {
		quality = p.Quality
	}

	var buf bytes.Buffer
	var err error

	switch format {

	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(&buf, img)
	case "gif":
		err = gif.Encode(&buf, img, nil)
	case "webp":
		err = webp.Encode(&buf, img, &webp.Options{Quality: float32(quality)})
	default:
		// 預設使用 JPEG
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

// GetContentType 根據格式取得 Content-Type
func GetContentType(format string) string {
	switch format {
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "avif":
		return "image/avif"
	case "jxl":
		return "image/jxl"
	default:
		return "image/jpeg"
	}
}
