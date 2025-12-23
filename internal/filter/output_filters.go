package filter

import (
	"image"
	"strconv"
	"strings"
)

// QualityFilter 品質控制濾鏡（標記用，實際在編碼時生效）
type QualityFilter struct{}

// NewQualityFilter 建立品質濾鏡
func NewQualityFilter() *QualityFilter {
	return &QualityFilter{}
}

// Name 返回濾鏡名稱
func (f *QualityFilter) Name() string {
	return "quality"
}

// Apply 品質濾鏡（不修改圖片，只標記品質值）
// params[0]: quality (1-100)
// 注意：實際品質設定在編碼階段處理
func (f *QualityFilter) Apply(img image.Image, params []string) (image.Image, error) {
	// Quality 濾鏡不直接修改圖片
	// 品質值會在 URL 解析階段被提取並用於編碼
	return img, nil
}

// GetQuality 從參數取得品質值
func (f *QualityFilter) GetQuality(params []string) int {
	if len(params) == 0 {
		return 85 // 預設品質
	}

	quality, err := strconv.Atoi(params[0])
	if err != nil || quality < 1 || quality > 100 {
		return 85
	}

	return quality
}

// FormatFilter 格式轉換濾鏡（標記用，實際在編碼時生效）
type FormatFilter struct{}

// NewFormatFilter 建立格式濾鏡
func NewFormatFilter() *FormatFilter {
	return &FormatFilter{}
}

// Name 返回濾鏡名稱
func (f *FormatFilter) Name() string {
	return "format"
}

// Apply 格式濾鏡（不修改圖片，只標記格式）
// params[0]: format (jpeg, png, webp, gif, avif)
// 注意：實際格式轉換在編碼階段處理
func (f *FormatFilter) Apply(img image.Image, params []string) (image.Image, error) {
	// Format 濾鏡不直接修改圖片
	// 格式值會在 URL 解析階段被提取並用於編碼
	return img, nil
}

// GetFormat 從參數取得格式
func (f *FormatFilter) GetFormat(params []string) string {
	if len(params) == 0 {
		return "jpeg"
	}

	format := strings.ToLower(params[0])
	validFormats := map[string]bool{
		"jpeg": true,
		"jpg":  true,
		"png":  true,
		"webp": true,
		"gif":  true,
		"avif": true,
	}

	if validFormats[format] {
		if format == "jpg" {
			return "jpeg"
		}
		return format
	}

	return "jpeg"
}

// StripExifFilter 移除 EXIF 濾鏡
type StripExifFilter struct{}

// NewStripExifFilter 建立移除 EXIF 濾鏡
func NewStripExifFilter() *StripExifFilter {
	return &StripExifFilter{}
}

// Name 返回濾鏡名稱
func (f *StripExifFilter) Name() string {
	return "strip_exif"
}

// Apply 移除 EXIF（重新繪製圖片即可移除 EXIF）
func (f *StripExifFilter) Apply(img image.Image, params []string) (image.Image, error) {
	// 將圖片重新繪製到新的 RGBA 畫布上
	// 這樣會移除所有 metadata（包括 EXIF）
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			result.Set(x, y, img.At(x, y))
		}
	}

	return result, nil
}

// StripICCFilter 移除 ICC Profile 濾鏡
type StripICCFilter struct{}

// NewStripICCFilter 建立移除 ICC Profile 濾鏡
func NewStripICCFilter() *StripICCFilter {
	return &StripICCFilter{}
}

// Name 返回濾鏡名稱
func (f *StripICCFilter) Name() string {
	return "strip_icc"
}

// Apply 移除 ICC Profile（重新繪製圖片）
func (f *StripICCFilter) Apply(img image.Image, params []string) (image.Image, error) {
	// 與 strip_exif 相同處理
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			result.Set(x, y, img.At(x, y))
		}
	}

	return result, nil
}

// AutoOrientFilter 自動方向校正濾鏡
type AutoOrientFilter struct{}

// NewAutoOrientFilter 建立自動方向校正濾鏡
func NewAutoOrientFilter() *AutoOrientFilter {
	return &AutoOrientFilter{}
}

// Name 返回濾鏡名稱
func (f *AutoOrientFilter) Name() string {
	return "autoorient"
}

// Apply 自動方向校正（基於 EXIF）
// 注意：目前簡化實作，只返回原圖
func (f *AutoOrientFilter) Apply(img image.Image, params []string) (image.Image, error) {
	// 實際的 EXIF 方向校正需要讀取 EXIF 資料
	// 這裡使用 imaging 庫的 AutoOrientation
	// 由於我們在載入時可能已經失去 EXIF，只返回原圖
	return img, nil
}
