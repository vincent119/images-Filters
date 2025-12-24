# Image Processing Pipeline

[繁體中文](TW/image-pipeline.md)

## Overview

The image processing pipeline describes the lifecycle of a request from ingestion to response. It is designed to be efficient, failing fast and caching often.

### Workflow Steps

1. **Request Ingestion**
   - Gin router receives the request.
   - Example: `GET /signature/300x200/image.jpg`

2. **Validation & Parsing**
   - **Security Middleware**: Validates `signature` against the path.
   - **Parser**: Decodes `300x200` into `ProcessingOptions` struct.

3. **Cache Layer 1 (Read)**
   - Check if the key (hash of parameters) exists in **Redis** (or Memory).
   - **HIT**: Return cached image immediately.
   - **MISS**: Proceed to next step.

4. **Image Loading**
   - **Loader**: Fetches the original image based on protocol (`http://`, `s3://`, or local file).
   - Validates image format and magic bytes.

5. **Processing Core**
   - **Decode**: Convert raw bytes to Image object.
   - **Operations**:
     - **Resize**: Lanczos resampling.
     - **Smart Crop**: (Optional) Content-aware cropping.
     - **Filters**: Apply filter chain (e.g., Blur, Grayscale).
   - **Encode**: Convert back to bytes (JPEG/PNG/WebP).

6. **Cache Layer 2 (Write)**
   - Store the processed result in Cache with TTL.

7. **Response**
   - Set `Content-Type` and HTTP Headers.
   - Stream bytes to client.
