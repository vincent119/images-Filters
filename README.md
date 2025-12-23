# Images Filters

[![Go Version](https://img.shields.io/badge/Go-1.23-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](link)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](link)

高效能圖片處理服務器，支援即時 resize、crop、flip、filters 和 watermark，參考 [Thumbor](https://github.com/thumbor/thumbor) 設計理念。

## ✨ 功能特點

- 🖼️ **圖片處理**：Resize、Crop、Flip、Rotate
- 🎨 **濾鏡效果**：Blur、Grayscale、Brightness、Contrast、Sharpen 等
- 💧 **浮水印**：支援圖片浮水印，可調整位置與透明度
- 🔒 **安全機制**：HMAC URL 簽名防止篡改
- 📦 **多種儲存**：本地、AWS S3、混合模式
- ⚡ **高效能**：Redis 快取、Worker Pool
- 📊 **監控**：Prometheus 指標

## 📸 支援的圖片格式

| 格式 | 讀取 | 寫入 | 備註 |
| ------ | :----: | :----: | ------ |
| JPEG | ✅ | ✅ | 最常用格式 |
| PNG | ✅ | ✅ | 支援透明 |
| WebP | ✅ | ✅ | 現代瀏覽器推薦 |
| AVIF | ✅ | ✅ | 2024 新格式 |
| JPEG XL | ✅ | ✅ | 未來趨勢 |
| GIF | ✅ | ✅ | 支援動畫 |
| HEIC | ✅ | ❌ | Apple 格式 |
| SVG | ✅ | ❌ | 向量圖轉換 |

## 🚀 快速開始

### 安裝

```bash
# Clone 專案
git clone https://github.com/vincent119/images-filters.git
cd images-filters

# 安裝依賴
go mod tidy

# 執行
make run
```

### Docker

```bash
# 建置映像
make docker-build

# 執行容器
docker run -p 8080:8080 images-filters:latest
```

## 📖 使用方式

### URL 格式

```bash
http://<server>/<signature>/<options>/<filters>/<image_path>
```

### 範例

```bash
# Resize 到 300x200
http://localhost:8080/unsafe/300x200/https%3A%2F%2Fexample.com%2Fimage.jpg

# 套用灰階濾鏡
http://localhost:8080/unsafe/300x200/filters:grayscale()/https%3A%2F%2Fexample.com%2Fimage.jpg

# 水平翻轉 + 模糊
http://localhost:8080/unsafe/-300x200/filters:blur(5)/https%3A%2F%2Fexample.com%2Fimage.jpg
```

## 🛠️ 開發

```bash
# 執行測試
make test

# 執行 Lint
make lint

# 格式化程式碼
make fmt

# 生成 Swagger 文檔
make swagger
```

## 📁 專案結構

```bash
images-Filters/
├── cmd/server/         # 應用程式入口
├── internal/
│   ├── api/            # HTTP 處理器
│   ├── service/        # 業務邏輯層
│   ├── processor/      # 圖片處理核心
│   ├── filter/         # 濾鏡管線
│   ├── loader/         # 圖片載入器
│   ├── storage/        # 儲存層
│   ├── security/       # 安全機制
│   └── cache/          # 快取層
├── pkg/                # 共用工具
├── config/             # 設定檔
├── docs/               # 文件
├── deploy/             # 部署設定
└── charts/             # Helm Charts
```

## 📝 License

MIT License
