# Deployment Guide

[English](../deployment.md)

## 1. 本地開發 (Local Development)

在您的機器上直接運行服務。

**前置需求:**

- Go 1.25.5+
- Redis (選用)

```bash
# 啟動 Redis (若需要)
docker run -d -p 6379:6379 redis:alpine

# 啟動伺服器
make run
```

### 2. Docker 部署

使用官方 Docker 映像檔進行部署。

```bash
# 建置映像
docker build -t images-filters .

# 執行容器
docker run -d \
  -p 8080:8080 \
  -e CACHE_TYPE=memory \
  -e SECURITY_ENABLED=false \
  images-filters
```

**Docker Compose:**
使用 `deploy/docker-compose.yaml` 啟動完整服務棧（應用 + Redis + Prometheus + Grafana）。

```bash
docker-compose -f deploy/docker-compose.yaml up -d
```

### 3. Kubernetes 部署

#### 使用 Helm

```bash
helm install images-filters ./deploy/helm/images-filters
```

#### 使用 Kustomize

```bash
# 部署至生產環境 Namespace
kubectl apply -k deploy/kustomize/overlays/prod
```

### 生產環境檢查清單 (Production Checklist)

- [ ] 設定 `SERVER_MODE=release`
- [ ] 啟用 `SECURITY_ENABLED=true` 並設定強密碼 `SECURITY_KEY`
- [ ] 使用 Redis 作為快取 (`CACHE_TYPE=redis`)
- [ ] 停用不安全路徑 (`SECURITY_ALLOW_UNSAFE=false`)
- [ ] 在 Kubernetes 中設定資源限制 (CPU/Memory)
- [ ] 設定水平自動擴縮 (HPA)
