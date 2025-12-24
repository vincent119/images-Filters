# Troubleshooting Guide

[English](../troubleshooting.md)

## 常見問題

### 1. "Signature Mismatch" 錯誤

**徵狀**: API 回傳 `400 Bad Request` 且訊息為 `invalid_signature`。
**原因**:

- 伺服器端的 `SECURITY_KEY` 與簽名使用的金鑰不同。
- 簽名器生成的 URL 路徑與實際請求的路徑不完全一致。
**解決方案**:
- 檢查環境變數設定。
- 使用 CLI 簽名工具進行驗證。

#### 2. "Image Not Found" 錯誤

**徵狀**: `404 Not Found`。
**原因**:

- `http` 載入器無法從來源獲取圖片。
- 本地檔案不存在。
**解決方案**:
- 檢查伺服器日誌以獲取詳細上游錯誤（如 DNS 錯誤、來源回傳 403）。

#### 3. 高延遲 / 處理緩慢

**徵狀**: 回應時間 > 1秒。
**原因**:

- 處理過大圖片（CPU 瓶頸）。
- 從慢速來源下載大圖（網路瓶頸）。
- 快取未命中。
**解決方案**:
- 啟用 Redis 快取。
- 檢查 Prometheus 指標 `imgfilter_processing_duration_seconds`。
- 若可能，調整原始圖片尺寸使其接近目標尺寸。

### 除錯模式

設定 `LOG_LEVEL=debug` 可查看詳細的請求流程與處理步驟。

```bash
export LOG_LEVEL=debug
./images-filters
```
