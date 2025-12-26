# Images Filters 圖片處理服務

[![Go Version](https://img.shields.io/badge/Go-1.25.5-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://github.com/vincent119/images-filters/actions/workflows/go.yml/badge.svg)](https://github.com/vincent119/images-filters/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/vincent119/images-filters/badge.svg?branch=main)](https://coveralls.io/github/vincent119/images-filters?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/vincent119/images-filters)](https://goreportcard.com/report/github.com/vincent119/images-filters)

[English](README.md)

高效能圖片處理服務器，支援即時縮放、裁切、翻轉、濾鏡與浮水印。專為速度與可擴充性優化。

## ✨ 功能特點

- 🖼️ **圖片處理**：即時縮放 (Resize)、裁切 (Crop)、翻轉 (Flip)、旋轉 (Rotate)、格式轉換。
- 🎨 **濾鏡效果**：模糊 (Blur)、灰階 (Grayscale)、亮度 (Brightness)、對比度 (Contrast)、銳化 (Sharpen) 等。
- 💧 **浮水印**：支援圖片浮水印與**隱形浮水印 (Blind Watermark)**。
- 🔒 **安全機制**：HMAC-SHA256 URL 簽名驗證，防止惡意竄改。
- 📦 **多種儲存**：支援本地檔案系統、AWS S3 以及混合模式（本地快取 + 遠端來源）。
- ⚡ **高效能**：內建 Redis 快取機制、Worker Pool 處理池、Go 並發優化。
- 📊 **可觀測性**：完整 Prometheus 監控指標、Grafana 儀表板、結構化日誌。
- 🐳 **雲原生**：提供 Docker 映像檔、Helm Charts 與 Kustomize 部署支援。

## 🚀 快速開始

### 安裝執行

```bash
# 下載專案
git clone https://github.com/vincent119/images-filters.git
cd images-filters

# 安裝依賴
go mod tidy

# 啟動服務
make run
```

### Docker 執行

```bash
# 啟動容器
docker run -p 8080:8080 vincent119/images-filters:latest
```

## 📖 使用說明

### URL 格式

```bash
http://<server>/<signature>/<options>/<filters>/<image_path>
```

### 範例請求

```bash
# 縮放到 300x200 (開發模式)
http://localhost:8080/unsafe/300x200/https%3A%2F%2Fexample.com%2Fimage.jpg

# 套用灰階濾鏡
http://localhost:8080/unsafe/300x200/filters:grayscale()/image.jpg

# 簽名 URL (生產模式)
http://localhost:8080/H9a8s.../300x200/image.jpg
```

更多詳細資訊請參閱 [完整文件](docs/TW/README.md)。

## 📚 文件索引 (Documentation)

### 核心文件 (Core Docs)

- [系統架構 (Architecture)](docs/TW/architecture.md)
- [API 規格 (API Specification)](docs/TW/api.md)
- [安全設計 (Security Design)](docs/TW/security.md)
- [配置說明 (Configuration)](docs/TW/configuration.md)
- [隱形浮水印 (Blind Watermark)](docs/TW/blind-watermark.md)

### 進階指南 (Advanced Guides)

- [圖片處理流程 (Image Pipeline)](docs/TW/image-pipeline.md)
- [快取策略 (Cache Strategy)](docs/TW/cache-strategy.md)
- [監控與指標 (Observability)](docs/TW/observability.md)
- [效能調優 (Performance)](docs/TW/performance.md)

### 維運與開發 (Ops & Dev)

- [部署指南 (Deployment)](docs/TW/deployment.md)
- [除錯指南 (Troubleshooting)](docs/TW/troubleshooting.md)
- [開發指南 (Developer Guide)](docs/TW/dev-guide.md)
- [貢獻規範 (Contributing)](docs/TW/contributing.md)

## 🛠️ 開發指令

```bash
# 執行測試
make test

# 程式碼檢查
make lint

# 產生 Swagger 文件
make swagger
```

## 📝 License

MIT License
