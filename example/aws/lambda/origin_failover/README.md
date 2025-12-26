# Lambda@Edge: Origin Group Failover

## 架構概述

此方案實現「S3 優先 + API Server Failover」的雙層架構：

```text
                     CloudFront
                         │
         ┌───────────────┼───────────────┐
         ▼               ▼               ▼
   [簽名驗證]      [Origin Request]  [Origin Response]
   (Viewer Req)    選擇 Origin       處理 Failover
         │               │               │
         ▼               ▼               ▼
    通過/拒絕     嘗試 S3 讀取      S3 失敗時
                         │           打 API Server
                         ▼
                  ┌──────┴──────┐
                  ▼             ▼
               找到 (200)    沒找到 (404)
                  │             │
                  ▼             ▼
               回傳 S3       Origin Response
               快取圖片      打 API Server
                              處理圖片
```

## 檔案說明

| 檔案 | 部署位置 | 功能 |
| ---- | -------- | ---- |
| `origin_request.js` | Origin Request | 將請求導向 S3（快取層） |
| `origin_response.js` | Origin Response | S3 Miss 時 Failover 到 API Server |

## 優缺點比較

### 方式 1：純 API Server Origin

```text
CloudFront → API Server → 處理圖片
```

| 優點 | 缺點 |
| ---- | ---- |
| 架構簡單 | API Server 負載高 |
| 維護容易 | 每次 Cache Miss 都要處理 |
| 不需額外 S3 成本 | 擴展性受限於 Server |

**適合**：流量小、圖片處理簡單、開發測試階段

---

### 方式 2：S3 + API Server Failover

```text
CloudFront → S3 (快取) → 沒有則 → API Server
```

| 優點 | 缺點 |
| ---- | ---- |
| API Server 負載極低 | 架構複雜度高 |
| S3 讀取速度快 | 需額外 S3 儲存成本 |
| 可無限擴展 | Lambda@Edge 有執行成本 |
| 降低處理延遲 | 需維護 S3 快取清理策略 |

**適合**：高流量、需要極致效能、生產環境

---

## 結論：哪個比較優秀？

| 場景 | 推薦方式 |
| ---- | -------- |
| 日請求 < 100 萬 | 方式 1（簡單直接） |
| 日請求 > 100 萬 | 方式 2（效能優先） |
| 開發/測試環境 | 方式 1 |
| 生產環境 + 高 SLA | 方式 2 |
| 預算有限 | 方式 1 |
| 追求最低延遲 | 方式 2 |

**建議策略**：先用方式 1，當流量成長後再遷移到方式 2。

---

## 部署步驟

### 1. 建立 S3 Bucket

```bash
aws s3 mb s3://your-images-cache-bucket --region ap-northeast-1

# 設定 Bucket Policy 允許 CloudFront 讀取
aws s3api put-bucket-policy --bucket your-images-cache-bucket \
  --policy file://s3-bucket-policy.json
```

### 2. 部署 Lambda@Edge

```bash
# 打包
cd example/aws/lambda/origin_failover
zip origin_request.zip origin_request.js
zip origin_response.zip origin_response.js

# 部署 Origin Request Lambda
aws lambda create-function \
  --region us-east-1 \
  --function-name images-filters-origin-request \
  --runtime nodejs18.x \
  --role arn:aws:iam::ACCOUNT_ID:role/images-filters-edge-role \
  --handler origin_request.handler \
  --zip-file fileb://origin_request.zip

# 部署 Origin Response Lambda
aws lambda create-function \
  --region us-east-1 \
  --function-name images-filters-origin-response \
  --runtime nodejs18.x \
  --role arn:aws:iam::ACCOUNT_ID:role/images-filters-edge-role \
  --handler origin_response.handler \
  --zip-file fileb://origin_response.zip

# 發布版本
aws lambda publish-version --region us-east-1 --function-name images-filters-origin-request
aws lambda publish-version --region us-east-1 --function-name images-filters-origin-response
```

### 3. 設定 CloudFront Distribution

在 CloudFront Behavior 中：
- **Origin Request**: 關聯 `images-filters-origin-request`
- **Origin Response**: 關聯 `images-filters-origin-response`

## 進階優化

### 寫回 S3 快取

在 `origin_response.js` 中，可以將 API Server 處理後的圖片存回 S3：

```javascript
const { S3Client, PutObjectCommand } = require('@aws-sdk/client-s3');

async function saveToS3(key, body, contentType) {
    const client = new S3Client({ region: S3_REGION });
    await client.send(new PutObjectCommand({
        Bucket: S3_BUCKET,
        Key: key,
        Body: Buffer.from(body, 'base64'),
        ContentType: contentType,
        CacheControl: 'max-age=31536000'
    }));
}
```

### 快取清理策略

- 使用 S3 Lifecycle Policy 自動清理舊圖片
- 設定 TTL（如 30 天後自動刪除）

```json
{
  "Rules": [
    {
      "ID": "CleanupProcessedImages",
      "Prefix": "processed/",
      "Status": "Enabled",
      "Expiration": { "Days": 30 }
    }
  ]
}
```
