---
trigger: always_on
---

---
trigger: always_on
---

# Golang 開發規則與提示詞

## 專案結構規範

遵循標準 Go 專案佈局：
- 回覆我一率使用中文
- 回覆我開頭稱呼我”兄弟“
- cmd/：應用程式入口點（Composition Root）。
main.go 負責初始化設定（Viper）、日誌、資料庫、Redis，組裝依賴（Repository → Service → Handler），並啟動 HTTP Server、Scheduler 或 Worker。
- internal/api/：HTTP 處理器（Handler / Controller）。
負責解析請求（body / query / header）、輸入驗證、呼叫 Service（Usecase），並統一格式化回應（DTO / Response）。
- internal/service/：業務邏輯層（Usecase / Application Layer）。
實作核心業務流程與規則，協調多個 Repository 或外部服務；介面定義於 internal/service/interface.go，不依賴 HTTP 或框架。
- internal/repository/：資料存取層（Infrastructure Adapter）。
使用 Gorm / SQL / Redis 實作資料存取，僅負責 CRUD 與查詢轉換；介面定義於 internal/repository/interface.go，不包含業務邏輯。
- internal/model/：資料模型層（Domain / Data Model）。
包含 Gorm Entity、Query / Filter 結構、必要的 Value Object，僅描述資料結構，不負責流程控制。
- routes/：路由定義層。
定義 HTTP Endpoint、對應 Handler，並掛載 Middleware（Auth / CORS / Tracing / Metrics）。
- config/：設定管理。
使用 Viper 統一管理環境變數與設定檔（env / yaml / json），不包含業務邏輯。
- pkg/：共用工具與基礎設施。
放置 Logger、Database、Redis、第三方 SDK 封裝等可重用元件，可被 internal/ 使用，但不可反向依賴。
- docs/：設計與使用者文件。
包含系統架構說明、API 規格、ADR（設計決策紀錄）、操作與維運文件（不含 godoc 自動產生文件）。


### 資料流向

Request
→ Route
→ API Handler（負責請求解析與回應格式）
→ Service / Usecase（業務流程與規則）
→ Repository（資料存取介面）
→ Database / Cache / External Service

## 程式碼風格規範

### 命名規則
- 使用 **駝峰式命名法** (CamelCase)
- 匯出的名稱首字母大寫，私有名稱首字母小寫
- 介面名稱通常以 `-er` 結尾（如 `Reader`, `Writer`, `Handler`）
- 套件名稱使用小寫單字，避免底線和混合大小寫
- 常數使用全大寫加底線（如 `MAX_RETRY_COUNT`）

### 錯誤處理
```go
// ✅ 正確：明確處理錯誤
result, err := someFunction()
if err != nil {
    return fmt.Errorf("someFunction failed: %w", err)
}

// ❌ 錯誤：忽略錯誤
result, _ := someFunction()
```

### 函式設計
- 函式應該簡短且專注於單一任務
- 參數數量盡量控制在 3-4 個以內
- 使用具名回傳值提高可讀性
- Context 應該作為第一個參數

```go
// ✅ 正確
func ProcessData(ctx context.Context, data []byte, opts ...Option) (Result, error)

// ❌ 錯誤
func ProcessData(data []byte, timeout int, retries int, debug bool, ctx context.Context)
```


### 資料庫模型

- 模型應該嵌入 `internal/model/base.go` 中的 `BaseModel` 或 `BaseModelWithDelete`。
- 使用 Gorm 標籤定義欄位。
- 範例：

  ```go
  type User struct {
      model.BaseModel
      Username string `gorm:"column:username;type:varchar(100);not null"`
      // ...
  }
  ```

### 依賴注入

- 使用建構子注入來連接元件。
- Handler 依賴 Service 介面。
- Service 依賴 Repository 介面。
- 範例：

  ```go
  // internal/api/user_handler.go
  func NewUserHandler(userService service.UserService, ...) *UserHandler { ... }
  ```

## 常用指令

### 建置與執行
```bash
# 編譯專案
// turbo
go build -o bin/app ./cmd/...

# 執行程式
// turbo
go run ./cmd/main.go

# 交叉編譯
GOOS=linux GOARCH=amd64 go build -o bin/app-linux ./cmd/...
```

### 測試
```bash
# 執行所有測試
// turbo
go test ./...

# 執行測試並顯示覆蓋率
// turbo
go test -cover ./...

# 產生覆蓋率報告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 執行特定測試
go test -run TestFunctionName ./path/to/package
```

### 程式碼品質
```bash
# 格式化程式碼
// turbo
go fmt ./...

# 靜態分析
// turbo
go vet ./...

# 使用 golangci-lint（推薦）
// turbo
golangci-lint run

# 整理模組依賴
// turbo
go mod tidy
```

### 更新 Swagger 文檔

修改 API 處理器或註解後：

```bash
swag init -g cmd/main.go
```


## 依賴管理

```bash
# 初始化模組
go mod init module-name

# 新增依賴
go get github.com/package/name@version

# 更新所有依賴
go get -u ./...

# 清理未使用的依賴
// turbo
go mod tidy
```

## 常用套件推薦

### Web 框架
- `net/http` - 標準庫，適合簡單 API
- `github.com/gin-gonic/gin` - 高效能 Web 框架


### 資料庫
- `database/sql` - 標準 SQL 介面
- `github.com/jmoiron/sqlx` - SQL 擴充
- `gorm.io/gorm` - ORM 框架

### 工具類
- `github.com/spf13/cobra` - CLI 框架
- `github.com/spf13/viper` - 設定管理
- `github.com/vincent119/zlogger` - 高效能日誌
- `github.com/go-playground/validator` - 驗證
- `github.com/uber-go` - dependency injection
- `github.com/prometheus/client_golang`  - Prometheus metrics
- `github.com/redis` - Redis Client
- `github.com/vincent119/commons` - 工具庫
- swagger
- Badges

## 進階提示

### 並發模式
```go
// 使用 errgroup 管理 goroutine
g, ctx := errgroup.WithContext(ctx)
for _, item := range items {
    item := item // 重要：建立新變數
    g.Go(func() error {
        return processItem(ctx, item)
    })
}
if err := g.Wait(); err != nil {
    return err
}
```

### 資源管理
```go
// 使用 defer 確保資源釋放
file, err := os.Open("file.txt")
if err != nil {
    return err
}
defer file.Close()
```

### 介面設計原則
- 接受介面，回傳結構體
- 保持介面小巧（1-3 個方法）
- 在使用端定義介面，而非實作端

```go
// ✅ 好的介面設計
type Reader interface {
    Read(p []byte) (n int, err error)
}

// ❌ 過大的介面
type DoEverything interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    Close() error
    Seek(offset int64, whence int) (int64, error)
    // ... 更多方法
}
```

## Linter 設定建議

在專案根目錄建立 `.golangci.yml`：
```yaml
run:
  timeout: 5m

linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - gosimple
    - ineffassign
    - unused
    - misspell

linters-settings:
  errcheck:
    check-blank: true
```
