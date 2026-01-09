# Images Filters Image Processing Service

[![Go Version](https://img.shields.io/badge/Go-1.25.5-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://github.com/vincent119/images-filters/actions/workflows/go.yml/badge.svg)](https://github.com/vincent119/images-filters/actions/workflows/go.yml)
[![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/vincent119/cca471fced090cd840f0d85a5e876305/raw/images-filters-coverage.json)](https://github.com/vincent119/images-filters)
[![Go Report Card](https://goreportcard.com/badge/github.com/vincent119/images-filters)](https://goreportcard.com/report/github.com/vincent119/images-filters)

[繁體中文](README_TW.md)

A high-performance image processing server supporting real-time resizing, cropping, flipping, filters, and watermarking. Optimized for speed and extensibility.

## ✨ Features

- 🖼️ **Image Processing**: Real-time Resize, Crop, Flip, Rotate, Format Conversion.
- 🎨 **Filters**: Blur, Grayscale, Brightness, Contrast, Sharpen, and more.
- 💧 **Watermark**: Support visible image watermarks and **invisible blind watermarks**.
- 🔒 **Security**: HMAC-SHA256 URL signing to prevent tampering.
- 📦 **Multiple Storage**: Local filesystem, AWS S3, and Mixed mode (local cache + remote source).
- ⚡ **High Performance**: Built-in Redis cache, Worker Pool processing, and Go concurrency.
- 📊 **Observability**: Prometheus metrics, Grafana dashboards, and structured logging.
- 🐳 **Cloud Native**: Docker images, Helm charts, and Kustomize deployment ready.

## 🚀 Quick Start

### Installation

```bash
# Clone repository
git clone https://github.com/vincent119/images-filters.git
cd images-filters

# Install dependencies
go mod tidy

# Run server
make run
```

### Docker

```bash
# Run with Docker
docker run -p 8080:8080 vincent119/images-filters:latest
```

## 📖 Usage

### URL Format

```bash
http://<server>/<signature>/<options>/<filters>/<image_path>
```

### Examples

```bash
# Resize to 300x200 (Unsafe mode)
http://localhost:8080/unsafe/300x200/https%3A%2F%2Fexample.com%2Fimage.jpg

# Apply Grayscale filter
http://localhost:8080/unsafe/300x200/filters:grayscale()/https%3A%2F%2Fexample.com%2Fimage.jpg

# Signed URL (Production)
http://localhost:8080/H9a8s.../300x200/image.jpg
```

For more details, please refer to the [Documentation](documentation/EN/README.md).

---

## 📚 Documentation

### Core Docs

- [Architecture](documentation/EN/references/architecture.md)
- [API Specification](documentation/EN/references/api.md)
- [Security Design](documentation/EN/references/security.md)
- [Configuration](documentation/EN/references/configuration.md)
- [Blind Watermark](documentation/EN/guides/blind-watermark.md)

### Advanced Guides

- [Image Pipeline](documentation/EN/guides/image-pipeline.md)
- [Cache Strategy](documentation/EN/guides/cache-strategy.md)
- [Observability](documentation/EN/guides/observability.md)
- [Performance](documentation/EN/guides/performance.md)

### Ops & Dev

- [Deployment](documentation/EN/guides/deployment.md)
- [Troubleshooting](documentation/EN/guides/troubleshooting.md)
- [Developer Guide](documentation/EN/guides/dev-guide.md)
- [Contributing](documentation/EN/guides/contributing.md)

## 🛠️ Development

```bash
# Run tests
make test

# Lint code
make lint

# Generate Swagger
make swagger
```

## 📝 License

MIT License. See [LICENSE](./LICENSE).
