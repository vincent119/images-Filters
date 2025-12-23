// Package parser 提供 URL 解析功能
// 解析圖片處理請求的 URL，包含尺寸、濾鏡、裁切等參數
package parser

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// ParsedURL 解析後的 URL 結構
type ParsedURL struct {
	// 安全相關
	Signature string // HMAC 簽名
	IsUnsafe  bool   // 是否為 unsafe 模式

	// 尺寸相關
	Width  int  // 目標寬度（0 表示自動計算）
	Height int  // 目標高度（0 表示自動計算）
	FlipH  bool // 水平翻轉
	FlipV  bool // 垂直翻轉
	FitIn  bool // Fit-in 模式（不裁切，保持比例）

	// 裁切相關
	CropLeft   int  // 裁切左邊界
	CropTop    int  // 裁切上邊界
	CropRight  int  // 裁切右邊界
	CropBottom int  // 裁切下邊界
	Smart      bool // 智慧裁切模式

	// 濾鏡
	Filters []Filter

	// 圖片來源
	ImagePath string // 原始圖片路徑或 URL
}

// Filter 濾鏡參數
type Filter struct {
	Name   string   // 濾鏡名稱
	Params []string // 濾鏡參數
}

// URLParser URL 解析器
type URLParser struct {
	// 正則表達式
	sizeRegex    *regexp.Regexp
	cropRegex    *regexp.Regexp
	filtersRegex *regexp.Regexp
	filterRegex  *regexp.Regexp
}

// NewURLParser 建立新的 URL 解析器
func NewURLParser() *URLParser {
	return &URLParser{
		// 尺寸格式：-?(\d*)x-?(\d*)
		sizeRegex: regexp.MustCompile(`^(-)?(\d*)x(-)?(\d*)$`),
		// 裁切格式：(\d+)x(\d+):(\d+)x(\d+)
		cropRegex: regexp.MustCompile(`^(\d+)x(\d+):(\d+)x(\d+)$`),
		// 濾鏡格式：filters:...
		filtersRegex: regexp.MustCompile(`^filters:(.+)$`),
		// 單個濾鏡格式：name(params)
		filterRegex: regexp.MustCompile(`^(\w+)\((.*?)\)$`),
	}
}

// Parse 解析 URL 路徑
// 路徑格式：/<signature>/<options>/<filters>/<image_path>
// 或：/unsafe/<options>/<filters>/<image_path>
func (p *URLParser) Parse(path string) (*ParsedURL, error) {
	// 移除開頭的斜線
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		return nil, fmt.Errorf("空的 URL 路徑")
	}

	result := &ParsedURL{}
	parts := strings.Split(path, "/")

	// 至少需要兩個部分：安全標記 + 圖片路徑
	if len(parts) < 2 {
		return nil, fmt.Errorf("URL 路徑格式錯誤：缺少必要參數")
	}

	// 解析第一個部分：簽名或 unsafe
	idx := 0
	if parts[idx] == "unsafe" {
		result.IsUnsafe = true
		idx++
	} else {
		result.Signature = parts[idx]
		idx++
	}

	// 解析剩餘部分
	for idx < len(parts) {
		part := parts[idx]

		// 檢查是否為 fit-in
		if part == "fit-in" {
			result.FitIn = true
			idx++
			continue
		}

		// 檢查是否為 smart
		if part == "smart" {
			result.Smart = true
			idx++
			continue
		}

		// 嘗試解析尺寸
		if p.parseSize(part, result) {
			idx++
			continue
		}

		// 嘗試解析裁切
		if p.parseCrop(part, result) {
			idx++
			continue
		}

		// 嘗試解析濾鏡
		if strings.HasPrefix(part, "filters:") {
			if err := p.parseFilters(part, result); err != nil {
				return nil, err
			}
			idx++
			continue
		}

		// 剩餘部分視為圖片路徑
		imagePath := strings.Join(parts[idx:], "/")
		decodedPath, err := url.QueryUnescape(imagePath)
		if err != nil {
			// 如果解碼失敗，使用原始路徑
			result.ImagePath = imagePath
		} else {
			result.ImagePath = decodedPath
		}
		break
	}

	// 驗證結果
	if result.ImagePath == "" {
		return nil, fmt.Errorf("缺少圖片路徑")
	}

	return result, nil
}

