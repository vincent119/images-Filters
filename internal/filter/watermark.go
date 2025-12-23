package filter

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

// WatermarkPosition 浮水印位置
type WatermarkPosition int

const (
	PositionCenter WatermarkPosition = iota
	PositionTopLeft
	PositionTopRight
	PositionBottomLeft
	PositionBottomRight
	PositionTop
	PositionBottom
	PositionLeft
	PositionRight
)

// WatermarkFilter 浮水印濾鏡
type WatermarkFilter struct {
	baseURL string // 用於載入浮水印圖片的基礎 URL
}

// NewWatermarkFilter 建立浮水印濾鏡
func NewWatermarkFilter() *WatermarkFilter {
	return &WatermarkFilter{}
}

// Name 返回濾鏡名稱
func (f *WatermarkFilter) Name() string {
	return "watermark"
}

// Apply 應用浮水印
// params[0]: watermark image URL or path
// params[1]: position (center, top-left, top-right, bottom-left, bottom-right, top, bottom, left, right)
// params[2]: alpha (0-100, 透明度，100 = 完全不透明)
// params[3]: x offset (可選)
// params[4]: y offset (可選)
// params[5]: scale (可選，浮水印縮放比例 0.1-2.0)
func (f *WatermarkFilter) Apply(img image.Image, params []string) (image.Image, error) {
	if len(params) == 0 {
		return img, nil // 沒有浮水印參數，返回原圖
	}

	// 解析參數
	watermarkPath := params[0]
	position := PositionBottomRight
	alpha := 100
	xOffset := 10
	yOffset := 10
	scale := 1.0

	if len(params) > 1 {
		position = parsePosition(params[1])
	}
	if len(params) > 2 {
		if a, err := strconv.Atoi(params[2]); err == nil {
			alpha = clampInt(a, 0, 100)
		}
	}
	if len(params) > 3 {
		if x, err := strconv.Atoi(params[3]); err == nil {
			xOffset = x
		}
	}
	if len(params) > 4 {
		if y, err := strconv.Atoi(params[4]); err == nil {
			yOffset = y
		}
	}
	if len(params) > 5 {
		if s, err := strconv.ParseFloat(params[5], 64); err == nil && s > 0 {
			scale = clamp(s, 0.1, 2.0)
		}
	}

	// 載入浮水印圖片
	watermark, err := f.loadWatermark(watermarkPath)
	if err != nil {
		// 浮水印載入失敗，返回原圖
		return img, nil
	}

	// 縮放浮水印
	if scale != 1.0 {
		newWidth := int(float64(watermark.Bounds().Dx()) * scale)
		newHeight := int(float64(watermark.Bounds().Dy()) * scale)
		watermark = imaging.Resize(watermark, newWidth, newHeight, imaging.Lanczos)
	}

	// 調整透明度
	if alpha < 100 {
		watermark = adjustAlpha(watermark, float64(alpha)/100.0)
	}

	// 計算位置
	x, y := calculatePosition(img.Bounds(), watermark.Bounds(), position, xOffset, yOffset)

	// 繪製浮水印
	result := image.NewRGBA(img.Bounds())
	draw.Draw(result, img.Bounds(), img, image.Point{}, draw.Over)
	draw.Draw(result, watermark.Bounds().Add(image.Point{X: x, Y: y}), watermark, image.Point{}, draw.Over)

	return result, nil
}

// loadWatermark 載入浮水印圖片
func (f *WatermarkFilter) loadWatermark(path string) (image.Image, error) {
	// 判斷是 URL 還是本地路徑
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return f.loadFromURL(path)
	}
	return f.loadFromFile(path)
}

// loadFromURL 從 URL 載入圖片
func (f *WatermarkFilter) loadFromURL(url string) (image.Image, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// 限制讀取大小（最大 5MB）
	limitReader := io.LimitReader(resp.Body, 5*1024*1024)
	return imaging.Decode(limitReader)
}

// loadFromFile 從檔案載入圖片
func (f *WatermarkFilter) loadFromFile(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return imaging.Decode(file)
}

// parsePosition 解析位置字串
func parsePosition(s string) WatermarkPosition {
	switch strings.ToLower(s) {
	case "center", "c":
		return PositionCenter
	case "top-left", "tl", "topleft":
		return PositionTopLeft
	case "top-right", "tr", "topright":
		return PositionTopRight
	case "bottom-left", "bl", "bottomleft":
		return PositionBottomLeft
	case "bottom-right", "br", "bottomright":
		return PositionBottomRight
	case "top", "t":
		return PositionTop
	case "bottom", "b":
		return PositionBottom
	case "left", "l":
		return PositionLeft
	case "right", "r":
		return PositionRight
	default:
		return PositionBottomRight
	}
}

// calculatePosition 計算浮水印位置
func calculatePosition(imgBounds, wmBounds image.Rectangle, pos WatermarkPosition, xOffset, yOffset int) (int, int) {
	imgW := imgBounds.Dx()
	imgH := imgBounds.Dy()
	wmW := wmBounds.Dx()
	wmH := wmBounds.Dy()

	switch pos {
	case PositionCenter:
		return (imgW - wmW) / 2, (imgH - wmH) / 2
	case PositionTopLeft:
		return xOffset, yOffset
	case PositionTopRight:
		return imgW - wmW - xOffset, yOffset
	case PositionBottomLeft:
		return xOffset, imgH - wmH - yOffset
	case PositionBottomRight:
		return imgW - wmW - xOffset, imgH - wmH - yOffset
	case PositionTop:
		return (imgW - wmW) / 2, yOffset
	case PositionBottom:
		return (imgW - wmW) / 2, imgH - wmH - yOffset
	case PositionLeft:
		return xOffset, (imgH - wmH) / 2
	case PositionRight:
		return imgW - wmW - xOffset, (imgH - wmH) / 2
	default:
		return imgW - wmW - xOffset, imgH - wmH - yOffset
	}
}

// adjustAlpha 調整圖片透明度
func adjustAlpha(img image.Image, alpha float64) *image.NRGBA {
	bounds := img.Bounds()
	result := image.NewNRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()

			// 調整 alpha 通道
			newAlpha := uint8(float64(a>>8) * alpha)

			result.Set(x, y, nrgbaColor{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: newAlpha,
			})
		}
	}

	return result
}

// nrgbaColor 是 NRGBA 顏色結構
type nrgbaColor struct {
	R, G, B, A uint8
}

// RGBA implements color.Color interface
func (c nrgbaColor) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = uint32(c.A)
	a |= a << 8
	return
}
