.PHONY: all build run test clean lint fmt vet tidy swagger help

# 變數定義
APP_NAME := images-filters
BUILD_DIR := bin
MAIN_FILE := ./cmd/server/main.go

# Go 指令
GO := go
GOFMT := gofmt
GOVET := $(GO) vet
GOLINT := golangci-lint

# 預設目標
all: lint test build

## build: 編譯應用程式
build:
	@echo "==> 編譯應用程式..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)

## run: 執行應用程式
run:
	@echo "==> 啟動服務器..."
	-$(GO) run $(MAIN_FILE)

## test: 執行所有測試
test:
	@echo "==> 執行測試..."
	$(GO) test -v ./...

## test-cover: 執行測試並顯示覆蓋率
test-cover:
	@echo "==> 執行測試並生成覆蓋率報告..."
	$(GO) test -cover -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## lint: 執行 golangci-lint
lint:
	@echo "==> 執行 Lint 檢查..."
	$(GOLINT) run

## fmt: 格式化程式碼
fmt:
	@echo "==> 格式化程式碼..."
	$(GOFMT) -w .

## vet: 執行 go vet
vet:
	@echo "==> 執行 go vet..."
	$(GOVET) ./...

## tidy: 整理依賴
tidy:
	@echo "==> 整理依賴..."
	$(GO) mod tidy

## swagger: 生成 Swagger 文檔
swagger:
	@echo "==> 生成 Swagger 文檔..."
	swag init -g $(MAIN_FILE) -o ./docs/swagger

## clean: 清理建置產物
clean:
	@echo "==> 清理建置產物..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

## docker-build: 建置 Docker 映像
docker-build:
	@echo "==> 建置 Docker 映像..."
	docker build -t $(APP_NAME):latest .

## docker-run: 執行 Docker 容器
docker-run:
	@echo "==> 啟動 Docker 容器..."
	docker run -p 8080:8080 $(APP_NAME):latest

## help: 顯示幫助訊息
help:
	@echo "可用的 Make 目標："
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
