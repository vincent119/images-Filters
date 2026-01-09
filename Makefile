.PHONY: all build build-signer run test clean lint fmt vet tidy swagger help sign

# 變數定義
APP_NAME := images-filters
SIGNER_NAME := signer
BUILD_DIR := bin
MAIN_FILE := ./cmd/server/main.go
SIGNER_FILE := ./cmd/signer/main.go

# Go 指令
GO := go
GOFMT := gofmt
GOVET := $(GO) vet
GOLINT := golangci-lint

# 預設目標
all: lint test build

## build: building application
build:
	@echo "==> building application..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)

## build-signer: building signer CLI tool
build-signer:
	@echo "==> building signer CLI tool..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(SIGNER_NAME) $(SIGNER_FILE)

## build-all: building all binaries
build-all: build build-signer

## run: running application
run:
	@echo "==> running server..."
	-$(GO) run $(MAIN_FILE)

## test: running all tests
test:
	@echo "==> running tests..."
	$(GO) test -v ./...

## test-cover: running tests and generating coverage report
test-cover:
	@echo "==> running tests and generating coverage report..."
	$(GO) test -cover -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

coverage:
	@echo "==> running tests and generating coverage report..."
	$(GO) test -cover -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

## lint: running golangci-lint
lint:
	@echo "==> running golangci-lint..."
	$(GOLINT) run

## fmt: formatting code
fmt:
	@echo "==> formatting code..."
	$(GOFMT) -w .

## vet: 執行 go vet
vet:
	@echo "==> running go vet..."
	$(GOVET) ./...

## tidy: 整理依賴
tidy:
	@echo "==>  Tidying dependencies......"
	$(GO) mod tidy

## swagger: 生成 Swagger 文檔
swagger:
	@echo "==> generating Swagger files..."
	swag init -g $(MAIN_FILE) -o ./docs/swagger

## clean: 清理建置產物
clean:
	@echo "==> cleaning build files..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

## docker-build: 建置 Docker 映像
docker-build:
	@echo "==> building Docker image..."
	docker build -t $(APP_NAME):latest .

## docker-run: 執行 Docker 容器
docker-run:
	@echo "==> running Docker container..."
	docker run -p 8080:8080 $(APP_NAME):latest

## sign: generate signed URL (usage: make sign PATH="300x200/test.jpg")
sign: build-signer
	@./$(BUILD_DIR)/$(SIGNER_NAME) sign -path "$(PATH)" -base "http://localhost:8080"

## verify: verify signed URL (usage: make verify URL="/signature/path")
verify: build-signer
	@./$(BUILD_DIR)/$(SIGNER_NAME) verify -url "$(URL)"

## help: 顯示幫助訊息
help:
	@echo "Available Make targets:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

