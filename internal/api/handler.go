// Package api 提供 HTTP API 處理器
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vincent119/images-filters/internal/parser"
	"github.com/vincent119/images-filters/internal/service"
)

// Handler HTTP 處理器
type Handler struct {
	imageService service.ImageService
	urlParser    *parser.URLParser
}

// NewHandler 建立新的處理器
func NewHandler(imageService service.ImageService) *Handler {
	return &Handler{
		imageService: imageService,
		urlParser:    parser.NewURLParser(),
	}
}

// HandleImage 處理圖片請求
// @Summary Process image
// @Description Process image based on URL parameters (Resize, Crop, Flip, Filters)
// @Tags Image
// @Produce octet-stream
// @Param path path string true "Image processing path"
// @Success 200 {file} binary "Processed image"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 404 {object} ErrorResponse "Image not found"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /{path} [get]
func (h *Handler) HandleImage(c *gin.Context) {
	// 取得完整路徑（使用 Request.URL.Path）
	path := c.Request.URL.Path

	// 解析 URL
	parsedURL, err := h.urlParser.Parse(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_URL",
			Message: err.Error(),
		})
		return
	}

	// 設定 Accept Header 用於內容協商
	parsedURL.AcceptHeader = c.Request.Header.Get("Accept")

	// 處理圖片
	imageData, contentType, err := h.imageService.ProcessImage(c.Request.Context(), parsedURL)
	if err != nil {
		// 根據錯誤類型返回不同的狀態碼
		statusCode := http.StatusInternalServerError
		errorCode := "PROCESSING_ERROR"

		if isNotFoundError(err) {
			statusCode = http.StatusNotFound
			errorCode = "IMAGE_NOT_FOUND"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorCode,
			Message: err.Error(),
		})
		return
	}

	// 設定快取標頭
	c.Header("Cache-Control", "public, max-age=31536000")

	// 返回圖片
	c.Data(http.StatusOK, contentType, imageData)
}

// HealthCheck health check endpoint
// @Summary Health check
// @Description Check if service is running properly
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /healthz [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, HealthResponse{
		Status: "ok",
	})
}

// ErrorResponse 錯誤回應
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// HealthResponse 健康檢查回應
type HealthResponse struct {
	Status string `json:"status"`
}

// isNotFoundError 檢查是否為找不到檔案的錯誤
func isNotFoundError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "不存在") ||
		contains(errStr, "not found") ||
		contains(errStr, "404")
}

// contains 檢查字串是否包含子字串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// UploadResponse 上傳回應
type UploadResponse struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

// HandleUpload 處理圖片上傳
// @Summary Upload image
// @Description Upload an image file to storage
// @Tags Image
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file to upload"
// @Success 200 {object} UploadResponse
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Security BearerAuth
// @Router /upload [post]
func (h *Handler) HandleUpload(c *gin.Context) {
	// 取得上傳的檔案
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_FILE",
			Message: "Failed to read uploaded file: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// 取得檔案資訊
	filename := header.Filename
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 驗證檔案類型
	if !isValidImageType(contentType) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_FILE_TYPE",
			Message: "Only image files are allowed",
		})
		return
	}

	// 呼叫 Service 進行上傳
	result, err := h.imageService.UploadImage(c.Request.Context(), filename, contentType, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "UPLOAD_ERROR",
			Message: err.Error(),
		})
		return
	}

	// 回傳結果
	c.JSON(http.StatusOK, UploadResponse{
		Path: result.Path,
		URL:  result.SignedURL,
	})
}

// isValidImageType 檢查是否為有效的圖片類型
func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg":    true,
		"image/png":     true,
		"image/gif":     true,
		"image/webp":    true,
		"image/avif":    true,
		"image/jxl":     true,
		"image/svg+xml": true,
		"image/heic":    true,
		"image/heif":    true,
		"image/bmp":     true,
		"image/tiff":    true,
	}
	return validTypes[contentType]
}
