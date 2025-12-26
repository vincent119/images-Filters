# Cache Strategy

[繁體中文](TW/cache-strategy.md)

## Strategy Overview

Caching is critical for performance. Images Filters employs a "Cache-Aside" strategy with optimizations for immutable content.

### Cache Key Design

The cache key is a unique identifier generated from the request parameters.

**Format:**
`images:processed:<hash>`

**Hash Construction:**
SHA256 of string: `options + filters + image_path`

Example: `300x200/filters:grayscale()/image.jpg` -> `SHA256(...)`

### Cache Tiers

1. **Browser Cache (Client)**
   - Controlled via HTTP Headers (`Cache-Control`, `ETag`).
   - Default: `public, max-age=31536000` (1 year).

2. **CDN (Edge)**
   - Recommended deployment architecture puts a CDN (Cloudflare/CloudFront) in front.
   - Offloads static asset delivery.

3. **Application Cache (Server)**
   - **Redis**: Recommended for production. Shared across instances.
   - **In-Memory**: Used for local development or single-instance deployments.

### Configuration

```yaml
cache:
  type: redis
  redis:
    ttl: 3600 # 1 hour default
    pool:
      size: 10
```

### Invalidation

Since image URLs are deterministic based on parameters:

- **Change Parameter**: Requests a new image logic, bypasses old cache.
- **Change Source Image**: If the source image changes but the name stays the same, the cache might serve stale content until TTL expires.
  - **Best Practice**: Use versioned filenames for source images (e.g., `image-v1.jpg`).
