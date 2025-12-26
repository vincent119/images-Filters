package service

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"testing"

	"github.com/vincent119/images-filters/internal/config"
)

// mockStorage 模擬儲存（專供 watermark_service 測試使用）
type mockStorage struct {
	data map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		data: make(map[string][]byte),
	}
}

func (m *mockStorage) Get(_ context.Context, key string) ([]byte, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, errors.New("file not found")
}

func (m *mockStorage) Put(_ context.Context, key string, data []byte) error {
	m.data[key] = data
	return nil
}

func (m *mockStorage) Exists(_ context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

func (m *mockStorage) Delete(_ context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockStorage) GetStream(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (m *mockStorage) PutStream(_ context.Context, _ string, _ io.Reader) error {
	return errors.New("not implemented")
}

// createTestImage 建立測試用的簡單圖片
func createTestImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// 填充純白色
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.White)
		}
	}

	buf := new(bytes.Buffer)
	_ = png.Encode(buf, img)
	return buf.Bytes()
}

func TestDetectWatermarkFromPath_Success(t *testing.T) {
	// Setup
	cfg := &config.Config{
		BlindWatermark: config.BlindWatermarkConfig{
			Enabled: true,
			Text:    "TEST",
		},
	}

	mockStore := newMockStorage()
	svc := NewWatermarkService(cfg, mockStore)

	// 建立測試圖片並存入 mock storage
	testPath := "uploads/2025/12/26/test.png"
	testImageData := createTestImage(128, 128)
	mockStore.data[testPath] = testImageData

	// Execute
	result, err := svc.DetectWatermarkFromPath(context.Background(), testPath)

	// Verify - 不應有錯誤
	if err != nil {
		t.Fatalf("預期無錯誤，但得到 %v", err)
	}

	// 確認結果結構存在
	if result == nil {
		t.Fatal("預期有結果，但得到 nil")
	}

	// 由於測試圖片沒有嵌入浮水印，提取的結果可能為空或不匹配
	// 這裡只驗證函式正確執行
	t.Logf("檢測結果: detected=%v, text=%q, confidence=%v",
		result.Detected, result.Text, result.Confidence)
}

func TestDetectWatermarkFromPath_FileNotFound(t *testing.T) {
	// Setup
	cfg := &config.Config{
		BlindWatermark: config.BlindWatermarkConfig{
			Enabled: true,
			Text:    "TEST",
		},
	}

	mockStore := newMockStorage()
	svc := NewWatermarkService(cfg, mockStore)

	// Execute - 使用不存在的路徑
	_, err := svc.DetectWatermarkFromPath(context.Background(), "nonexistent/path.jpg")

	// Verify - 應該返回錯誤
	if err == nil {
		t.Fatal("預期有錯誤，但得到 nil")
	}

	expectedErrMsg := "file not found"
	if err.Error() != expectedErrMsg {
		t.Errorf("預期錯誤訊息 %q，但得到 %q", expectedErrMsg, err.Error())
	}
}

func TestDetectWatermark_InvalidImage(t *testing.T) {
	// Setup
	cfg := &config.Config{
		BlindWatermark: config.BlindWatermarkConfig{
			Enabled: true,
			Text:    "TEST",
		},
	}

	mockStore := newMockStorage()
	svc := NewWatermarkService(cfg, mockStore)

	// 建立無效的圖片資料
	invalidData := bytes.NewReader([]byte("this is not an image"))

	// Execute
	_, err := svc.DetectWatermark(context.Background(), invalidData)

	// Verify - 應該返回解碼錯誤
	if err == nil {
		t.Fatal("預期有錯誤，但得到 nil")
	}
}

func TestDetectWatermark_EmptyExpectedText(t *testing.T) {
	// Setup - 沒有設定預期文字
	cfg := &config.Config{
		BlindWatermark: config.BlindWatermarkConfig{
			Enabled: true,
			Text:    "", // 空字串
		},
	}

	mockStore := newMockStorage()
	svc := NewWatermarkService(cfg, mockStore)

	// 建立測試圖片
	testImageData := createTestImage(128, 128)
	reader := bytes.NewReader(testImageData)

	// Execute
	result, err := svc.DetectWatermark(context.Background(), reader)

	// Verify
	if err != nil {
		t.Fatalf("預期無錯誤，但得到 %v", err)
	}

	// 當沒有預期文字時，confidence 應為 0.5
	if result.Confidence != 0.5 {
		t.Errorf("預期 confidence 為 0.5，但得到 %v", result.Confidence)
	}
}
