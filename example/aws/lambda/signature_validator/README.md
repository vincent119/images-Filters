# Lambda@Edge: HMAC 簽名驗證

## 概述

此 Lambda@Edge 函式在 CloudFront 邊緣節點執行完整的 HMAC-SHA256 簽名驗證，在請求到達 Origin 之前拒絕無效請求。

## 功能

1. **完整簽名驗證** - HMAC-SHA256 + Base64 URL-safe 編碼
2. **時序安全比對** - 防止計時攻擊
3. **開發模式支援** - 可配置是否允許 `/unsafe/` 路徑

## 與 CloudFront Function 的差異

| 特性 | CloudFront Function | Lambda@Edge |
| ---- | ------------------- | ----------- |
| 執行時間限制 | 1ms | 5-30 秒 |
| 記憶體限制 | 2MB | 128MB-10GB |
| 網路存取 | ❌ | ✅ |
| Secrets Manager | ❌ | ✅ |
| 完整 HMAC 驗證 | ❌ | ✅ |
| 成本 | 較低 | 較高 |

**建議策略**：使用 CloudFront Function 做格式驗證 + Lambda@Edge 做完整簽名驗證

## 部署步驟

### 1. 建立 IAM 角色

```bash
# 建立信任政策
cat > trust-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "lambda.amazonaws.com",
          "edgelambda.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

# 建立角色
aws iam create-role \
  --role-name images-filters-edge-role \
  --assume-role-policy-document file://trust-policy.json

# 附加基本執行政策
aws iam attach-role-policy \
  --role-name images-filters-edge-role \
  --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
```

### 2. 打包並上傳 Lambda

```bash
# 打包
cd example/aws/lambda/signature_validator
zip -r function.zip index.js

# 建立函式（必須在 us-east-1）
aws lambda create-function \
  --region us-east-1 \
  --function-name images-filters-signature-validator \
  --runtime nodejs18.x \
  --role arn:aws:iam::ACCOUNT_ID:role/images-filters-edge-role \
  --handler index.handler \
  --zip-file fileb://function.zip

# 發布版本
aws lambda publish-version \
  --region us-east-1 \
  --function-name images-filters-signature-validator
```

### 3. 關聯到 CloudFront Distribution

在 CloudFront Console：

1. 選擇 Distribution → Behaviors
2. 編輯目標 Behavior
3. 在 **Lambda function associations** 區塊
4. 選擇 **Viewer request** 或 **Origin request**
5. 輸入 Lambda ARN（含版本號）

或使用 AWS CLI：

```bash
aws cloudfront update-distribution \
  --id DISTRIBUTION_ID \
  --distribution-config file://distribution-config.json
```

## 安全金鑰管理

### 方法 1：硬編碼（不推薦）

直接在程式碼中設定 `SECURITY_KEY`。

### 方法 2：Secrets Manager（推薦）

```javascript
const { SecretsManagerClient, GetSecretValueCommand } = require('@aws-sdk/client-secrets-manager');

const client = new SecretsManagerClient({ region: 'us-east-1' });
let cachedKey = null;

async function getSecurityKey() {
    if (cachedKey) return cachedKey;

    const response = await client.send(new GetSecretValueCommand({
        SecretId: 'images-filters/security-key'
    }));

    cachedKey = response.SecretString;
    return cachedKey;
}
```

**注意**：使用 Secrets Manager 會增加延遲，建議使用記憶體快取。

## 監控

Lambda@Edge 日誌會寫入執行區域的 CloudWatch：

- 日誌組：`/aws/lambda/us-east-1.images-filters-signature-validator`
- 指標：`Invocations`, `Errors`, `Duration`, `Throttles`

## 成本估算

- Lambda@Edge 計費：每 100 萬次請求約 $0.60
- 執行時間：每 GB-秒 $0.00005001
- 建議搭配 CloudFront 快取減少呼叫次數
