package parser

import (
	"testing"
)

func TestURLParser_Parse(t *testing.T) {
	parser := NewURLParser()

	tests := []struct {
		name      string
		path      string
		want      *ParsedURL
		wantError bool
	}{
		{
			name: "基本 unsafe + 尺寸",
			path: "/unsafe/300x200/https://example.com/image.jpg",
			want: &ParsedURL{
				IsUnsafe:  true,
				Width:     300,
				Height:    200,
				ImagePath: "https://example.com/image.jpg",
			},
		},
		{
			name: "簽名模式",
			path: "/K97LekICOXT9MbO3X1u8BBkrjbu5/300x200/image.jpg",
			want: &ParsedURL{
				Signature: "K97LekICOXT9MbO3X1u8BBkrjbu5",
				Width:     300,
				Height:    200,
				ImagePath: "image.jpg",
			},
		},
		{
			name: "水平翻轉",
			path: "/unsafe/-300x200/image.jpg",
			want: &ParsedURL{
				IsUnsafe:  true,
				Width:     300,
				Height:    200,
				FlipH:     true,
				ImagePath: "image.jpg",
			},
		},
		{
			name: "垂直翻轉",
			path: "/unsafe/300x-200/image.jpg",
			want: &ParsedURL{
				IsUnsafe:  true,
				Width:     300,
				Height:    200,
				FlipV:     true,
				ImagePath: "image.jpg",
			},
		},
		{
			name: "雙向翻轉",
			path: "/unsafe/-300x-200/image.jpg",
			want: &ParsedURL{
				IsUnsafe:  true,
				Width:     300,
				Height:    200,
				FlipH:     true,
				FlipV:     true,
				ImagePath: "image.jpg",
			},
		},
		{
			name: "只指定寬度",
			path: "/unsafe/300x0/image.jpg",
			want: &ParsedURL{
				IsUnsafe:  true,
				Width:     300,
				Height:    0,
				ImagePath: "image.jpg",
			},
		},
		{
			name: "只指定高度",
			path: "/unsafe/0x200/image.jpg",
			want: &ParsedURL{
				IsUnsafe:  true,
				Width:     0,
				Height:    200,
				ImagePath: "image.jpg",
			},
		},
		{
			name: "fit-in 模式",
			path: "/unsafe/fit-in/300x200/image.jpg",
			want: &ParsedURL{
				IsUnsafe:  true,
				FitIn:     true,
				Width:     300,
				Height:    200,
				ImagePath: "image.jpg",
			},
		},
		{
			name: "smart 裁切",
			path: "/unsafe/300x200/smart/image.jpg",
			want: &ParsedURL{
				IsUnsafe:  true,
				Width:     300,
				Height:    200,
				Smart:     true,
				ImagePath: "image.jpg",
			},
		},
		{
			name: "手動裁切",
			path: "/unsafe/10x20:100x150/image.jpg",
			want: &ParsedURL{
				IsUnsafe:   true,
				CropLeft:   10,
				CropTop:    20,
				CropRight:  100,
				CropBottom: 150,
				ImagePath:  "image.jpg",
			},
		},
		{
			name: "單個濾鏡",
			path: "/unsafe/300x200/filters:blur(7)/image.jpg",
			want: &ParsedURL{
				IsUnsafe:  true,
				Width:     300,
				Height:    200,
				ImagePath: "image.jpg",
				Filters: []Filter{
					{Name: "blur", Params: []string{"7"}},
				},
			},
		},
		{
			name: "多個濾鏡",
			path: "/unsafe/300x200/filters:blur(7):grayscale()/image.jpg",
			want: &ParsedURL{
				IsUnsafe:  true,
				Width:     300,
				Height:    200,
				ImagePath: "image.jpg",
				Filters: []Filter{
					{Name: "blur", Params: []string{"7"}},
					{Name: "grayscale", Params: nil},
				},
			},
		},
		{
			name: "多參數濾鏡",
			path: "/unsafe/300x200/filters:watermark(logo.png,10,20,50)/image.jpg",
			want: &ParsedURL{
				IsUnsafe:  true,
				Width:     300,
				Height:    200,
				ImagePath: "image.jpg",
				Filters: []Filter{
					{Name: "watermark", Params: []string{"logo.png", "10", "20", "50"}},
				},
			},
		},
		{
			name: "完整複雜路徑",
			path: "/unsafe/fit-in/-300x-200/smart/filters:blur(5):grayscale()/https://example.com/path/to/image.jpg",
			want: &ParsedURL{
				IsUnsafe:  true,
				FitIn:     true,
				Width:     300,
				Height:    200,
				FlipH:     true,
				FlipV:     true,
				Smart:     true,
				ImagePath: "https://example.com/path/to/image.jpg",
				Filters: []Filter{
					{Name: "blur", Params: []string{"5"}},
					{Name: "grayscale", Params: nil},
				},
			},
		},
		{
			name:      "空路徑",
			path:      "",
			wantError: true,
		},
		{
			name:      "只有斜線",
			path:      "/",
			wantError: true,
		},
		{
			name:      "缺少圖片路徑",
			path:      "/unsafe/300x200",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.path)

			if tt.wantError {
				if err == nil {
					t.Error("期望錯誤但沒有發生")
				}
				return
			}

			if err != nil {
				t.Fatalf("解析錯誤: %v", err)
			}

			// 比較結果
			if result.IsUnsafe != tt.want.IsUnsafe {
				t.Errorf("IsUnsafe = %v; want %v", result.IsUnsafe, tt.want.IsUnsafe)
			}
			if result.Signature != tt.want.Signature {
				t.Errorf("Signature = %s; want %s", result.Signature, tt.want.Signature)
			}
			if result.Width != tt.want.Width {
				t.Errorf("Width = %d; want %d", result.Width, tt.want.Width)
			}
			if result.Height != tt.want.Height {
				t.Errorf("Height = %d; want %d", result.Height, tt.want.Height)
			}
			if result.FlipH != tt.want.FlipH {
				t.Errorf("FlipH = %v; want %v", result.FlipH, tt.want.FlipH)
			}
			if result.FlipV != tt.want.FlipV {
				t.Errorf("FlipV = %v; want %v", result.FlipV, tt.want.FlipV)
			}
			if result.FitIn != tt.want.FitIn {
				t.Errorf("FitIn = %v; want %v", result.FitIn, tt.want.FitIn)
			}
			if result.Smart != tt.want.Smart {
				t.Errorf("Smart = %v; want %v", result.Smart, tt.want.Smart)
			}
			if result.CropLeft != tt.want.CropLeft {
				t.Errorf("CropLeft = %d; want %d", result.CropLeft, tt.want.CropLeft)
			}
			if result.CropTop != tt.want.CropTop {
				t.Errorf("CropTop = %d; want %d", result.CropTop, tt.want.CropTop)
			}
			if result.CropRight != tt.want.CropRight {
				t.Errorf("CropRight = %d; want %d", result.CropRight, tt.want.CropRight)
			}
			if result.CropBottom != tt.want.CropBottom {
				t.Errorf("CropBottom = %d; want %d", result.CropBottom, tt.want.CropBottom)
			}
			if result.ImagePath != tt.want.ImagePath {
				t.Errorf("ImagePath = %s; want %s", result.ImagePath, tt.want.ImagePath)
			}

			// 比較濾鏡
			if len(result.Filters) != len(tt.want.Filters) {
				t.Errorf("Filters count = %d; want %d", len(result.Filters), len(tt.want.Filters))
			} else {
				for i, filter := range result.Filters {
					if filter.Name != tt.want.Filters[i].Name {
						t.Errorf("Filter[%d].Name = %s; want %s", i, filter.Name, tt.want.Filters[i].Name)
					}
					if len(filter.Params) != len(tt.want.Filters[i].Params) {
						t.Errorf("Filter[%d].Params count = %d; want %d", i, len(filter.Params), len(tt.want.Filters[i].Params))
					}
				}
			}
		})
	}
}

