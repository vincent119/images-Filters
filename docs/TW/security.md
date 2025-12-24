# Security Design

[English](../security.md)

## 概述

安全性是圖片處理服務的首要考量，旨在防止未經授權的資源消耗與拒絕服務攻擊 (DoS)。Images Filters 在所有生產環境請求中實作了 HMAC-SHA256 簽名驗證。

### URL 簽名 (HMAC)

為了防止攻擊者請求任意尺寸的圖片並耗盡伺服器 CPU/Memory，所有 URL 都必須經過簽名。

#### 演算法

1. **建構路徑**: 包含選項、濾鏡與圖片路徑的部分 URL。
   範例: `300x200/filters:blur(5)/image.jpg`
2. **金鑰**: 伺服器設定中的 `SECURITY_KEY`。
3. **簽名**: 使用金鑰對路徑計算 HMAC-SHA256 值。
4. **編碼**: 將結果進行 Base64 URL-safe 編碼。

#### 實作範例 (Go)

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
)

func SignURL(key string, path string) string {
    mac := hmac.New(sha256.New, []byte(key))
    mac.Write([]byte(path))
    signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))
    return signature
}
```

### 存取控制

- **不安全路徑**: `/unsafe/...` 嚴格僅供開發使用。生產環境必須停用 (`SECURITY_ALLOW_UNSAFE=false`)。
- **來源驗證**: 目前服務支援任意來源 URL。未來版本將加入 `http` 載入器的網域白名單 (Whitelist)。

### DoS 防護

- **資源限制**: 在設定中限制 `max_width` 與 `max_height`，防止處理超大圖片（Pixel Bombs）。
- **逾時控制**: 對圖片下載與處理設定嚴格的逾時時間，確保 Worker 執行緒能被釋放。
