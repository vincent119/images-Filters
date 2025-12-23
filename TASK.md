# 圖片處理服務器 - 任務計畫

> 此文件追蹤圖片處理服務器專案的開發進度

---

## Phase 1: 核心功能 (MVP)

### 1.1 專案初始化

- [x] 建立 Go module (`go mod init`)
- [x] 建立專案目錄結構
- [x] 建立 `.golangci.yml` Linter 設定
- [x] 設定 Makefile
- [x] 建立 `.gitignore`
- [x] 初始化 README.md（含 Badges）
- [x] 設定 Dockerfile 基礎版本
- [x] 建立 `docs/` 目錄結構

### 1.2 設定管理

- [x] 安裝 Viper 依賴
- [x] 建立 `internal/config/config.go`
- [x] 定義設定結構體 (Config struct)
- [x] 建立 `config/config.yaml` 範例
- [x] 實作環境變數覆蓋功能
- [x] 撰寫設定載入單元測試
- [x] validate config in struct (使用 go-playground/validator)

### 1.3 HTTP 服務器基礎

- [x] 安裝 Gin 框架
- [x] 建立 `cmd/server/main.go` 入口
- [x] 實作基本路由設定
- [x] 建立健康檢查端點 `/healthz`
- [x] 實作優雅關閉 (Graceful Shutdown)
- [x] 設定 CORS 中介層
- [x] 設定 zlogger for gin middleware (`pkg/logger/gin_middleware.go`)
- [x] 設定 zlogger for service (fmt print 改成 zlogger)
- [x] 設定 zlogger for loader
- [x] logger config output stdout = console
- [x] /healthz and /metrics skip path not write logs
- [x] fmt print 改成 zlogger（已確認無 fmt.Print 使用）
- [x] add swagger path and implement base auth for single config

```yaml
swagger:
  enabled: true
  path: "/swagger"
  username: ""
  password: ""
```

### 1.4 URL 解析器

- [x] 建立 `internal/parser/url_parser.go`
- [x] 定義 ParsedURL 結構體
- [x] 實作尺寸解析 (`300x200`, `300x0`, `0x200`)
- [x] 實作翻轉標記解析 (`-300x200`, `300x-200`)
- [x] 實作 fit-in 模式解析
- [x] 實作裁切座標解析 (`10x20:100x150`)
- [x] 實作濾鏡參數解析 (`filters:blur(7):grayscale()`)
- [x] 實作圖片路徑/URL 解析
- [x] 撰寫 URL 解析單元測試

### 1.5 圖片載入器

- [x] 建立 `internal/loader/interface.go` 定義 Loader 介面
- [x] 實作 `internal/loader/http_loader.go` (HTTP/HTTPS 載入)
- [x] 實作 `internal/loader/file_loader.go` (本地檔案載入)
- [x] 實作 Loader Factory 模式
- [x] 處理載入錯誤和逾時
- [ ] 撰寫載入器單元測試

### 1.6 基本圖片處理

- [x] 安裝 imaging 套件
- [x] 建立 `internal/processor/processor.go` 處理核心
- [x] 實作 Resize 功能
  - [x] 等比例縮放
  - [x] 固定尺寸縮放
  - [x] 只指定寬度/高度
- [x] 實作 Crop 功能
  - [x] 手動座標裁切
  - [x] 中心裁切
- [x] 實作 Flip 功能
  - [x] 水平翻轉
  - [x] 垂直翻轉
- [x] 撰寫處理器單元測試

### 1.7 圖片編碼器

- [x] 建立編碼器（整合於 `processor.go`）
- [x] 實作 JPEG 編碼 (可調品質)
- [x] 實作 PNG 編碼
- [x] 實作 WebP 編碼
- [x] 實作 GIF 編碼
- [x] 實作格式自動偵測
- [x] 撰寫編碼器單元測試

### 1.8 本地儲存

- [x] 建立 `internal/storage/interface.go` 定義介面
- [x] 實作 `internal/storage/local.go`
  - [x] Get 方法
  - [x] Put 方法
  - [x] Exists 方法
  - [x] Delete 方法
- [x] 實作目錄自動建立
- [x] 撰寫儲存單元測試

### 1.9 Service 層建立

