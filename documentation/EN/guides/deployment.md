# Deployment Guide

[繁體中文](TW/deployment.md)

## 1. Local Development

Run the service directly on your machine.

**Prerequisites:**

- Go 1.25.5+
- Redis (Optional)

```bash
# Start Redis (if needed)
docker run -d -p 6379:6379 redis:alpine

# Run Server
make run
```

### 2. Docker Deployment

Deploy using the official Docker image.

```bash
# Build Image
docker build -t images-filters .

# Run Container
docker run -d \
  -p 8080:8080 \
  -e CACHE_TYPE=memory \
  -e SECURITY_ENABLED=false \
  images-filters
```

**Docker Compose:**
Use `deploy/docker-compose.yaml` for a full stack (App + Redis + Prometheus + Grafana).

```bash
docker-compose -f deploy/docker-compose.yaml up -d
```

### 3. Kubernetes Deployment

#### Using Helm

```bash
helm install images-filters ./deploy/helm/images-filters
```

#### Using Kustomize

```bash
# Deploy to Production namespace
kubectl apply -k deploy/kustomize/overlays/prod
```

### Production Checklist

- [ ] Set `SERVER_MODE=release`
- [ ] Enable `SECURITY_ENABLED=true` and set a strong `SECURITY_KEY`
- [ ] Use Redis for caching (`CACHE_TYPE=redis`)
- [ ] Disable `ALLOW_UNSAFE` (`SECURITY_ALLOW_UNSAFE=false`)
- [ ] Configure Resource Limits (CPU/Memory) in Kubernetes
- [ ] Configure Horizontal Pod Autoscaler (HPA)
