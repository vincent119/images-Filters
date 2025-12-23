package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/loader"
	"github.com/vincent119/images-filters/internal/metrics"
	"github.com/vincent119/images-filters/internal/parser"
	"github.com/vincent119/images-filters/internal/processor"
	"github.com/vincent119/images-filters/internal/storage"
	"github.com/vincent119/images-filters/pkg/logger"
)

// imageService 圖片處理服務實作
type imageService struct {
	cfg       *config.Config
	loader    *loader.LoaderFactory
	processor *processor.Processor
	metrics   metrics.Metrics
	storage   storage.Storage
}

// NewImageService 建立圖片處理服務
func NewImageService(cfg *config.Config, store storage.Storage, opts ...ServiceOption) ImageService {
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
		logger.String("storage_type", cfg.Storage.Type),
		logger.Int("default_quality", cfg.Processing.DefaultQuality),
		logger.Int("max_width", cfg.Processing.MaxWidth),
		logger.Int("max_height", cfg.Processing.MaxHeight),
	)

	return &imageService{
		cfg:       cfg,
		loader:    loaderFactory,
		processor: proc,
		metrics:   options.metrics,
		storage:   store,
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

	// 1. 檢查儲存是否存在 (Cache Hit)
	resultKey := s.generateKey(parsedURL)
	if data, err := s.storage.Get(ctx, resultKey); err == nil {
		logger.Debug("cache hit", logger.String("key", resultKey))

		// 判斷 Content-Type
		// 這裡假設我們只存圖片，且可以從 key 或內容推斷
		// 為簡化，我們重新使用 determineFormat 推斷 ContentType，或者儲存 metadata
		// 簡單起見，我們再次依賴 determineFormat (雖然可能不精確如果 format 被改了)
		// 更好的方式是儲存 metadata，但在 key 中包含 format 即可
		format := s.determineFormat(parsedURL)
		return data, processor.GetContentType(format), nil
	}

	// 2. 載入圖片 (Source)
	var imageData []byte
	var err error

	// 如果是 HTTP URL，使用 Loader
	if strings.HasPrefix(parsedURL.ImagePath, "http") {
		imageData, err = s.loader.Load(ctx, parsedURL.ImagePath)
	} else {
		// 嘗試從 Storage 讀取 (Source)
		// 注意: MixedStorage.Get 會先查 Result，再查 Source
		// 這裡我們直接查原始路徑
		imageData, err = s.storage.Get(ctx, parsedURL.ImagePath)

		// 如果 Storage 找不到 (例如 LocalStorage 沒設好或其實在 Loader 路徑)
		// fallback 到 file loader (如果 configured root path matches)
		// 但這會變複雜。簡單起見，若 storage 失敗則嘗試 loader
		if err != nil {
			logger.Debug("storage load failed, falling back to loader", logger.Err(err))
			imageData, err = s.loader.Load(ctx, parsedURL.ImagePath)
		}
	}

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

	// 7. 儲存結果
	// 非同步寫入儲存，避免阻塞回應
	go func() {
		if err := s.storage.Put(context.Background(), resultKey, outputData); err != nil {
			logger.Warn("failed to save image to storage",
				logger.String("key", resultKey),
				logger.Err(err),
			)
		}
	}()

	logger.Debug("image processing completed",
		logger.String("image_path", parsedURL.ImagePath),
		logger.String("result_key", resultKey),
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

// generateKey 產生快取鍵值
func (s *imageService) generateKey(p *parser.ParsedURL) string {
	// 基礎鍵值：路徑
	base := p.ImagePath

	// 參數簽名
	// 格式: w{width}_h{height}_f{format}_q{quality}_...

	// 處理參數
	params := []string{
		fmt.Sprintf("w%d", p.Width),
		fmt.Sprintf("h%d", p.Height),
		fmt.Sprintf("fh%v", p.FlipH),
		fmt.Sprintf("fv%v", p.FlipV),
		fmt.Sprintf("fit%v", p.FitIn),
		fmt.Sprintf("sm%v", p.Smart),
		fmt.Sprintf("c%d_%d_%d_%d", p.CropLeft, p.CropTop, p.CropRight, p.CropBottom),
	}

	// 濾鏡
	for _, f := range p.Filters {
		params = append(params, fmt.Sprintf("%s(%s)", f.Name, strings.Join(f.Params, ",")))
	}

	// 格式與品質
	format := s.determineFormat(p)
	params = append(params, fmt.Sprintf("fmt_%s", format))
	params = append(params, fmt.Sprintf("q%d", s.cfg.Processing.DefaultQuality))

	// 組合
	paramStr := strings.Join(params, "-")

	// 雜湊處理參數部分以縮短長度
	hash := sha256.Sum256([]byte(paramStr))
	hashStr := hex.EncodeToString(hash[:])[:16]

	// 加上副檔名
	ext := filepath.Ext(base)
	if ext == "" {
		ext = fmt.Sprintf(".%s", format)
	}

	// 結果：cache/hash/filename
	// 為了避免單一目錄過大，可以使用 hash 前綴分層
	return fmt.Sprintf("cache/%s/%s/%s", hashStr[:2], hashStr[2:], filepath.Base(base))
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
