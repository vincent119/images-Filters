# System Limitations

[繁體中文](TW/limitations.md)

## Image Formats

- **Input**: JPEG, PNG, WebP, GIF, AVIF, HEIC (if supported by lib).
- **Output**: JPEG, PNG, WebP.
- **Note**: Vector graphics (SVG) are rasterized upon loading.

### Size Limits

To ensure stability, the system enforces the following defaults (configurable):

- **Max Input Dimensions**: 5000 x 5000 pixels.
- **Max Output Dimensions**: Defined by `processing.max_width/height`.
- **Max File Size**: Memory dependent (e.g., 50MB source file limit recommended).

### Performance

- **Animated GIFs**: Processing large animated GIFs is CPU intensive and may be slow. Only the first frame might be processed in some filter operations unless specifically handled.
- **Concurrency**: Limited by the number of worker threads (`processing.workers`). Setting this too high on a single core machine will cause context switching overhead.

### Feature Constraints

- **Smart Crop**: Relies on entropy calculation, which might not always perfectly center the subject.
- **Filters**: Some complex filters (e.g., convolution) are expensive.
