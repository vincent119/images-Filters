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
- [x] 撰寫 Service 層單元測試

### 1.10 API 處理器整合

- [x] 建立 `internal/api/handler.go`
- [x] Handler 依賴 Service 介面（依賴注入）
- [x] 設定正確的 Content-Type
- [x] 實作錯誤處理和 HTTP 狀態碼
- [x] 建立 `routes/routes.go` 路由定義

### 1.11 Phase 1 測試驗證

- [/] 啟動服務器進行手動測試
- [x] 測試基本 resize 功能
- [x] 測試 crop 功能
- [x] 測試 flip 功能
- [x] 測試不同圖片格式
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
- [x] 撰寫 DI 模組測試

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

- [x] 測試 HMAC 簽名驗證
- [x] 測試 unsafe 模式
- [x] 測試 S3 儲存
- [x] 測試混合儲存模式
- [x] 測試來源白名單

---

## Phase 4: 效能優化

### 4.1 Redis 快取

- [x] 安裝 go-redis
- [x] 建立 `internal/cache/interface.go`
- [x] 實作 `internal/cache/redis.go`
  - [x] Get 方法
  - [x] Set 方法 (含 TTL)
  - [x] Delete 方法
  - [x] Exists 方法
- [x] 實作快取鍵生成策略
- [x] 整合快取到處理流程
- [x] 撰寫 Redis 快取測試
- [x] connect pool
- [x] TLS connection
- [x] add redis Username if empty string use requirepass

### 4.2 記憶體快取

- [x] 實作 `internal/cache/memory.go`
- [x] 使用 LRU 策略
- [x] 設定最大記憶體限制
- [x] 實作 TTL 過期機制

### 4.3 Worker Pool

- [x] 實作 Worker Pool (Semaphore 模式)
- [x] 限制同時處理數量
- [x] 實作任務佇列 (Buffered Channel)
- [x] 支援優雅關閉

### 4.4 串流處理

- [x] 實作大圖片串流讀取 (Stream Reading) <!-- id: 4 -->
  - [x] 更新 `Storage` 介面 (`GetStream`, `PutStream`) <!-- id: 5 -->
  - [x] 更新 `Loader` 介面 (`LoadStream`) <!-- id: 6 -->
  - [x] 實作 `S3Storage`, `LocalStorage`, `HTTPLoader`, `FileLoader` 的串流方法 <!-- id: 7 -->
  - [x] 修改 `Processor` 支援 `io.Reader` 輸入 <!-- id: 8 -->
  - [x] 整合至 `ImageService` 流程 <!-- id: 9 -->
- [x] 壓力測試與基準測試 (Benchmark) <!-- id: 10 -->
  - [x] 比較串流前後的記憶體使用量 (減少約 25% 記憶體佔用) <!-- id: 11 -->

### 4.5 Prometheus 監控

- [x] 安裝 prometheus client
- [x] 建立 `/metrics` 端點
- [x] 實作處理時間指標

#### HTTP 入口層指標

- [x] 實作請求總數指標（依 method / route / status）
- [x] 實作請求處理時間指標（Histogram，P50/P95/P99）
- [x] 實作進行中請求數（inflight requests）
- [x] 實作請求大小指標（request bytes）
- [x] 實作回應大小指標（response bytes）
- [x] 實作錯誤率指標（4xx / 5xx 分類）

#### 圖片處理核心指標

- [x] 實作圖片處理總耗時指標
- [x] 拆分處理階段耗時（decode / transform / encode）
- [x] 實作圖片處理操作類型計數（resize / crop / flip / watermark / filter）
- [x] 實作圖片處理錯誤分類指標（decode_failed / unsupported / timeout / oom）
- [x] 實作輸入圖片尺寸分佈指標
- [x] 實作輸出圖片尺寸分佈指標

#### 快取（Cache）指標

- [x] 實作快取命中 / 未命中計數
- [x] 實作快取命中率指標
- [x] 實作快取讀取延遲指標
- [x] 實作快取寫入延遲指標
- [x] 實作快取淘汰（eviction）計數（若有 LRU / TTL）

#### 儲存後端（S3 / 本地 / 其他）

- [x] 實作儲存後端操作計數（get / put）
- [x] 實作儲存後端延遲指標
- [x] 實作儲存後端錯誤分類（timeout / not_found / permission）
- [x] 實作儲存後端重試次數指標

#### 安全與風控

- [x] 實作請求簽名驗證成功 / 失敗計數
- [x] 實作被拒絕請求原因指標（bad_signature / expired / rate_limited）
- [x] 實作流量限制（Rate Limit）觸發次數指標

#### 系統與效能觀測

- [x] 启用 Go runtime 預設指標（GC / goroutines / memory）
- [x] 實作圖片 buffer pool 使用率指標（如使用 sync.Pool）
- [x] 實作服務啟動時間指標（uptime）

#### 可觀測性整合