- [x] 建立 `internal/service/interface.go` 定義 Service 介面
- [x] 實作 `internal/service/image_service.go` 圖片處理業務邏輯
- [x] 實作圖片處理流程（載入 → 處理 → 編碼）
- [ ] 撰寫 Service 層單元測試

### 1.10 API 處理器整合

- [x] 建立 `internal/api/handler.go`
- [x] Handler 依賴 Service 介面（依賴注入）
- [x] 設定正確的 Content-Type
- [x] 實作錯誤處理和 HTTP 狀態碼
- [x] 建立 `routes/routes.go` 路由定義

### 1.11 Phase 1 測試驗證

- [/] 啟動服務器進行手動測試
- [ ] 測試基本 resize 功能
- [ ] 測試 crop 功能
- [ ] 測試 flip 功能
- [ ] 測試不同圖片格式
- [x] 確認所有單元測試通過

### 1.12 Phase 1 prometheus metrics

- [x] 建立 `internal/metrics/interface.go` 定義介面
- [x] 實作 `internal/metrics/prometheus.go`
- [x] 實作 `internal/metrics/gin_middleware.go` Gin 中介層
- [x] 建立 `/metrics` 路由並實作 Basic Auth (`internal/metrics/handler.go`)
- [x] 建立數據收集邏輯 (處理時間、請求次數、錯誤次數, 請求路徑, 請求方法, 請求狀態碼)
- [x] 數據展示於 `/metrics` 端點（Prometheus 標準格式）
- [x] 撰寫 metrics 單元測試 (`prometheus_test.go`)
- [x] 加入 image type count, image size count 指標（整合於 service 層）

### 1.13 uber-go/fx 依賴注入重構

- [x] 安裝 `uber-go/fx` 套件
- [x] 建立 `internal/fx/` 模組目錄
- [x] 實作 Config Module (`internal/fx/config.go`)
- [x] 實作 Logger Module (`internal/fx/logger.go`)
- [x] 實作 Metrics Module (`internal/fx/metrics.go`)
- [x] 實作 Service Module (`internal/fx/service.go`)
- [x] 實作 HTTP Server Module (`internal/fx/server.go`)
- [x] 重構 `cmd/server/main.go` 使用 fx.New()
- [x] 實作 Lifecycle hooks（啟動/關閉）
- [ ] 撰寫 DI 模組測試

---

## Phase 2: 濾鏡與浮水印

### 2.1 濾鏡管線架構

- [x] 建立 `internal/filter/interface.go` 定義濾鏡介面
- [x] 實作濾鏡管線 (Filter Pipeline)
- [x] 實作濾鏡參數解析器
- [x] 建立濾鏡註冊機制

### 2.2 基本濾鏡實作

- [x] 實作 `blur.go` - 模糊濾鏡
- [x] 實作 `grayscale.go` - 灰階濾鏡
- [x] 實作 `brightness.go` - 亮度調整
- [x] 實作 `contrast.go` - 對比度調整
- [x] 實作 `saturation.go` - 飽和度調整
- [x] 實作 `sharpen.go` - 銳化濾鏡

### 2.3 顏色處理濾鏡

- [x] 實作 `rgb.go` - RGB 調整
- [x] 實作 `sepia.go` - 復古色調
- [x] 實作 `equalize.go` - 均衡化
- [x] 實作 `gamma.go` - Gamma 校正
- [x] 實作 `hue.go` - 色相調整

### 2.4 特效濾鏡

- [x] 實作 `rotate.go` - 旋轉
- [x] 實作 `round_corners.go` - 圓角
- [x] 實作 `noise.go` - 雜訊效果
- [x] 實作 `fliph.go` - 水平翻轉
- [x] 實作 `flipv.go` - 垂直翻轉
- [x] 實作 `pixelate.go` - 像素化

### 2.5 輸出控制濾鏡

- [x] 實作 `quality.go` - 品質控制
- [x] 實作 `format.go` - 格式轉換
- [x] 實作 `strip_exif.go` - 移除 EXIF
- [x] 實作 `strip_icc.go` - 移除 ICC Profile
- [x] 實作 `autoorient.go` - 自動方向校正

### 2.6 浮水印功能

