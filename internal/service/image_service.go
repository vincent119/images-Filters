package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/vincent119/images-filters/internal/cache"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/images-filters/internal/filter"
	"github.com/vincent119/images-filters/internal/loader"
	"github.com/vincent119/images-filters/internal/metrics"
	"github.com/vincent119/images-filters/internal/parser"
	"github.com/vincent119/images-filters/internal/processor"
	"github.com/vincent119/images-filters/internal/security"
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
	cache     cache.Cache
	sem       chan struct{} // Semaphore for concurrency control
}


// UploadImage 上傳圖片並回傳簽名 URL
func (s *imageService) UploadImage(ctx context.Context, filename string, contentType string, reader io.Reader) (*UploadResult, error) {
	var inputReader io.Reader = reader

	// 0. 檢查是否啟用隱形浮水印
	if s.cfg.BlindWatermark.Enabled {
		logger.Debug("applying blind watermark", logger.String("filename", filename))

		// 必須解碼圖片才能處理
		img, err := imaging.Decode(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image for watermarking: %w", err)
		}

		// 建立濾鏡
		bwFilter := filter.NewBlindWatermarkFilter()
		bwFilter.Text = s.cfg.BlindWatermark.Text
		bwFilter.SecurityKey = s.cfg.BlindWatermark.SecurityKey

		// 應用濾鏡
		// 這裡不傳 params，使用 config 設定的預設值
		watermarkedImg, err := bwFilter.Apply(img, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to apply blind watermark: %w", err)
		}

		// 編碼回 Buffer
		buf := new(bytes.Buffer)
		format, err := imaging.FormatFromExtension(filepath.Ext(filename))
		if err != nil {
			// Fallback to JPEG if unknown
			format = imaging.JPEG
		}

		if err := imaging.Encode(buf, watermarkedImg, format); err != nil {
			return nil, fmt.Errorf("failed to encode watermarked image: %w", err)
		}

		// 使用新的 reader
		inputReader = buf
	}

	// 1. 產生儲存路徑 (uploads/{date}/{hash}_{filename})
	now := time.Now()
	datePrefix := now.Format("2006/01/02")

	// 產生唯一前綴避免檔名衝突
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", filename, now.UnixNano())))
	hashStr := hex.EncodeToString(hash[:])[:8]

	// 組合完整路徑
	savedPath := fmt.Sprintf("uploads/%s/%s_%s", datePrefix, hashStr, filepath.Base(filename))

	logger.Debug("uploading image",
		logger.String("filename", filename),
		logger.String("saved_path", savedPath),
		logger.String("content_type", contentType),
	)

	// 2. 儲存至 Storage
	if err := s.storage.PutStream(ctx, savedPath, inputReader); err != nil {
		logger.Error("failed to upload image",
			logger.String("saved_path", savedPath),
			logger.Err(err),
		)
		if s.metrics != nil {
			s.metrics.RecordError("upload_error")
		}
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	logger.Info("image uploaded successfully",
		logger.String("saved_path", savedPath),
	)

	// 3. 產生簽名 URL (使用 security.Signer)
	// 先產生用於訪問原始圖片的路徑 (不做任何處理)
	// URL 格式: /{signature}/{image_path}
	signedURL := s.generateSignedURL(savedPath)

	// 4. 記錄指標
	if s.metrics != nil {
		s.metrics.RecordStorageOperation("local", "put")
	}

	return &UploadResult{
		Path:      savedPath,
		SignedURL: signedURL,
	}, nil
}



// NewImageService 建立圖片處理服務
func NewImageService(cfg *config.Config, store storage.Storage, c cache.Cache, opts ...ServiceOption) ImageService {
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

	// Initialize semaphore
	workers := cfg.Processing.Workers
	if workers < 1 {
		workers = 1
	}

	return &imageService{
		cfg:       cfg,
		loader:    loaderFactory,
		processor: proc,
		metrics:   options.metrics,
		storage:   store,
		cache:     c,
		sem:       make(chan struct{}, workers),
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

	// 1. 檢查快取 (Cache Hit)
	resultKey := s.generateKey(parsedURL)
	cacheStart := time.Now()
	if data, err := s.cache.Get(ctx, resultKey); err == nil {
		logger.Debug("cache hit", logger.String("key", resultKey))
		if s.metrics != nil {
			s.metrics.RecordCacheHit("memory")
			s.metrics.RecordCacheLatency("get", "memory", time.Since(cacheStart).Seconds())
		}
		format := s.determineFormat(parsedURL)
		return data, processor.GetContentType(format), nil
	}
	if s.metrics != nil {
		s.metrics.RecordCacheMiss("memory")
		s.metrics.RecordCacheLatency("get", "memory", time.Since(cacheStart).Seconds())
	}

	// 2. 檢查持久化儲存 (Persistent Cache)
	storageStart := time.Now()
	if data, err := s.storage.Get(ctx, resultKey); err == nil {
		logger.Debug("storage hit", logger.String("key", resultKey))
		if s.metrics != nil {
			s.metrics.RecordStorageOperation("local", "get")
			s.metrics.RecordStorageLatency("local", "get", time.Since(storageStart).Seconds())
		}
		format := s.determineFormat(parsedURL)
		// 回寫快取 (Cache Miss but Storage Hit)
		if err := s.cache.Set(ctx, resultKey, data, 0); err != nil {
			logger.Warn("failed to set cache", logger.Err(err))
		}
		return data, processor.GetContentType(format), nil
	}
	if s.metrics != nil {
		s.metrics.RecordStorageLatency("local", "get", time.Since(storageStart).Seconds())
	}

	// 3. 限制並發處理 (Worker Pool)
	// 僅針對 Cache Miss 的請求進行限制
	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	case <-ctx.Done():
		return nil, "", ctx.Err()
	}

	// 4. 載入圖片 (Source)
	var imageReader io.ReadCloser
	var err error

	// 如果是 HTTP URL，使用 Loader
	if strings.HasPrefix(parsedURL.ImagePath, "http") {
		imageReader, err = s.loader.LoadStream(ctx, parsedURL.ImagePath)
	} else {
		// 嘗試從 Storage 讀取 (Source)
		imageReader, err = s.storage.GetStream(ctx, parsedURL.ImagePath)

		// 如果 Storage 找不到，嘗試 loader fallback
		if err != nil {
			logger.Debug("storage load stream failed, falling back to loader", logger.Err(err))
			imageReader, err = s.loader.LoadStream(ctx, parsedURL.ImagePath)
		}
	}

	if err != nil {
		logger.Warn("failed to load image stream",
			logger.String("image_path", parsedURL.ImagePath),
			logger.Err(err),
		)
		// 記錄錯誤指標
		if s.metrics != nil {
			s.metrics.RecordError("load_error")
		}
		return nil, "", fmt.Errorf("failed to load image: %w", err)
	}
	defer imageReader.Close()

	logger.Debug("image stream loaded successfully",
		logger.String("image_path", parsedURL.ImagePath),
	)

	// 5. 建立處理選項
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

	// 記錄處理操作類型
	if s.metrics != nil {
		if opts.Width > 0 || opts.Height > 0 {
			s.metrics.RecordProcessingOperation("resize")
		}
		if opts.CropLeft > 0 || opts.CropTop > 0 || opts.CropRight > 0 || opts.CropBottom > 0 {
			s.metrics.RecordProcessingOperation("crop")
		}
		if opts.FlipH {
			s.metrics.RecordProcessingOperation("flip_h")
		}
		if opts.FlipV {
			s.metrics.RecordProcessingOperation("flip_v")
		}
		if opts.Smart {
			s.metrics.RecordProcessingOperation("smart_crop")
		}
		for _, f := range parsedURL.Filters {
			s.metrics.RecordProcessingOperation(f.Name)
		}
	}

	// 6. 處理圖片（含階段耗時計量）
	decodeStart := time.Now()
	// Processor.Process now accepts io.Reader
	processedImage, err := s.processor.Process(imageReader, opts)
	if s.metrics != nil {
		s.metrics.RecordProcessingDuration("decode_transform", time.Since(decodeStart).Seconds())
	}
	if err != nil {
		logger.Error("failed to process image",
			logger.String("image_path", parsedURL.ImagePath),
			logger.Err(err),
		)
		if s.metrics != nil {
			s.metrics.RecordProcessingError("process_failed")
			s.metrics.RecordError("process_error")
		}
		return nil, "", fmt.Errorf("failed to process image: %w", err)
	}

	// 記錄輸出圖片尺寸
	if s.metrics != nil {
		bounds := processedImage.Bounds()
		s.metrics.RecordOutputImageSize(bounds.Dx(), bounds.Dy())
	}

	// 7. 編碼輸出
	encodeStart := time.Now()
	outputData, err := s.processor.Encode(processedImage, opts.Format, opts.Quality)
	if s.metrics != nil {
		s.metrics.RecordProcessingDuration("encode", time.Since(encodeStart).Seconds())
	}
	if err != nil {
		logger.Error("failed to encode image",
			logger.String("format", opts.Format),
			logger.Err(err),
		)
		if s.metrics != nil {
			s.metrics.RecordProcessingError("encode_failed")
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

	// 7. 儲存結果 (Async)
	go func() {
		// 寫入儲存
		if err := s.storage.Put(context.Background(), resultKey, outputData); err != nil {
			logger.Warn("failed to save image to storage",
				logger.String("key", resultKey),
				logger.Err(err),
			)
		}
		// 寫入快取
		if err := s.cache.Set(context.Background(), resultKey, outputData, 0); err != nil {
			logger.Warn("failed to set cache", logger.String("key", resultKey), logger.Err(err))
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

	// 從圖片路徑推斷格式（優先級 3）
	ext := strings.ToLower(filepath.Ext(parsedURL.ImagePath))
	ext = strings.TrimPrefix(ext, ".")

	// 內容協商 (Content Negotiation) (優先級 2)
	// 如果沒有強制指定 filter，且 URL 副檔名不是強制性的 (某些情況下我們希望保留原擴展名行為，
	// 但通常為了優化體驗，我們允許自動切換為更佳格式，除非使用者特意指定了 format filter)
	// 這裡策略：只要沒有顯式指定 format filter，我們就嘗試協商
	if negotiation := s.negotiateFormat(parsedURL.AcceptHeader); negotiation != "" {
		return negotiation
	}

	if isValidFormat(ext) {
		return normalizeFormat(ext)
	}

	// 使用預設格式 (優先級 4)
	return s.cfg.Processing.DefaultFormat
}

// negotiateFormat 根據 Accept 標頭協商最佳格式
func (s *imageService) negotiateFormat(acceptHeader string) string {
	if acceptHeader == "" {
		return ""
	}

	// 簡單的字串匹配，優先級 AVIF > JXL > WebP
	// 注意：這裡不解析 q-value，僅做存在性檢查，符合大多數 CDN/Server 實作
	if strings.Contains(acceptHeader, "image/avif") {
		return "avif"
	}
	if strings.Contains(acceptHeader, "image/jxl") {
		return "jxl"
	}
	if strings.Contains(acceptHeader, "image/webp") {
		return "webp"
	}

	return ""
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


// generateSignedURL 產生簽名 URL
func (s *imageService) generateSignedURL(imagePath string) string {
	// 如果安全機制未啟用，使用 unsafe 路徑
	if !s.cfg.Security.Enabled {
		return "/unsafe/" + imagePath
	}

	// 使用統一的 Signer
	signer := security.NewSigner(s.cfg.Security.SecurityKey)
	return signer.SignURL(imagePath)
}
