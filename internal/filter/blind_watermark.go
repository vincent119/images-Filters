package filter

import (
	"image"
	"image/color"
	"math"

	"strconv"

	"github.com/disintegration/imaging"
)

// BlindWatermarkFilter 隱形浮水印濾鏡
type BlindWatermarkFilter struct {
	Text        string
	SecurityKey string
	Strength    float64
}

// NewBlindWatermarkFilter 建立隱形浮水印濾鏡
func NewBlindWatermarkFilter() *BlindWatermarkFilter {
	return &BlindWatermarkFilter{
		Strength: 10.0, // 預設強度
	}
}

// Name 回傳濾鏡名稱
func (f *BlindWatermarkFilter) Name() string {
	return "blind_watermark"
}

// Apply 套用濾鏡
// params[0]: text (浮水印文字)
// params[1]: strength (強度, optiona, default 10.0)
func (f *BlindWatermarkFilter) Apply(img image.Image, params []string) (image.Image, error) {
	// 1. 解析參數
	text := f.Text
	strength := f.Strength
	key := f.SecurityKey

	if len(params) > 0 && params[0] != "" {
		text = params[0]
	}
	if len(params) > 1 {
		if s, err := strconv.ParseFloat(params[1], 64); err == nil && s > 0 {
			strength = s
		}
	}

	// 如果沒有文字，不處理
	if text == "" {
		return img, nil
	}

	// 1. 轉換為 YCbCr 或灰階 (為了方便處理，這裡先假設處理亮度通道 Y)
	// 簡單起見，我們先將浮水印嵌入到 Resize 後的圖片
	// 注意：DCT 需要圖片尺寸是 8 的倍數，或者我們會對 8x8 區塊進行處理

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 建立可修改的圖片副本 (RGBA)
	rgba := imaging.Clone(img)

	// 產生浮水印訊號 (基於 Text 和 SecurityKey)
	watermarkSignal := generateWatermarkSignal(text, key)

	// 嵌入浮水印
	embedWatermark(rgba, width, height, watermarkSignal, strength)

	return rgba, nil
}

// generateWatermarkSignal 產生浮水印訊號
// 這裡將文字轉換為二進位序列，並重複填充以適應圖片
func generateWatermarkSignal(text, key string) []int {
	// 簡單實作：將文字轉為 ascii bits
	var bits []int
	for _, char := range text {
		for i := 0; i < 8; i++ {
			if (char>>uint(7-i))&1 == 1 {
				bits = append(bits, 1)
			} else {
				bits = append(bits, -1) // 使用 1 和 -1
			}
		}
	}
	if len(bits) == 0 {
		return []int{1, -1, 1, -1} // Default pattern
	}
	return bits
}

// embedWatermark 嵌入浮水印
func embedWatermark(img *image.NRGBA, width, height int, signal []int, strength float64) {
	// 對每個 8x8 區塊進行 DCT
	// 修改中頻係數
	// IDCT

	// 這裡僅提供框架，詳細 DCT 運算較為複雜
	// 為此示範，我們將實作一個簡化的空間域 LSB 隱寫作為 placeholder，
	// 因為完整的 DCT 實作程式碼量較大，且需引入矩陣運算庫或手寫 DCT。
	// 但為了符合 "Blind Watermark" 要求，我會盡量寫一個簡易的 DCT 變換。

	processBlocks(img, width, height, signal, strength)
}

// processBlocks 處理 8x8 區塊
func processBlocks(img *image.NRGBA, width, height int, signal []int, strength float64) {
	signalIdx := 0
	signalLen := len(signal)

	for y := 0; y <= height-8; y += 8 {
		for x := 0; x <= width-8; x += 8 {
			// 提取 Y 通道 8x8
			block := extractYBlock(img, x, y)

			// DCT
			dctBlock := dct2d(block)

			// 嵌入 bit
			bit := signal[signalIdx%signalLen]
			embedInDct(dctBlock, bit, strength)

			// IDCT
			idctBlock := idct2d(dctBlock)

			// 寫回圖片
			updateYBlock(img, x, y, idctBlock)

			signalIdx++
		}
	}
}

// extractYBlock 提取 8x8 Y 通道
func extractYBlock(img *image.NRGBA, xStart, yStart int) [][]float64 {
	block := make([][]float64, 8)
	for i := 0; i < 8; i++ {
		block[i] = make([]float64, 8)
		for j := 0; j < 8; j++ {
			c := img.At(xStart+j, yStart+i)
			r, g, b, _ := c.RGBA()
			// RGB to Y (SDTV formula)
			y := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
			block[i][j] = y - 128 // Shift to zero-mean
		}
	}
	return block
}

// updateYBlock 更新 8x8 Y 通道
func updateYBlock(img *image.NRGBA, xStart, yStart int, block [][]float64) {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			// Get original color to preserve Cb Cr
			c := img.At(xStart+j, yStart+i)
			r, g, b, a := c.RGBA()
			r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8)

			// yOld := 0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8)
			cb := -0.1687*float64(r8) - 0.3313*float64(g8) + 0.5*float64(b8)
			cr := 0.5*float64(r8) - 0.4187*float64(g8) - 0.0813*float64(b8)

			yNew := block[i][j] + 128

			// YCbCr back to RGB
			newR := yNew + 1.402*cr
			newG := yNew - 0.34414*cb - 0.71414*cr
			newB := yNew + 1.772*cb

			img.Set(xStart+j, yStart+i, color.NRGBA{
				R: clip(newR),
				G: clip(newG),
				B: clip(newB),
				A: a8,
			})
		}
	}
}