- [x] 設計統一的 metrics 命名規則（避免 label 爆炸）
- [x] 製作 Prometheus Recording Rules（P95 / 錯誤率）
- [x] 建立 Grafana Dashboard（HTTP / 圖片處理 / Cache / Storage）
- [x] 設定 Alert 規則（高錯誤率 / 高延遲 / Cache 命中率下降）

#### Grafana Dashboard JSON

- [x] 建立 Grafana Dashboard JSON
  - [x] 複製到 ./example/grafana-dashboard.json

#### Alert Manager rules

- [x] 建立 Alert Manager rules
  - [x] 複製到 ./example/alert_rules/alert_rules.yml

### 4.6 Phase 4 測試驗證

- [x] 壓力測試 (wrk/ab)
- [x] 記憶體使用測試
- [x] 快取效能測試

## Phase 5: 進階功能與新格式

### 5.1 AVIF 格式支援

- [x] 安裝 `github.com/gen2brain/avif`
- [x] 實作 AVIF 解碼
- [x] 實作 AVIF 編碼
- [x] 支援品質控制
- [x] 撰寫 AVIF 測試

### 5.2 JPEG XL 格式支援

- [x] 安裝 `github.com/gen2brain/jpegxl`
- [x] 實作 JPEG XL 解碼
- [x] 實作 JPEG XL 編碼
- [ ] 支援無損轉換 JPEG (目前僅支援 Pixel-based 編碼)
- [x] 撰寫 JPEG XL 測試

### 5.3 HEIC 格式支援

- [x] 安裝 `github.com/gen2brain/heic` (僅支援解碼)
- [x] 實作 HEIC 解碼 (透過 Import 註冊)
- [x] 轉換為其他格式輸出 (整合至 Process 流程)
- [x] 撰寫 HEIC 測試 (GetContentType)

### 5.4 SVG 渲染

- [x] 安裝 `github.com/srwiley/oksvg`
- [x] 實作 SVG 解析 (透過 oksvg)
- [x] 實作 SVG → 點陣圖渲染 (透過 rasterx)
- [x] 支援自訂輸出尺寸 (SetTarget / calculateDimensions)
- [x] 撰寫 SVG 測試 (渲染與縮放驗證)

### 5.5 自動格式選擇

- [x] 解析 Accept header (整合至 ParsedURL)
- [x] 根據瀏覽器支援選擇最佳格式 (negotiateFormat)
- [x] 實作格式優先級設定 (Filter > Accept > Ext > Default)
- [x] 支援強制格式參數覆蓋 (Filters check)

### 5.6 智慧裁切

- [x] 研究臉部偵測方案 (選擇 smartcrop)
- [x] 整合臉部偵測庫 (muesli/smartcrop)
- [x] 實作 Processor 智慧裁切邏輯
- [x] 撰寫智慧裁切單元測試
- [x] 實作基於臉部的智慧裁切 (使用 smartcrop 演算法)
- [x] 實作 `smart` 參數支援 (Processor 整合)

### 5.7 Phase 5 測試驗證

- [x] 測試 AVIF 編解碼
- [x] 測試 JPEG XL 編解碼
- [x] 測試 HEIC 解碼
- [x] 測試 SVG 渲染
- [x] 測試自動格式選擇
- [x] 測試智慧裁切

---

## 6 部署與文件

### 6.1 文件撰寫（docs/）

#### 以下文件分為英文版本（docs/）與繁體中文版本（docs/TW/），並有 README.md (EN) 與 README_TW.md (TW)

- [x] 完善 README.md (EN) 與 README_TW.md (TW)（含 Badges、專案簡介、快速開始、核心功能說明、測試覆蓋率）
- [x] 建立 docs/architecture.md & docs/TW/architecture.md 系統架構說明
- [x] 建立 docs/api.md & docs/TW/api.md API 規格文件
- [x] 建立 docs/adr/ & docs/TW/adr/ ADR 目錄
- [x] 建立 docs/adr/README.md & docs/TW/adr/README.md
- [x] 撰寫設定說明文件 (docs/configuration.md & docs/TW/configuration.md)
- [x] 撰寫部署指南 (docs/deployment.md & docs/TW/deployment.md)

- [x] 建立 docs/security.md & docs/TW/security.md
- [x] 建立 docs/image-pipeline.md & docs/TW/image-pipeline.md
- [x] 建立 docs/cache-strategy.md & docs/TW/cache-strategy.md
- [x] 建立 docs/observability.md & docs/TW/observability.md
- [x] 建立 docs/troubleshooting.md & docs/TW/troubleshooting.md
- [x] 建立 docs/limitations.md & docs/TW/limitations.md

- [x] 建立 docs/performance.md & docs/TW/performance.md
- [x] 建立 docs/rate-limit.md & docs/TW/rate-limit.md
- [x] 建立 docs/error-handling.md & docs/TW/error-handling.md
- [x] 建立 docs/versioning.md & docs/TW/versioning.md
- [x] 建立 docs/deprecation.md & docs/TW/deprecation.md
- [x] 建立 docs/extensibility.md & docs/TW/extensibility.md
- [x] 建立 docs/dev-guide.md & docs/TW/dev-guide.md
- [x] 建立 docs/contributing.md & docs/TW/contributing.md
- [x] 建立 docs/compliance.md & docs/TW/compliance.md
- [x] 建立 docs/monitoring.md & docs/TW/monitoring.md
- [x] 建立 docs/configuration.md & docs/TW/configuration.md
- [x] 建立 config/config.sample.yaml

