# 隱形浮水印 (Blind Watermark)

[English](../blind-watermark.md)

**隱形浮水印**功能允許您將看不見的版權資訊嵌入圖片中，並在日後提取以進行驗證。與可見浮水印不同，隱形浮水印不會影響圖片的視覺品質，且即使經過圖片壓縮、縮放（在一定限度內）或裁切（部分抗性），浮水印仍然存在。

## 運作原理

本專案使用頻域嵌入 (DCT - 離散餘弦變換) 邏輯：

1. **嵌入 (Embedding)**:
   - 將圖片透過 DCT 轉換為頻域。
   - 將浮水印文字轉換為二進位數據。
   - 將二進位數據嵌入到圖片的中頻係數中。
   - 將圖片透過 IDCT 轉回空間域。

2. **檢測 (Detection)**:
    - 將待檢測的圖片轉換為頻域。
    - 從係數中提取二進位數據。
    - 重組文字並與預期的浮水印文字進行比對。

## 設定配置

在 `config.yaml` 中啟用隱形浮水印：

```yaml
# 隱形浮水印設定
blind_watermark:
  enabled: true
  text: "COPYRIGHT"  # 要嵌入的文字
  strength: 5.0      # 嵌入強度 (數值越高越穩健，但雜訊可能越明顯)
```

## 使用方式

### 1. 自動嵌入

啟用後，所有處理過的圖片都會自動套用隱形浮水印，除非特別禁用。

範例請求：
`GET /<signature>/300x200/smart/image.jpg`

回傳的圖片將包含隱藏文字 "COPYRIGHT"。

### 2. 透過 API 手動檢測

您可以使用 `/detect` API 來檢測浮水印。此 API 需要在 header 中帶上 `Authorization` 與您的 `SECURITY_KEY`。

#### 透過檔案上傳檢測

```bash
curl -X POST -H "Authorization: Bearer <YOUR_SECURITY_KEY>" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@/path/to/suspected_image.jpg" \
  http://localhost:8080/detect
```

#### 透過儲存路徑檢測

如果圖片已經存在於您的儲存後端 (本地或 S3)：

```bash
curl -X POST -H "Authorization: Bearer <YOUR_SECURITY_KEY>" \
  -d "path=uploads/2025/12/26/image.jpg" \
  http://localhost:8080/detect
```

**回應範例:**

```json
{
  "detected": true,
  "text": "COPYRIGHT",
  "confidence": 0.98
}
```

### 3. 圖片上傳與浮水印

您可以使用 `/upload` API 上傳原始圖片。系統會儲存原始檔，但在透過標準處理流程讀取時，可以自動加上浮水印。

## 限制與注意事項

- **壓縮**: 極高程度的壓縮 (例如 JPEG 品質 < 50) 可能會破壞浮水印。
- **裁切**: 雖然浮水印是全域嵌入的，但若裁切部分過小（例如小於 64x64），可能無法成功提取。
- **旋轉**: 目前實作對旋轉較敏感。檢測前建議先將圖片擺正。

## 穩健性測試

我們已針對以下情境進行測試：

- **縮放**: 可抵抗縮小至 50%。
- **JPEG 壓縮**: 可抵抗品質 80 以上的壓縮。
- **格式轉換**: 可抵抗 PNG <-> JPEG 之間的轉換。
