# Rate Limiting

[繁體中文](TW/rate-limit.md)

## Strategy

Images Filters does not currently enforce built-in application-level rate limiting (Token Bucket/Leaky Bucket) in the default middleware stack.

**Recommendation:**
Rate limiting should be handled at the **Infrastructure Layer** (Load Balancer, Ingress Controller, or API Gateway).

### Infrastructure Configuration

#### Nginx Ingress Controller

Use annotations to limit RPS per IP:

```yaml
metadata:
  annotations:
    nginx.ingress.kubernetes.io/limit-rps: "100"
    nginx.ingress.kubernetes.io/limit-connections: "20"
```

#### Redis-based Limiting (Future Roadmap)

We plan to implement distributed rate limiting using Redis in `internal/middleware/ratelimit`.

### Abuse Prevention

- **HMAC Signatures**: The primary defense against abuse. Unsigned requests (if unsafe mode is off) are rejected immediately (Cost ~0).
- **Source Whitelist**: Restrict fetching images to trusted domains only.
