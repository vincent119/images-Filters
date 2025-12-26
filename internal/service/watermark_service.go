package service

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/filter"
	"github.com/vincent119/images-filters/internal/storage"
	"github.com/vincent119/images-filters/pkg/logger"
)

// DetectionResult 浮水印檢測結果
type DetectionResult struct {
	Detected   bool    `json:"detected"`
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
}

// WatermarkService 浮水印服務介面
type WatermarkService interface {
	DetectWatermark(ctx context.Context, file io.Reader) (*DetectionResult, error)
	DetectWatermarkFromPath(ctx context.Context, path string) (*DetectionResult, error)
}

// watermarkService 實作
type watermarkService struct {
	cfg     *config.Config
	storage storage.Storage
}

// NewWatermarkService 建立新的浮水印服務
func NewWatermarkService(cfg *config.Config, s storage.Storage) WatermarkService {
	return &watermarkService{
		cfg:     cfg,
		storage: s,
	}
}

// DetectWatermarkFromPath 從路徑檢測浮水印
func (s *watermarkService) DetectWatermarkFromPath(ctx context.Context, path string) (*DetectionResult, error) {
	// 讀取檔案
	// Storage 介面通常是 Get(ctx, key) ([]byte, error)
	data, err := s.storage.Get(ctx, path)
	if err != nil {
		logger.Error("failed to get image from storage", logger.String("path", path), logger.Err(err))
		return nil, err
	}

	return s.DetectWatermark(ctx, bytes.NewReader(data))
}

// DetectWatermark 檢測圖片中的浮水印
func (s *watermarkService) DetectWatermark(ctx context.Context, file io.Reader) (*DetectionResult, error) {
	// 1. 解碼圖片
	img, err := imaging.Decode(file)
	if err != nil {
		logger.Error("failed to decode image for detection", logger.Err(err))
		return nil, err
	}

	// 2. 準備濾鏡與參數
	bwFilter := filter.NewBlindWatermarkFilter()
	expectedText := s.cfg.BlindWatermark.Text
	// 如果配置中沒有設定文字，預設使用一個合理的長度嘗試提取，或者直接報錯？
	// 這裡假設我們想檢測"任何"浮水印，但我們需要長度。
	// 如果配置有文字，就用該長度。否則預設 16。
	length := len(expectedText)

	// 3. 提取
	extractedText, err := bwFilter.Extract(img, length)
	if err != nil {
		return nil, err
	}

	// 4. 分析結果
	detected := false
	confidence := 0.0

	// 簡單移除空字符
	extractedText = strings.Trim(extractedText, "\x00")

	if expectedText != "" {
		// 計算相似度 (Levenshtein Distance 簡易版或直接比較)
		// 這裡做一個簡單的包含檢查或相等檢查
		if extractedText == expectedText {
			detected = true
			confidence = 1.0
		} else {
			// 如果部分匹配（例如雜訊導致幾個字元錯了）
			// 計算匹配率
			matchCount := 0
			minLen := len(expectedText)
			if len(extractedText) < minLen {
				minLen = len(extractedText)
			}
			for i := 0; i < minLen; i++ {
				if extractedText[i] == expectedText[i] {
					matchCount++
				}
			}
			if minLen > 0 {
				confidence = float64(matchCount) / float64(len(expectedText))
			}
			// 設定一個閾值，例如 70% 相似度視為檢測到
			if confidence >= 0.7 {
				detected = true
			}
		}
	} else {
		// 如果沒有預期文字，只要提取出的文字看起來是 ASCII 可讀的，就當作檢測到？
		// 這裡簡單回傳提取結果
		detected = true
		confidence = 0.5 // 不確定
	}

	return &DetectionResult{
		Detected:   detected,
		Text:       extractedText,
		Confidence: confidence,
	}, nil
}
