package filter

import (
	"image"
	"image/color"
	"testing"
)

func TestBlindWatermarkFilter_Apply(t *testing.T) {
	// 建立測試圖片
	width, height := 64, 64
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	// 填充一些顏色
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.NRGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: 100,
				A: 255,
			})
		}
	}

	tests := []struct {
		name   string
		params []string
		text   string
		key    string
	}{
		{
			name:   "Default",
			params: nil,
			text:   "Test",
			key:    "secret",
		},
		{
			name:   "With Params",
			params: []string{"Hello", "20.0"},
			text:   "",
			key:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewBlindWatermarkFilter()
			f.Text = tt.text
			f.SecurityKey = tt.key

			// 執行濾鏡
			res, err := f.Apply(img, tt.params)
			if err != nil {
				t.Fatalf("Apply failed: %v", err)
			}

			// 檢查回傳類型
			resImg, ok := res.(*image.NRGBA)
			if !ok {
				t.Fatalf("Expected NRGBA image")
			}

			// 檢查尺寸
			if resImg.Bounds() != img.Bounds() {
				t.Fatalf("Bounds changed: got %v, want %v", resImg.Bounds(), img.Bounds())
			}

			// 檢查像素是否發生變化 (浮水印應該會修改像素)
			// 注意：如果強度為 0 或圖片太小可能不變，但這裡設定應該會變
			changed := false
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					c1 := img.At(x, y).(color.NRGBA)
					c2 := resImg.At(x, y).(color.NRGBA)
					if c1 != c2 {
						changed = true
						break
					}
				}
				if changed {
					break
				}
			}

			if !changed {
				t.Errorf("Image data did not change after applying watermark")
			}
		})
	}
}

func TestDct2d(t *testing.T) {
	// 簡單測試 DCT/IDCT 可逆性 (在一定誤差範圍內)
	block := make([][]float64, 8)
	for i := 0; i < 8; i++ {
		block[i] = make([]float64, 8)
		for j := 0; j < 8; j++ {
			block[i][j] = float64(i + j)
		}
	}

	dct := dct2d(block)
	idct := idct2d(dct)

	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			diff := block[i][j] - idct[i][j]
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.0001 {
				t.Errorf("DCT/IDCT mismatch at [%d][%d]: got %f, want %f", i, j, idct[i][j], block[i][j])
			}
		}
	}
}

func TestBlindWatermarkFilter_Extract(t *testing.T) {
	// 1. 準備圖片
	width, height := 128, 128
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	// 填充隨機雜訊/圖案，避免純色導致 DCT 係數全零
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.NRGBA{
				R: uint8((x * y) % 255),
				G: uint8((x + y) % 255),
				B: uint8((x ^ y) % 255),
				A: 255,
			})
		}
	}

	// 2. 嵌入浮水印
	text := "Copyright"
	f := NewBlindWatermarkFilter()
	f.Text = text
	f.SecurityKey = "secret_key"
	f.Strength = 50.0 // 足夠強度以抗干擾

	watermarkedImg, err := f.Apply(img, nil)
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}

	// 3. 提取浮水印
	// 預期提取同樣長度的文字
	extracted, err := f.Extract(watermarkedImg, len(text))
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// 4. 驗證
	t.Logf("Original: %s", text)
	t.Logf("Extracted: %s", extracted)

	// 注意：DCT 隱寫可能會有誤碼，這裡檢查是否完全匹配
	// 如果有誤碼，可以改用相似度檢查
	if extracted != text {
		t.Errorf("Extracted text mismatch: got %q, want %q", extracted, text)
	}
}
