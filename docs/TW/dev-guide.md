# Developer Guide

[English](../dev-guide.md)

## 前置需求

- **Go**: 1.25.5 或更新版本。
- **Docker**: 用於執行依賴服務與容器建置。
- **Make**: 用於執行自動化建置指令。
- **GolangCI-Lint**: 用於程式碼品質檢查。

### 環境建置

1. **下載專案**:

   ```bash
   git clone https://github.com/vincent119/images-filters.git
   ```

2. **安裝模組**:

   ```bash
   go mod download
   ```

### 本地執行

```bash
# 使用預設設定執行
make run

# 開啟除錯日誌執行
LOG_LEVEL=debug make run
```

### 測試

```bash
# 執行單元測試
make test

# 執行 Race Detector 測試
go test -race ./...

# 查看覆蓋率
make coverage
```

### 程式碼檢查 (Linting)

我們使用 `golangci-lint` 並採用嚴格設定。

```bash
make lint
```

### 依賴注入

我們使用 `uber-go/fx` 進行依賴注入。當新增元件時：

1. 定義建構子 `NewComponent(...)`。
2. 在 `cmd/server/main.go` 中使用 `fx.Provide(NewComponent)` 進行註冊。