// parseSize 解析尺寸
// 格式：300x200, -300x200, 300x-200, -300x-200, 300x0, 0x200
func (p *URLParser) parseSize(s string, result *ParsedURL) bool {
	matches := p.sizeRegex.FindStringSubmatch(s)
	if matches == nil {
		return false
	}

	// matches[1] = 水平翻轉標記 "-"
	// matches[2] = 寬度
	// matches[3] = 垂直翻轉標記 "-"
	// matches[4] = 高度

	if matches[1] == "-" {
		result.FlipH = true
	}
	if matches[3] == "-" {
		result.FlipV = true
	}

	if matches[2] != "" {
		width, err := strconv.Atoi(matches[2])
		if err == nil {
			result.Width = width
		}
	}

	if matches[4] != "" {
		height, err := strconv.Atoi(matches[4])
		if err == nil {
			result.Height = height
		}
	}

	return true
}

// parseCrop 解析裁切座標
// 格式：10x20:100x150 （左上角 x 右下角）
func (p *URLParser) parseCrop(s string, result *ParsedURL) bool {
	matches := p.cropRegex.FindStringSubmatch(s)
	if matches == nil {
		return false
	}

	result.CropLeft, _ = strconv.Atoi(matches[1])
	result.CropTop, _ = strconv.Atoi(matches[2])
	result.CropRight, _ = strconv.Atoi(matches[3])
	result.CropBottom, _ = strconv.Atoi(matches[4])

	return true
}

// parseFilters 解析濾鏡
// 格式：filters:blur(7):grayscale():brightness(50)
func (p *URLParser) parseFilters(s string, result *ParsedURL) error {
	matches := p.filtersRegex.FindStringSubmatch(s)
	if matches == nil {
		return fmt.Errorf("無效的濾鏡格式: %s", s)
	}

	filterStr := matches[1]
	// 分割多個濾鏡
	filterParts := splitFilters(filterStr)

	for _, part := range filterParts {
		filter, err := p.parseFilter(part)
		if err != nil {
			return err
		}
		result.Filters = append(result.Filters, filter)
	}

	return nil
}

// parseFilter 解析單個濾鏡
func (p *URLParser) parseFilter(s string) (Filter, error) {
	matches := p.filterRegex.FindStringSubmatch(s)
	if matches == nil {
		return Filter{}, fmt.Errorf("無效的濾鏡格式: %s", s)
	}

	filter := Filter{
		Name: matches[1],
	}

	// 解析參數
	if matches[2] != "" {
		params := strings.Split(matches[2], ",")
		for _, param := range params {
			filter.Params = append(filter.Params, strings.TrimSpace(param))
		}
	}

	return filter, nil
}

// splitFilters 分割濾鏡字串
// 處理如 blur(7):grayscale():watermark(img,10,10,50) 的情況
func splitFilters(s string) []string {
	var filters []string
	var current strings.Builder
	depth := 0

	for _, c := range s {
		switch c {
		case '(':
			depth++
			current.WriteRune(c)
		case ')':
			depth--
			current.WriteRune(c)
		case ':':
			if depth == 0 {
				if current.Len() > 0 {
					filters = append(filters, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(c)
			}
		default:
			current.WriteRune(c)
		}
	}

	if current.Len() > 0 {
		filters = append(filters, current.String())
	}

	return filters
}

// HasCrop 檢查是否有裁切設定
func (p *ParsedURL) HasCrop() bool {
	return p.CropLeft > 0 || p.CropTop > 0 || p.CropRight > 0 || p.CropBottom > 0
}

// HasResize 檢查是否有縮放設定
func (p *ParsedURL) HasResize() bool {
	return p.Width > 0 || p.Height > 0
}

// HasFlip 檢查是否有翻轉設定
func (p *ParsedURL) HasFlip() bool {
	return p.FlipH || p.FlipV
}

// HasFilters 檢查是否有濾鏡
func (p *ParsedURL) HasFilters() bool {
	return len(p.Filters) > 0
}