### 6.2 Docker 部署

- [x] 最終化 Dockerfile
- [x] 建立 docker-compose.yaml
  - [x] to /deploy/docker-compose.yaml
- [x] 建立 .dockerignore

### 6.3 Kubernetes 部署 - Kustomize

- [x] 建立 `deploy/kustomize/base/` 目錄結構
- [x] 建立 base deployment.yaml
- [x] 建立 base service.yaml
- [x] 建立 base configmap.yaml
- [x] 建立 `deploy/kustomize/overlays/prod/` 生產環境
- [x] 設定環境變數與 Secret 參照
- [x] 建立 HPA (Horizontal Pod Autoscaler) 設定
- [x] 建立 PDB (Pod Disruption Budget) 設定

### 6.4 Kubernetes 部署 - Helm Chart

- [x] 建立 `deploy/helm/images-filters/` Helm chart 目錄
- [x] 建立 Chart.yaml
- [x] 建立 values.yaml (預設值)
- [x] 建立 values-prod.yaml
- [x] 建立 templates/deployment.yaml
- [x] 建立 templates/service.yaml
- [x] 建立 templates/configmap.yaml
- [x] 建立 templates/secret.yaml
- [x] 建立 templates/ingress.yaml
- [x] 建立 templates/hpa.yaml
- [x] 建立 templates/serviceaccount.yaml
- [x] 建立 templates/_helpers.tpl
- [x] 建立 templates/NOTES.txt
- [x] 撰寫 Helm chart README

---

---

## Phase 7: 圖片上傳 API

### 7.1 上傳功能實作

- [x] 擴充 Service 介面支援 Upload
- [x] 實作 Service Upload 邏輯 (包含簽名生成)
- [x] 實作上傳安全中介層 (Bearer Auth)
- [x] 實作 API Upload Handler (Multipart)
- [x] 註冊 `POST /upload` 路由
- [x] 撰寫 Upload 單元測試
- [x] 手動驗證 Upload 功能與 Signed URL

---

## Phase 8: 隱形浮水印

### 8.1 設定與自動化

- [x] 更新 Config 結構支援 `BlindWatermark` (`internal/config/config.go`)
- [x] 更新 `image_service.go` 自動套用浮水印邏輯

### 8.2 核心算法實作

- [x] 實作 DCT/IDCT 變換 (`internal/filter/blind_watermark.go`)
- [x] 實作文字轉二進位編碼邏輯
- [x] 實作頻域嵌入邏輯
- [x] 註冊 `blind_watermark` 濾鏡
- [x] 撰寫單元測試

### 8.3 浮水印檢測服務

- [x] 建立 `WatermarkService` 介面 (`internal/service/watermark_service.go`)
- [x] 實作 `DetectWatermark` 方法（從 io.Reader 檢測）
- [x] 實作浮水印提取與比對邏輯
- [x] 建立 `WatermarkHandler` (`internal/api/watermark_handler.go`)
- [x] 實作 `HandleDetect` API 端點
- [x] 註冊 `/detect` 路由（含認證中介層）
- [x] 實作 fx 依賴注入整合

### 8.4 支援路徑檢測

- [x] 修改 `WatermarkService` 介面支援 `DetectWatermarkFromPath` 方法
- [x] 實作從 Storage 讀取檔案進行檢測
- [x] 更新 `/detect` API 新增 `path` 參數支援
- [x] 更新 Swagger 註解
- [x] 撰寫 `watermark_service_test.go` 單元測試
- [x] 驗證所有測試通過

---

## Phase 9: CDN 邊緣處理整合

### 9.1 CloudFront Function

- [x] 建立 `example/aws/cloudfront_function/` 目錄
- [x] 實作 `url_validator.js` - URL 格式驗證
- [x] 撰寫部署說明 `README.md`

### 9.2 Lambda@Edge

- [x] 建立 `example/aws/lambda/signature_validator/` 目錄
- [x] 實作 `index.js` - 完整 HMAC 簽名驗證
- [x] 建立 `package.json`
- [x] 撰寫部署說明 `README.md`

### 9.3 Origin Group Failover (進階)

- [x] 建立 `example/aws/lambda/origin_failover/` 目錄
- [x] 實作 `origin_request.js` - S3 優先讀取
- [x] 實作 `origin_response.js` - S3 Miss 時 Failover 到 API Server
- [x] 撰寫部署說明與架構比較 `README.md`

### 9.4 文件更新

- [x] 更新 `IMPLEMENTATION_PLAN.md`
- [x] 更新 `TASK.md`

---

## 7 備註

- ⭐ 標記為高優先級任務
- 🚧 標記為進行中任務
- ⚠️ 標記為有風險/阻塞任務