- [x] 實作 `watermark.go`
  - [x] 圖片浮水印
  - [x] 位置控制 (9 種位置 + x,y offset)
  - [x] 透明度控制
  - [x] 比例縮放
- [x] 支援多個浮水印（可組合多個 watermark 濾鏡）

### 2.7 Phase 2 測試驗證

- [x] 測試各種濾鏡效果
- [x] 測試濾鏡組合 (連續套用)
- [x] 測試浮水印功能
- [x] 確認所有單元測試通過

---

## Phase 3: 安全與儲存

### 3.1 HMAC 安全機制

- [x] 建立 `internal/security/hmac.go`
- [x] 實作 HMAC-SHA256 簽名生成
- [x] 實作 Base64 URL-safe 編碼
- [x] 實作簽名驗證邏輯
- [x] 建立 `internal/api/middleware.go` 安全中介層
- [x] 處理 `/unsafe/` 路徑 (開發模式)
- [x] 撰寫安全機制單元測試

### 3.2 URL 簽名工具庫

- [x] 建立 `internal/security/url_signer.go`
- [x] 實作 SignURL 方法
- [x] 實作 VerifyURL 方法
- [x] 建立 CLI 簽名工具
- [x] 撰寫使用文件

### 3.3 來源白名單

- [x] 實作 allowed_sources 設定
- [x] 支援萬用字元 (`*.example.com`)
- [x] 實作來源驗證中介層
- [x] 撰寫白名單單元測試

### 3.4 AWS S3 儲存

- [x] 安裝 AWS SDK v2
- [x] 實作 `internal/storage/s3.go`
  - [x] Get 方法
  - [x] Put 方法
  - [x] Exists 方法
  - [x] Delete 方法
- [x] 支援認證設定
- [x] 支援區域設定
- [x] 撰寫 S3 儲存測試 (mock)

### 3.5 混合儲存模式

- [x] 實作 `internal/storage/mixed.go`
- [x] 支援原始檔/結果檔分離儲存
- [x] 實作儲存路由邏輯
- [x] 撰寫混合儲存測試

- [x] 實作 `internal/fx/storage.go`
- [x] 註冊 Storage Module 到 `cmd/server/main.go`
- [x] 注入 Storage 到 `internal/service/image_service.go`

### 3.6 無儲存模式

- [x] 實作 `internal/storage/no_storage.go`
- [x] 用於測試/基準測試

### 3.7 Phase 3 測試驗證

- [ ] 測試 HMAC 簽名驗證
- [ ] 測試 unsafe 模式
- [ ] 測試 S3 儲存
- [ ] 測試混合儲存模式
- [ ] 測試來源白名單

---

## Phase 4: 效能優化

### 4.1 Redis 快取

- [ ] 安裝 go-redis
- [ ] 建立 `internal/cache/interface.go`
- [ ] 實作 `internal/cache/redis.go`
  - [ ] Get 方法
  - [ ] Set 方法 (含 TTL)
  - [ ] Delete 方法
  - [ ] Exists 方法
- [ ] 實作快取鍵生成策略
- [ ] 整合快取到處理流程
- [ ] 撰寫 Redis 快取測試

### 4.2 記憶體快取

- [ ] 實作 `internal/cache/memory.go`
- [ ] 使用 LRU 策略
- [ ] 設定最大記憶體限制
- [ ] 實作 TTL 過期機制

### 4.3 Worker Pool

- [ ] 實作 Worker Pool
- [ ] 限制同時處理數量
- [ ] 實作任務佇列
- [ ] 支援優雅關閉

### 4.4 串流處理

- [ ] 實作大圖片串流讀取
- [ ] 實作串流寫入
- [ ] 減少記憶體佔用

### 4.5 Prometheus 監控

- [ ] 安裝 prometheus client
- [ ] 建立 `/metrics` 端點
- [ ] 實作處理時間指標
- [ ] 實作請求計數指標
- [ ] 實作錯誤率指標
- [ ] 實作快取命中率指標

### 4.6 Phase 4 測試驗證

- [ ] 壓力測試 (wrk/ab)
- [ ] 記憶體使用測試
- [ ] 快取效能測試
- [ ] 監控指標驗證

---

