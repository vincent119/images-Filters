# Configuration Guide

[English](../configuration.md)

## 概述

Images Filters 使用 `config/config.yaml` 進行各項參數設定，並支援透過環境變數覆蓋設定值。
優先順序為：環境變數 > 設定檔 > 預設值。

### 設定檔 (`config.yaml`)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug" # 運行模式：debug, release, test

logging:
  level: "info" # 日誌等級：debug, info, warn, error
  format: "json" # 格式：json, text, console
  output: "stdout" # 輸出：stdout, file
  file_path: "./logs/app.log"

processing:
  default_quality: 80 # 圖片預設品質
  max_width: 5000 # 最大寬度限制
  max_height: 5000 # 最大高度限制
  workers: 4 # 處理圖片的 Worker 數量

cache:
  type: "redis" # 快取類型：memory, redis
  memory:
    max_size: 536870912 # 記憶體上限 (Bytes)
    ttl: 3600 # 過期時間 (秒)
  redis:
    host: "localhost"
    port: 6379
    username: "" # ACL 使用者 (可選)
    password: ""
    db: 0
    ttl: 3600
    pool:
      size: 10 # 連線池大小
      timeout: 4
    tls:
      enabled: false # 是否啟用 TLS

security:
  enabled: true # 是否啟用簽名驗證
  security_key: "your-secret-key" # HMAC 簽名密鑰
  allow_unsafe: false # 是否允許 /unsafe 路徑
```

### 環境變數 (Environment Variables)

所有設定鍵值皆對應至環境變數。嵌套結構使用底線 `_` 分隔。

| 變數名稱 | 說明 | 預設值 |
| ---------- | ------------- | --------- |
| `SERVER_PORT` | 伺服器監聽埠 | `8080` |
| `LOG_LEVEL` | 日誌等級 | `info` |
| `CACHE_TYPE` | 快取後端 | `memory` |
| `CACHE_REDIS_HOST` | Redis 主機 | `localhost` |
| `SECURITY_ENABLED` | 啟用 HMAC 檢查 | `false` |
| `SECURITY_KEY` | HMAC 密鑰 | `""` |
