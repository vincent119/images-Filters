# CloudFront Function: URL 驗證與正規化

## 概述

此 CloudFront Function 在邊緣節點執行輕量級 URL 驗證，快速拒絕明顯無效的請求，減少對 Origin 的負擔。

## 功能

1. **簽名格式驗證** - 檢查簽名長度與字元格式
2. **URL 正規化** - 移除重複斜線，統一格式
3. **開發模式支援** - 允許 `/unsafe/` 路徑通過

## 限制

- **執行時間**: 1ms（無法執行複雜的 HMAC 驗證）
- **記憶體**: 2MB
- **不支援**: 網路請求、檔案系統存取

## 部署步驟

### 1. 建立 Function

```bash
aws cloudfront create-function \
  --name images-filters-url-validator \
  --function-config '{"Comment":"URL validation for images-filters","Runtime":"cloudfront-js-1.0"}' \
  --function-code fileb://url_validator.js
```

### 2. 發布 Function

```bash
aws cloudfront publish-function \
  --name images-filters-url-validator \
  --if-match <ETAG>
```

### 3. 關聯到 Distribution

在 CloudFront Distribution 的 Behavior 設定中：

1. 選擇目標 Behavior（如 `Default (*)`）
2. 在 **Function associations** 區塊
3. **Viewer request** 選擇 `images-filters-url-validator`

## 測試

```bash
# 測試有效 URL
curl -I "https://your-distribution.cloudfront.net/ABC123.../300x200/image.jpg"

# 測試無效簽名（應返回 403）
curl -I "https://your-distribution.cloudfront.net/short/300x200/image.jpg"

# 測試 unsafe 模式
curl -I "https://your-distribution.cloudfront.net/unsafe/300x200/image.jpg"
```

## 監控

CloudFront Function 執行指標可在 CloudWatch 中查看：

- `FunctionInvocations` - 調用次數
- `FunctionExecutionErrors` - 執行錯誤
- `FunctionThrottles` - 節流次數
