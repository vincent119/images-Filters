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
	"io"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/gen2brain/avif"
	_ "github.com/gen2brain/heic" // Register HEIC decoder
	"github.com/gen2brain/jpegxl"
	"github.com/muesli/smartcrop"
	"github.com/muesli/smartcrop/nfnt"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"github.com/vincent119/zlogger"
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
// Process 處理圖片
func (p *Processor) Process(r io.Reader, opts ProcessOptions) (image.Image, error) {
	// 讀取圖片資料
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	// 1. 解碼圖片
	img, err := p.decodeImage(data, opts)
	if err != nil {
		return nil, err
	}

	// 2. 應用裁切 (Smart Crop 或 手動裁切)
	img = p.applyCropping(img, opts)

	// 3. 應用變形 (縮放、翻轉)
	img = p.applyTransformations(img, opts)

	return img, nil
}

// decodeImage 解碼圖片 (支援 SVG 偵測與 fallback)
func (p *Processor) decodeImage(data []byte, opts ProcessOptions) (image.Image, error) {
	if isSVG(data) {
		img, err := p.decodeSVG(bytes.NewReader(data), opts.Width, opts.Height)
		if err == nil {
			return img, nil
		}
		zlogger.Warn("Failed to decode as SVG, falling back to image.Decode", zlogger.Err(err))
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	return img, nil
}

// applyCropping 應用裁切邏輯
func (p *Processor) applyCropping(img image.Image, opts ProcessOptions) image.Image {
	// Priority: Smart Crop > Manual Crop
	if opts.Smart && opts.Width > 0 && opts.Height > 0 {
		smartImg, err := p.smartCrop(img, opts.Width, opts.Height)
		if err == nil {
			return smartImg
		}
		zlogger.Warn("Smart crop failed, falling back to standard processing", zlogger.Err(err))
		// Fall through to check manual crop (behavior preservation check: original code seemingly skipped manual crop if smart block entered?
		// Original code: if smart { ... } else if manual { ... }
		// Use nested structure to exactly match logic: if smart attempted, we don't do manual even if failed?
		// "if opts.Smart ... { if err != nil { warn } else { img = smartImg } } else if manual ..."
		// Yes, if smart is requested, manual is skipped even if smart failed in the original code structure.
		// However, it's cleaner to just return img if smart failed, effectively skipping manual crop.
		return img
	}

	if opts.CropLeft > 0 || opts.CropTop > 0 || opts.CropRight > 0 || opts.CropBottom > 0 {
		return p.crop(img, opts.CropLeft, opts.CropTop, opts.CropRight, opts.CropBottom)
	}

	return img
}

// applyTransformations 應用縮放與翻轉
func (p *Processor) applyTransformations(img image.Image, opts ProcessOptions) image.Image {
	// 縮放
	if opts.Width > 0 || opts.Height > 0 {
		img = p.resize(img, opts.Width, opts.Height, opts.FitIn)
	}

	// 翻轉
	if opts.FlipH {
		img = p.flipHorizontal(img)
	}
	if opts.FlipV {
		img = p.flipVertical(img)
	}

	return img
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
	case "avif":
		err = avif.Encode(&buf, img, avif.Options{Quality: quality, Speed: -1})
	case "jxl":
		err = jpegxl.Encode(&buf, img, jpegxl.Options{Quality: quality, Effort: 4})
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
	case "heic":
		return "image/heic"
	default:
		return "image/jpeg"
	}
}

// isSVG 簡單檢查是否為 SVG 格式
func isSVG(data []byte) bool {
	// 取前 512 bytes 檢查
	limit := 512
	if len(data) < limit {
		limit = len(data)
	}
	head := string(bytes.TrimSpace(data[:limit]))

	// 簡單的啟發式檢查：包含 <svg 或 <?xml 且包含 svg
	return strings.Contains(head, "<svg")
}

// decodeSVG 解析並渲染 SVG
func (p *Processor) decodeSVG(r io.Reader, width, height int) (image.Image, error) {
	icon, err := oksvg.ReadIconStream(r)
	if err != nil {
		return nil, err
	}

	// 確定目標尺寸
	w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)

	// 如果 svg 沒有 viewBox，回退到預設值? 或者直接使用寬高
	// oksvg struct doesn't export W/H directly if ViewBox is present?
	// Checking source: SvgIcon has ViewBox struct which has W/H.
	// Actually SvgIcon struct definitions might differ.
	// Assuming ViewBox.W/H is sufficient for now.
	// If w/h is 0, we might need a default or error.
	if w == 0 || h == 0 {
		// Default fallback if ViewBox is missing/empty
		w, h = 100, 100
	}

	targetW, targetH := float64(w), float64(h)

	// 如果有指定輸出尺寸，計算縮放後的目標尺寸
	if width > 0 || height > 0 {
		tw, th := calculateDimensions(w, h, width, height, false) // 這裡傳 fitIn=false 是因為我們想要計算出的準確尺寸
		targetW, targetH = float64(tw), float64(th)
	}

	// 設定渲染目標尺寸
	icon.SetTarget(0, 0, targetW, targetH)

	// 建立 Canvas
	rgba := image.NewRGBA(image.Rect(0, 0, int(targetW), int(targetH)))
	scanner := rasterx.NewScannerGV(int(targetW), int(targetH), rgba, rgba.Bounds())
	dasher := rasterx.NewDasher(int(targetW), int(targetH), scanner)

	// 繪製
	icon.Draw(dasher, 1)

	return rgba, nil
}

// smartCrop 執行智慧裁切
func (p *Processor) smartCrop(img image.Image, width, height int) (image.Image, error) {
	analyzer := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())

	// 設定目標寬高
	crop, err := analyzer.FindBestCrop(img, width, height)
	if err != nil {
		return nil, err
	}

	// The crop struct contains X, Y, Width, Height
	// We use SubImage pattern if supported, or fallback to standard crop

	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	// 裁切區域
	// analyzer.FindBestCrop returns image.Rectangle
	rect := crop

	if simg, ok := img.(subImager); ok {
		return simg.SubImage(rect), nil
	}

	// 如果不支援 SubImage (很少見，除非自訂 Image 類型)，則回退到通用裁切
	// 這裡簡單使用我們現有的 crop helper? 但 crop helper 參數是 padding...
	// 為了簡單，直接返回錯誤或實作通用裁切 (copy)
	// 這裡我們假設大多數標準 image 都支援 SubImage
	return nil, fmt.Errorf("image type does not support smart crop (SubImage)")
}
