# Security Design

[繁體中文](TW/security.md)

## Overview

Security is a primary concern for image processing services to prevent unauthorized resource usage and Denial of Service (DoS) attacks. Images Filters implements HMAC-SHA256 signature verification for all production requests.

### URL Signature (HMAC)

To prevent attackers from requesting arbitrary image sizes and exhausting server CPU/Memory, all URLs must be signed.

#### Algorithm

1. **Construct path**: The part of the URL containing options, filters, and image path.
   Example: `300x200/filters:blur(5)/image.jpg`
2. **Key**: The `SECURITY_KEY` configured in the server.
3. **Sign**: Calculate HMAC-SHA256 of the path using the key.
4. **Encode**: Base64 URL-safe encode the result.

#### Implementation Example (Go)

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
)

func SignURL(key string, path string) string {
    mac := hmac.New(sha256.New, []byte(key))
    mac.Write([]byte(path))
    signature := base64.URLEncoding.EncodeToString(mac.Sum(nil))
    return signature
}
```

### Access Control

- **Unsafe Path**: `/unsafe/...` is strictly for development. It MUST be disabled in production (`SECURITY_ALLOW_UNSAFE=false`).
- **Source Validation**: The service currently supports unrestricted source URLs. Future versions will include whitelist domains for `http` loader.

### DoS Protection

- **Resource Limits**: Restrict `max_width` and `max_height` in configuration to prevent processing extremely large images (pixel bombs).
- **Timeouts**: Strict timeouts on image fetching and processing to release worker threads.
