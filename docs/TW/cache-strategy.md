# Cache Strategy

[English](../cache-strategy.md)

## 策略概述

快取對效能至關重要。Images Filters 採用 "Cache-Aside" 策略，並針對不可變內容進行最佳化。

### 快取鍵 (Key) 設計

快取鍵是根據請求參數生成的唯一識別碼。

**格式:**
`images:processed:<hash>`

**雜湊建構:**
`options + filters + image_path` 的 SHA256 值。

範例: `300x200/filters:grayscale()/image.jpg` -> `SHA256(...)`

### 快取層級

1. **瀏覽器快取 (Client)**
   - 透過 HTTP 標頭控制 (`Cache-Control`, `ETag`)。
   - 預設: `public, max-age=31536000` (1 年)。

2. **CDN (Edge)**
   - 建議的部署架構會在前方設置 CDN (Cloudflare/CloudFront)。
   - 卸載靜態資源傳輸負載。

3. **應用程式快取 (Server)**
   - **Redis**: 生產環境推薦。可跨多實例共享。
   - **記憶體 (In-Memory)**: 用於本地開發或單機部署。

### 設定

```yaml
cache:
  type: redis
  redis:
    ttl: 3600 # 預設 1 小時
    pool:
      size: 10
```

### 快取失效 (Invalidation)

由於圖片 URL 是基於參數決定的：

- **更變參數**: 請求新的圖片邏輯，自然繞過舊快取。
- **變更原始圖片**: 若原始圖片內容變更但檔名未變，快取可能會提供舊內容直到 TTL 過期。
  - **最佳實踐**: 對原始圖片使用版本化檔名 (例如 `image-v1.jpg`)。
