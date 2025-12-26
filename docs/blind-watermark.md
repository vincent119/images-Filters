# Blind Watermark

[繁體中文](TW/blind-watermark.md)

The **Blind Watermark** feature allows you to embed invisible copyright information into images and retrieve it later for verification. Unlike visible watermarks, blind watermarks do not affect the visual quality of the image and persist even after image compression, scaling (within limits), and cropping (partial resistance).

## How It Works

This project uses frequency domain embedding (DCT - Discrete Cosine Transform) logic:

1. **Embedding**:
    - The image is converted to the frequency domain using DCT.
    - The watermark text is converted into binary data.
    - The binary data is embedded into the mid-frequency coefficients of the image.
    - The image is converted back to the spatial domain (IDCT).

2. **Detection**:
    - The suspected image is converted to the frequency domain.
    - The binary data is extracted from the coefficients.
    - The text is reconstructed and compared with the expected watermark text.

## Configuration

Enable blind watermarking in `config.yaml`:

```yaml
# Blind Watermark Configuration
blind_watermark:
  enabled: true
  text: "COPYRIGHT"  # The text to embed
  strength: 5.0      # Embedding strength (higher = more robust but more visible noise)
```

## Usage

### 1. Automatic Embedding

When enabled, the blind watermark is automatically applied to all processed images unless specifically disabled.

Example Request:
`GET /<signature>/300x200/smart/image.jpg`

The returned image will contain the hidden text "COPYRIGHT".

### 2. Manual Detection via API

You can detect watermarks using the `/detect` API. This API requires the `Authorization` header with your `SECURITY_KEY`.

#### Detect by File Upload

```bash
curl -X POST -H "Authorization: Bearer <YOUR_SECURITY_KEY>" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@/path/to/suspected_image.jpg" \
  http://localhost:8080/detect
```

#### Detect by Storage Path

If the image is already stored in your storage backend (local or S3):

```bash
curl -X POST -H "Authorization: Bearer <YOUR_SECURITY_KEY>" \
  -d "path=uploads/2025/12/26/image.jpg" \
  http://localhost:8080/detect
```

**Response:**

```json
{
  "detected": true,
  "text": "COPYRIGHT",
  "confidence": 0.98
}
```

### 3. Image Upload with Watermarking

You can use the `/upload` API to upload an original image. The system will store the original (raw) but can serve watermarked versions upon request via the standard processing pipeline.

## Limitations

- **Compression**: Extremely high compression (e.g., JPEG quality < 50) may destroy the watermark.
- **Cropping**: While the watermark is embedded globally, cropping too small a portion might make recovery impossible.
- **Rotation**: Standard implementation is sensitive to rotation. Images must be corrected for orientation before detection.

## Robustness Testing

We have tested the watermark against:

- **Resize**: Resistant to scaling down to 50%.
- **JPEG Compression**: Resistant to quality 80+.
- **Format Conversion**: Resistant to PNG <-> JPEG conversion.