## Phase 5: 進階功能與新格式

### 5.1 AVIF 格式支援

- [ ] 安裝 `github.com/gen2brain/avif`
- [ ] 實作 AVIF 解碼
- [ ] 實作 AVIF 編碼
- [ ] 支援品質控制
- [ ] 撰寫 AVIF 測試

### 5.2 JPEG XL 格式支援

- [ ] 安裝 `github.com/ArtificialLegacy/go-jxl`
- [ ] 實作 JPEG XL 解碼
- [ ] 實作 JPEG XL 編碼
- [ ] 支援無損轉換 JPEG
- [ ] 撰寫 JPEG XL 測試

### 5.3 HEIC 格式支援

- [ ] 安裝 `github.com/jdeng/goheif`
- [ ] 實作 HEIC 解碼
- [ ] 轉換為其他格式輸出
- [ ] 撰寫 HEIC 測試

### 5.4 SVG 渲染

- [ ] 安裝 `github.com/srwiley/oksvg`
- [ ] 實作 SVG 解析
- [ ] 實作 SVG → 點陣圖渲染
- [ ] 支援自訂輸出尺寸
- [ ] 撰寫 SVG 測試

### 5.5 自動格式選擇

- [ ] 解析 Accept header
- [ ] 根據瀏覽器支援選擇最佳格式
- [ ] 實作格式優先級設定
- [ ] 支援強制格式參數覆蓋

### 5.6 智慧裁切

- [ ] 研究臉部偵測方案
- [ ] 整合臉部偵測庫
- [ ] 實作基於臉部的智慧裁切
- [ ] 實作 `smart` 參數支援

### 5.7 Phase 5 測試驗證

- [ ] 測試 AVIF 編解碼
- [ ] 測試 JPEG XL 編解碼
- [ ] 測試 HEIC 解碼
- [ ] 測試 SVG 渲染
- [ ] 測試自動格式選擇
- [ ] 測試智慧裁切

---

## 部署與文件

### 文件撰寫 (docs/)

- [ ] 完善 README.md（含 Badges）
- [ ] 建立 `docs/architecture.md` 系統架構說明
- [ ] 建立 `docs/api.md` API 規格文件
- [ ] 建立 `docs/adr/` ADR 目錄（設計決策紀錄）
- [ ] 撰寫設定說明文件
- [ ] 撰寫部署指南

### Docker 部署

- [ ] 最終化 Dockerfile
- [ ] 建立 docker-compose.yaml
- [ ] 建立 docker-compose.prod.yaml

### Kubernetes 部署 - Kustomize

- [ ] 建立 `deploy/kustomize/base/` 目錄結構
- [ ] 建立 base deployment.yaml
- [ ] 建立 base service.yaml
- [ ] 建立 base configmap.yaml
- [ ] 建立 `deploy/kustomize/overlays/dev/` 開發環境
- [ ] 建立 `deploy/kustomize/overlays/staging/` 測試環境
- [ ] 建立 `deploy/kustomize/overlays/prod/` 生產環境
- [ ] 設定環境變數與 Secret 參照
- [ ] 建立 HPA (Horizontal Pod Autoscaler) 設定
- [ ] 建立 PDB (Pod Disruption Budget) 設定

### Kubernetes 部署 - Helm Chart

- [ ] 建立 `charts/images-filters/` Helm chart 目錄
- [ ] 建立 Chart.yaml
- [ ] 建立 values.yaml (預設值)
- [ ] 建立 values-dev.yaml
- [ ] 建立 values-staging.yaml
- [ ] 建立 values-prod.yaml
- [ ] 建立 templates/deployment.yaml
- [ ] 建立 templates/service.yaml
- [ ] 建立 templates/configmap.yaml
- [ ] 建立 templates/secret.yaml
- [ ] 建立 templates/ingress.yaml
- [ ] 建立 templates/hpa.yaml
- [ ] 建立 templates/serviceaccount.yaml
- [ ] 建立 templates/_helpers.tpl
- [ ] 建立 templates/NOTES.txt
- [ ] 撰寫 Helm chart README

---

## 備註

- ⭐ 標記為高優先級任務
- 🚧 標記為進行中任務
- ⚠️ 標記為有風險/阻塞任務
