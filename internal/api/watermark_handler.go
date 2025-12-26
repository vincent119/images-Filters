package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vincent119/images-filters/internal/service"
)

// WatermarkHandler 浮水印處理器
type WatermarkHandler struct {
	service service.WatermarkService
}

// NewWatermarkHandler 建立新的浮水印處理器
func NewWatermarkHandler(svc service.WatermarkService) *WatermarkHandler {
	return &WatermarkHandler{
		service: svc,
	}
}

// HandleDetect 處理浮水印檢測請求
// @Summary Detect watermark
// @Description Detect blind watermark in image. Either upload a file or provide a storage path.
// @Tags Watermark
// @Accept multipart/form-data
// @Produce json
// @Param file formData file false "Image file to detect (optional if path is provided)"
// @Param path formData string false "Storage path of existing image (e.g. uploads/2025/12/26/image.jpg)"
// @Success 200 {object} service.DetectionResult
// @Failure 400 {object} ErrorResponse "Bad request - neither file nor path provided"
// @Failure 500 {object} ErrorResponse "Internal error"
// @Router /detect [post]
func (h *WatermarkHandler) HandleDetect(c *gin.Context) {
	// 1. 嘗試取得檔案
	file, _, err := c.Request.FormFile("file")
	if err == nil {
		defer file.Close()
		// 有檔案，正常檢測
		result, err := h.service.DetectWatermark(c.Request.Context(), file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "DETECTION_ERROR",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	// 2. 如果沒有檔案，檢查 path 參數
	path := c.PostForm("path")
	if path != "" {
		result, err := h.service.DetectWatermarkFromPath(c.Request.Context(), path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "DETECTION_ERROR",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	// 3. 都沒有
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   "INVALID_REQUEST",
		Message: "Either 'file' or 'path' must be provided",
	})
}
