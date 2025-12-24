# API Specification

[繁體中文](TW/api.md)

## Overview

- **Base URL**: `http://<host>:<port>`
- **Content-Type**: Generally returns image binary data (`image/jpeg`, `image/png`, etc.) or JSON for errors.
- **Metrics Endpoint**: `/metrics` (Prometheus format)
- **Health Check**: `/healthz`

### Endpoints

#### 1. Process Image

Process an image with specified options and filters.

- **Development URL** (Unsafe):
  `GET /unsafe/<options>/<filters>/<image_path>`

- **Production URL** (Signed):
  `GET /<signature>/<options>/<filters>/<image_path>`

**Parameters:**

| Parameter | Description | Format / Example |
| ----------- | ------------- | ------------------ |
| `signature` | HMAC-SHA256 signature | Base64 encoded string |
| `options` | Processing options | `widthxheight` (e.g., `300x200`, `-300x200` for flip) |
| `filters` | Filter chain (Optional) | `filters:filter1(args):filter2(args)` |
| `image_path` | Source image path/URL | URL encoded path (e.g., `images/test.jpg` or `http%3A%2F%2F...`) |

**Supported Filters:**

- `blur(sigma)` : Apply Gaussian blur.
- `grayscale()` : Convert to grayscale.
- `brightness(factor)` : Adjust brightness (-100 to 100).
- `contrast(factor)` : Adjust contrast (-100 to 100).
- `watermark(image_url,opacity,x,y)` : Add watermark.

**Response:**

- `200 OK`: Returns the processed image binary.
- `400 Bad Request`: Invalid parameters or signature.
- `404 Not Found`: Image source not found.
- `500 Internal Server Error`: Processing failed.

#### 2. Health Check

Check service health status.

- **URL**: `GET /healthz`
- **Response**:

  ```json
  {
    "status": "ok",
    "timestamp": "2024-03-20T10:00:00Z"
  }
  ```

#### 3. Metrics

Expose Prometheus metrics.

- **URL**: `GET /metrics`
- **Format**: Prometheus text format.

### Error Codes

Error responses are returned in JSON format (except for some 404s which might return standard server pages depending on config).

```json
{
  "error": "code",
  "message": "Human readable error message"
}
```

| Code | Message Example | Description |
| ------ | ----------------- | ------------- |
| `invalid_signature` | "HMAC signature mismatch" | The URL signature is invalid or expired. |
| `image_not_found` | "failed to fetch source image" | The requested image could not be found. |
| `invalid_params` | "invalid width parameter" | Creating processing plan failed due to bad inputs. |
| `processing_error` | "decode failed" | Internal error during image manipulation. |
| `rate_limit` | "too many requests" | Request rate limit exceeded. |
