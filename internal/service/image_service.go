package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/loader"
	"github.com/vincent119/images-filters/internal/metrics"
	"github.com/vincent119/images-filters/internal/parser"
	"github.com/vincent119/images-filters/internal/processor"
	"github.com/vincent119/images-filters/pkg/logger"
)

// imageService 圖片處理服務實作
type imageService struct {
	cfg       *config.Config
	loader    *loader.LoaderFactory
	processor *processor.Processor
	metrics   metrics.Metrics
}

// NewImageService 建立圖片處理服務
func NewImageService(cfg *config.Config, opts ...ServiceOption) ImageService {
	// 處理選項
	options := &serviceOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 建立載入器
	httpLoader := loader.NewHTTPLoader(
		loader.WithMaxSize(cfg.Server.MaxRequestSize),
	)
	fileLoader := loader.NewFileLoader(
		loader.WithRootPath(cfg.Storage.Local.RootPath),
		loader.WithFileMaxSize(cfg.Server.MaxRequestSize),
	)
	loaderFactory := loader.NewLoaderFactory(httpLoader, fileLoader)

	// 建立處理器
	proc := processor.NewProcessor(
		cfg.Processing.DefaultQuality,
		cfg.Processing.MaxWidth,
		cfg.Processing.MaxHeight,
	)

	logger.Info("image service initialized",
		logger.String("storage_root", cfg.Storage.Local.RootPath),
		logger.Int("default_quality", cfg.Processing.DefaultQuality),
		logger.Int("max_width", cfg.Processing.MaxWidth),
		logger.Int("max_height", cfg.Processing.MaxHeight),
	)

	return &imageService{
		cfg:       cfg,
		loader:    loaderFactory,
		processor: proc,
		metrics:   options.metrics,
	}
}

// ProcessImage 處理圖片
func (s *imageService) ProcessImage(ctx context.Context, parsedURL *parser.ParsedURL) ([]byte, string, error) {
	logger.Debug("start processing image",
		logger.String("image_path", parsedURL.ImagePath),
		logger.Int("width", parsedURL.Width),
		logger.Int("height", parsedURL.Height),
		logger.Bool("flip_h", parsedURL.FlipH),
		logger.Bool("flip_v", parsedURL.FlipV),
		logger.Bool("fit_in", parsedURL.FitIn),
	)

	// 1. 載入圖片
	imageData, err := s.loader.Load(ctx, parsedURL.ImagePath)
	if err != nil {
		logger.Warn("failed to load image",
			logger.String("image_path", parsedURL.ImagePath),
			logger.Err(err),
		)
		// 記錄錯誤指標
		if s.metrics != nil {
			s.metrics.RecordError("load_error")
		}
		return nil, "", fmt.Errorf("failed to load image: %w", err)
	}

	logger.Debug("image loaded successfully",
		logger.String("image_path", parsedURL.ImagePath),
		logger.Int("size_bytes", len(imageData)),
	)

	// 2. 建立處理選項
	opts := processor.ProcessOptions{
		Width:      parsedURL.Width,
		Height:     parsedURL.Height,
		FlipH:      parsedURL.FlipH,
		FlipV:      parsedURL.FlipV,
		FitIn:      parsedURL.FitIn,
		CropLeft:   parsedURL.CropLeft,
		CropTop:    parsedURL.CropTop,
		CropRight:  parsedURL.CropRight,
		CropBottom: parsedURL.CropBottom,
		Smart:      parsedURL.Smart,
		Quality:    s.cfg.Processing.DefaultQuality,
		Format:     s.determineFormat(parsedURL),
	}

	// 3. 處理圖片
	processedImage, err := s.processor.Process(imageData, opts)
	if err != nil {
		logger.Error("failed to process image",
			logger.String("image_path", parsedURL.ImagePath),
			logger.Err(err),
		)
		// 記錄錯誤指標
		if s.metrics != nil {
			s.metrics.RecordError("process_error")
		}
		return nil, "", fmt.Errorf("failed to process image: %w", err)
	}

	// 4. 編碼輸出
	outputData, err := s.processor.Encode(processedImage, opts.Format, opts.Quality)
	if err != nil {
		logger.Error("failed to encode image",
			logger.String("format", opts.Format),
			logger.Err(err),
		)
		// 記錄錯誤指標
		if s.metrics != nil {
			s.metrics.RecordError("encode_error")
		}
		return nil, "", fmt.Errorf("failed to encode image: %w", err)
	}

	// 5. 取得 Content-Type
	contentType := processor.GetContentType(opts.Format)

	// 6. 記錄圖片處理指標
	if s.metrics != nil {
		s.metrics.RecordImageProcessed(opts.Format, int64(len(outputData)))
	}

	logger.Debug("image processing completed",
		logger.String("image_path", parsedURL.ImagePath),
		logger.String("format", opts.Format),
		logger.Int("output_size", len(outputData)),
	)

	return outputData, contentType, nil
}

// determineFormat 決定輸出格式
func (s *imageService) determineFormat(parsedURL *parser.ParsedURL) string {
	// 檢查是否有 format 濾鏡
	for _, filter := range parsedURL.Filters {
		if filter.Name == "format" && len(filter.Params) > 0 {
			return normalizeFormat(filter.Params[0])
		}
	}

	// 從圖片路徑推斷格式
	ext := strings.ToLower(filepath.Ext(parsedURL.ImagePath))
	ext = strings.TrimPrefix(ext, ".")

	if isValidFormat(ext) {
		return normalizeFormat(ext)
	}

	// 使用預設格式
	return s.cfg.Processing.DefaultFormat
}

// normalizeFormat 標準化格式名稱
func normalizeFormat(format string) string {
	format = strings.ToLower(format)
	switch format {
	case "jpg":
		return "jpeg"
	case "jpeg", "png", "gif", "webp", "avif", "jxl":
		return format
	default:
		return "jpeg"
	}
}

// isValidFormat 檢查是否為有效格式
func isValidFormat(format string) bool {
	validFormats := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"gif":  true,
		"webp": true,
		"avif": true,
		"jxl":  true,
	}
	return validFormats[strings.ToLower(format)]
}