func TestParsedURL_HasMethods(t *testing.T) {
	t.Run("HasCrop", func(t *testing.T) {
		p := &ParsedURL{CropLeft: 10}
		if !p.HasCrop() {
			t.Error("HasCrop() = false; want true")
		}

		p = &ParsedURL{}
		if p.HasCrop() {
			t.Error("HasCrop() = true; want false")
		}
	})

	t.Run("HasResize", func(t *testing.T) {
		p := &ParsedURL{Width: 300}
		if !p.HasResize() {
			t.Error("HasResize() = false; want true")
		}

		p = &ParsedURL{}
		if p.HasResize() {
			t.Error("HasResize() = true; want false")
		}
	})

	t.Run("HasFlip", func(t *testing.T) {
		p := &ParsedURL{FlipH: true}
		if !p.HasFlip() {
			t.Error("HasFlip() = false; want true")
		}

		p = &ParsedURL{}
		if p.HasFlip() {
			t.Error("HasFlip() = true; want false")
		}
	})

	t.Run("HasFilters", func(t *testing.T) {
		p := &ParsedURL{Filters: []Filter{{Name: "blur"}}}
		if !p.HasFilters() {
			t.Error("HasFilters() = false; want true")
		}

		p = &ParsedURL{}
		if p.HasFilters() {
			t.Error("HasFilters() = true; want false")
		}
	})
}
