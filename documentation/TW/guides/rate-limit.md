# Rate Limiting

[English](../rate-limit.md)

## 策略

Images Filters 目前在預設的中介軟體堆疊中，尚未強制執行內建的應用程式層級限流 (Token Bucket/Leaky Bucket)。

**建議:**
限流應在 **基礎設施層** (Load Balancer, Ingress Controller, 或 API Gateway) 處理。

### 基礎設施設定

#### Nginx Ingress Controller

使用 Annotations 限制每個 IP 的 RPS：

```yaml
metadata:
  annotations:
    nginx.ingress.kubernetes.io/limit-rps: "100"
    nginx.ingress.kubernetes.io/limit-connections: "20"
```

#### 基於 Redis 的限流 (未來規劃)

我們計劃在 `internal/middleware/ratelimit` 中實作基於 Redis 的分散式限流功能。

### 防濫用機制

- **HMAC 簽名**: 防止濫用的主要防線。未簽名的請求（若已關閉不安全模式）將被立即拒絕（消耗極低）。
- **來源白名單**: 限制僅能從受信任的網域下載原始圖片。
