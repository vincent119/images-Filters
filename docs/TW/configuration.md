# 設定指南

[English](../configuration.md)

## 概述

Images Filters 使用階層式設定系統，由 `config/config.yaml` 控制。所有數值皆可透過環境變數覆蓋。
優先順序為：環境變數 > 設定檔 > 預設值。

### 設定檔 (`config.yaml`)

完整範例請參考 [`config/config.sample.yaml`](../../config/config.sample.yaml)。

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  max_request_size: 10485760 # 10MB

processing:
  default_quality: 85
  max_width: 4096
  max_height: 4096
  workers: 4
  default_format: "jpeg"

security:
  enabled: true
  security_key: "your-secret-key"
  allow_unsafe: false
  allowed_sources: []
  max_width: 4096    # 請求允許的最大寬度
  max_height: 4096   # 請求允許的最大高度

storage:
  type: "local" # local, s3, mixed
  local:
    root_path: "./data/images"
  s3:
    bucket: "my-bucket"
    region: "us-east-1"
    access_key: ""
    secret_key: ""

cache:
  enabled: true
  type: "redis" # memory, redis
  memory:
    max_size: 536870912 # 512MB
    ttl: 3600
  redis:
    host: "localhost"
    port: 6379
    pool:
      size: 10
      timeout: 4
    tls:
      enabled: false

logging:
  level: "info" # debug, info, warn, error
  format: "json" # json, text, console
  output: "stdout" # stdout, file

metrics:
  enabled: true
  namespace: "imgfilter"
  path: "/metrics"
  username: ""  # Basic Auth（可選）
  password: ""

swagger:
  enabled: true
  path: "/swagger"
  username: ""  # Basic Auth（可選）
  password: ""

blind_watermark:
  enabled: true
  text: "COPYRIGHT"
```

### 環境變數

所有設定鍵值皆對應至環境變數。陣列與巢狀物件使用底線 `_` 分隔，並加上前綴 `IMG_`。

| 變數名稱 | 說明 | 預設值 |
| -------- | ---- | ------ |
| `IMG_SERVER_PORT` | 伺服器監聽埠號 | `8080` |
| `IMG_LOGGING_LEVEL` | 日誌等級 | `info` |
| `IMG_CACHE_TYPE` | 快取後端 | `memory` |
| `IMG_CACHE_REDIS_HOST` | Redis 主機 | `localhost` |
| `IMG_SECURITY_ENABLED` | 啟用 HMAC 簽名驗證 | `false` |
| `IMG_SECURITY_SECURITY_KEY` | HMAC 金鑰 | `""` |
| `IMG_STORAGE_TYPE` | 儲存後端 | `local` |
| `IMG_BLIND_WATERMARK_ENABLED` | 啟用隱形浮水印 | `true` |
| `IMG_BLIND_WATERMARK_TEXT` | 浮水印文字 | `""` |
| `IMG_METRICS_NAMESPACE` | Prometheus 命名空間 | `imgfilter` |
