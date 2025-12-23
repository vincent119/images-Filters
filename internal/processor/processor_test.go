package processor

import (
	"testing"
)

func TestCalculateDimensions(t *testing.T) {
	tests := []struct {
		name           string
		originalWidth  int
		originalHeight int
		targetWidth    int
		targetHeight   int
		fitIn          bool
		wantWidth      int
		wantHeight     int
	}{
		{
			name:           "原尺寸（兩個都是 0）",
			originalWidth:  800,
			originalHeight: 600,
			targetWidth:    0,
			targetHeight:   0,
			fitIn:          false,
			wantWidth:      800,
			wantHeight:     600,
		},
		{
			name:           "只指定寬度",
			originalWidth:  800,
			originalHeight: 600,
			targetWidth:    400,
			targetHeight:   0,
			fitIn:          false,
			wantWidth:      400,
			wantHeight:     300,
		},
		{
			name:           "只指定高度",
			originalWidth:  800,
			originalHeight: 600,
			targetWidth:    0,
			targetHeight:   300,
			fitIn:          false,
			wantWidth:      400,
			wantHeight:     300,
		},
		{
			name:           "指定固定尺寸（填滿模式）",
			originalWidth:  800,
			originalHeight: 600,
			targetWidth:    300,
			targetHeight:   200,
			fitIn:          false,
			wantWidth:      300,
			wantHeight:     200,
		},
		{
			name:           "Fit-in 較寬圖片",
			originalWidth:  800,
			originalHeight: 400,
			targetWidth:    300,
			targetHeight:   300,
			fitIn:          true,
			wantWidth:      300,
			wantHeight:     150,
		},
		{
			name:           "Fit-in 較高圖片",
			originalWidth:  400,
			originalHeight: 800,
			targetWidth:    300,
			targetHeight:   300,
			fitIn:          true,
			wantWidth:      150,
			wantHeight:     300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := calculateDimensions(
				tt.originalWidth, tt.originalHeight,
				tt.targetWidth, tt.targetHeight,
				tt.fitIn,
			)

			if w != tt.wantWidth {
				t.Errorf("width = %d; want %d", w, tt.wantWidth)
			}
			if h != tt.wantHeight {
				t.Errorf("height = %d; want %d", h, tt.wantHeight)
			}
		})
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{"jpeg", "image/jpeg"},
		{"jpg", "image/jpeg"},
		{"png", "image/png"},
		{"gif", "image/gif"},
		{"webp", "image/webp"},
		{"avif", "image/avif"},
		{"jxl", "image/jxl"},
		{"unknown", "image/jpeg"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			if got := GetContentType(tt.format); got != tt.want {
				t.Errorf("GetContentType(%s) = %s; want %s", tt.format, got, tt.want)
			}
		})
	}
}
