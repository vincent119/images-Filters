# 建置階段
FROM golang:1.23-alpine AS builder

# 安裝必要的建置工具
RUN apk add --no-cache git ca-certificates tzdata

# 設定工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum
COPY go.mod go.sum* ./

# 下載依賴
RUN go mod download

# 複製原始碼
COPY . .

# 編譯應用程式
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/images-filters ./cmd/server

# 執行階段
FROM alpine:3.19

# 安裝必要的執行時依賴
RUN apk add --no-cache ca-certificates tzdata

# 建立非 root 使用者
RUN adduser -D -g '' appuser

# 設定工作目錄
WORKDIR /app

# 從建置階段複製執行檔
COPY --from=builder /app/bin/images-filters .

# 複製設定檔（如果存在）
COPY --from=builder /app/config/config.yaml* ./config/

# 切換到非 root 使用者
USER appuser

# 暴露埠號
EXPOSE 8080

# 健康檢查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

# 啟動應用程式
ENTRYPOINT ["./images-filters"]