func clip(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// dct2d 2D DCT
func dct2d(block [][]float64) [][]float64 {
	// 簡單實作，可優化
	res := make([][]float64, 8)
	for i := 0; i < 8; i++ {
		res[i] = make([]float64, 8)
	}

	c1 := 0.0
	c2 := 0.0

	for u := 0; u < 8; u++ {
		for v := 0; v < 8; v++ {
			if u == 0 {
				c1 = 1.0 / math.Sqrt(2)
			} else {
				c1 = 1.0
			}
			if v == 0 {
				c2 = 1.0 / math.Sqrt(2)
			} else {
				c2 = 1.0
			}

			sum := 0.0
			for x := 0; x < 8; x++ {
				for y := 0; y < 8; y++ {
					sum += block[y][x] *
						math.Cos((2*float64(x)+1)*float64(u)*math.Pi/16) *
						math.Cos((2*float64(y)+1)*float64(v)*math.Pi/16)
				}
			}
			res[v][u] = 0.25 * c1 * c2 * sum
		}
	}
	return res
}

// idct2d 2D IDCT
func idct2d(block [][]float64) [][]float64 {
	res := make([][]float64, 8)
	for i := 0; i < 8; i++ {
		res[i] = make([]float64, 8)
	}

	c1 := 0.0
	c2 := 0.0

	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			sum := 0.0
			for u := 0; u < 8; u++ {
				for v := 0; v < 8; v++ {
					if u == 0 {
						c1 = 1.0 / math.Sqrt(2)
					} else {
						c1 = 1.0
					}
					if v == 0 {
						c2 = 1.0 / math.Sqrt(2)
					} else {
						c2 = 1.0
					}

					sum += c1 * c2 * block[v][u] *
						math.Cos((2*float64(x)+1)*float64(u)*math.Pi/16) *
						math.Cos((2*float64(y)+1)*float64(v)*math.Pi/16)
				}
			}
			res[y][x] = 0.25 * sum
		}
	}
	return res
}

// embedInDct 修改中頻係數 (Cox 演算法變體)
func embedInDct(block [][]float64, bit int, strength float64) {
	// 選取一對中頻係數，例如 (4,3) 和 (3,4)
	v1 := block[3][4]
	v2 := block[4][3]

	if bit == 1 {
		// 使 v1 > v2 + strength
		if v1 <= v2+strength {
			diff := (v2 + strength - v1) / 2
			v1 += diff + 0.1
			v2 -= diff + 0.1 // 確保變動在兩者之間
		}
	} else {
		// 使 v1 < v2 - strength
		if v1 >= v2-strength {
			diff := (v1 - (v2 - strength)) / 2
			v1 -= diff + 0.1
			v2 += diff + 0.1
		}
	}
	block[3][4] = v1
	block[4][3] = v2
}

// Extract 提取隱形浮水印
// length: 預期文字長度 (字元數)
func (f *BlindWatermarkFilter) Extract(img image.Image, length int) (string, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 如果沒有指定長度，預設嘗試提取一些 (比如 16 chars)
	if length <= 0 {
		length = 16
	}

	bitLength := length * 8
	votes := make([]float64, bitLength)
	counts := make([]int, bitLength)

	// 複製圖片以確保不修改原圖 (雖然提取不需要修改，但為了讀取像素一致性)
	// 這裡直接讀取
	nrgba, ok := img.(*image.NRGBA)
	if !ok {
		nrgba = imaging.Clone(img)
	}

	signalIdx := 0

	for y := 0; y <= height-8; y += 8 {
		for x := 0; x <= width-8; x += 8 {
			// 提取 Y 通道
			block := extractYBlock(nrgba, x, y)

			// DCT
			dctBlock := dct2d(block)

			// 提取 bit
			val, diff := extractFromDct(dctBlock)

			// 投票機制: 累加差異值 (正值代表 1, 負值代表 0/-1)
			// 使用 diff 作為權重，信心度越高權重越大
			voteIdx := signalIdx % bitLength
			votes[voteIdx] += val * diff
			counts[voteIdx]++

			signalIdx++
		}
	}

	// 重組 bits
	var bits []int
	for _, v := range votes {
		if v > 0 {
			bits = append(bits, 1)
		} else {
			bits = append(bits, 0)
		}
	}

	return bitsToString(bits), nil
}

// extractFromDct 從 DCT 係數提取 bit
// 回傳: (bit value 1/-1, confidence/diff)
func extractFromDct(block [][]float64) (float64, float64) {
	v1 := block[3][4]
	v2 := block[4][3]

	if v1 >= v2 {
		return 1.0, math.Abs(v1 - v2)
	}
	return -1.0, math.Abs(v1 - v2)
}

// bitsToString 將 bits 轉換回字串
func bitsToString(bits []int) string {
	var bytes []byte
	for i := 0; i < len(bits); i += 8 {
		var val byte
		for j := 0; j < 8; j++ {
			if i+j < len(bits) && bits[i+j] == 1 {
				val |= 1 << uint(7-j)
			}
		}
		bytes = append(bytes, val)
	}
	// 去除結尾的 null chars (如果有)
	result := string(bytes)
	// result = strings.TrimRight(result, "\x00")
	return result
}
