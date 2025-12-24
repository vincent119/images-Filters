# API Specification

[English](../api.md)

## 概述

- **基礎 URL**: `http://<host>:<port>`
- **內容類型**: 通常回傳圖片二進位資料 (`image/jpeg`, `image/png` 等) 或 JSON 錯誤訊息。
- **監控端點**: `/metrics` (Prometheus 格式)
- **健康檢查**: `/healthz`

### API 端點

#### 1. 圖片處理 (Process Image)

根據指定選項與濾鏡處理圖片。

- **開發模式 URL** (Unsafe):
  `GET /unsafe/<options>/<filters>/<image_path>`

- **生產模式 URL** (Signed):
  `GET /<signature>/<options>/<filters>/<image_path>`

**參數說明:**

| 參數 | 說明 | 格式 / 範例 |
| ----------- | ------------- | ------------------ |
| `signature` | HMAC-SHA256 簽名 | Base64 編碼字串 |
| `options` | 處理選項 | `寬x高` (例如 `300x200`，負值代表翻轉如 `-300x200`) |
| `filters` | 濾鏡鏈 (可選) | `filters:濾鏡1(參數):濾鏡2(參數)` |
| `image_path` | 原始圖片路徑/URL | URL 編碼後路徑 (例如 `images/test.jpg` 或 `http%3A%2F%2F...`) |

**支援的濾鏡:**

- `blur(sigma)` : 高斯模糊。
- `grayscale()` : 轉為灰階。
- `brightness(factor)` : 調整亮度 (-100 到 100)。
- `contrast(factor)` : 調整對比度 (-100 到 100)。
- `watermark(image_url,opacity,x,y)` : 添加浮水印。

**回應:**

- `200 OK`: 回傳處理後的圖片檔案。
- `400 Bad Request`: 參數錯誤或簽名無效。
- `404 Not Found`: 找不到原始圖片。
- `500 Internal Server Error`: 圖片處理失敗。

#### 2. 健康檢查 (Health Check)

檢查服務運作狀態。

- **URL**: `GET /healthz`
- **回應**:

  ```json
  {
    "status": "ok",
    "timestamp": "2024-03-20T10:00:00Z"
  }
  ```

#### 3. 監控指標 (Metrics)

提供 Prometheus 格式監控數據。

- **URL**: `GET /metrics`

### 錯誤代碼 (Error Codes)

錯誤回應使用 JSON 格式。

```json
{
  "error": "code",
  "message": "錯誤描述訊息"
}
```

| 代碼 | 訊息範例 | 說明 |
| ------ | ----------------- | ------------- |
| `invalid_signature` | "HMAC signature mismatch" | URL 簽名無效或不符。 |
| `image_not_found` | "failed to fetch source image" | 無法讀取指定原始圖片。 |
| `invalid_params` | "invalid width parameter" | 輸入參數格式錯誤。 |
| `processing_error` | "decode failed" | 圖片解碼或處裡過程發生內部錯誤。 |
| `rate_limit` | "too many requests" | 超出請求頻率限制。 |
