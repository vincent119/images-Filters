# 建置階段
FROM golang:1.23-bookworm AS builder

# 安裝必要的建置工具和 libwebp
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    ca-certificates \
    tzdata \
    libwebp-dev \
    && rm -rf /var/lib/apt/lists/*

# 設定工作目錄
WORKDIR /app

# 複製 go.mod 和 go.sum
COPY go.mod go.sum* ./

# 下載依賴
RUN go mod download

# 複製原始碼
COPY . .

# 編譯應用程式（啟用 CGO 支援 WebP）
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /app/bin/images-filters ./cmd/server

# 執行階段
FROM debian:bookworm-slim

ARG USER_NAME=appuser
ENV TZ=Asia/Taipei

# 安裝必要的執行時依賴
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    libwebp7 \
    wget \
    && rm -rf /var/lib/apt/lists/* \
    && ln -snf /usr/share/zoneinfo/$TZ /etc/localtime \
    && echo $TZ > /etc/timezone

# 建立非 root 使用者
RUN groupadd -r ${USER_NAME} \
    && useradd -r -g ${USER_NAME} ${USER_NAME} \
    && mkdir /app/config

# 設定工作目錄
WORKDIR /app

# 從建置階段複製執行檔
COPY --from=builder --chown=${USER_NAME}:${USER_NAME} /app/bin/images-filters .

# 複製設定檔（如果存在）
COPY --from=builder --chown=${USER_NAME}:${USER_NAME} /app/config/config.yaml* ./config/

# 切換到非 root 使用者
USER ${USER_NAME}

# 暴露埠號
EXPOSE 8080

# 健康檢查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

# 啟動應用程式
ENTRYPOINT ["./images-filters"]

