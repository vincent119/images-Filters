# Images Filters Image Processing Service

[![Go Version](https://img.shields.io/badge/Go-1.25.5-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://github.com/vincent119/images-filters/actions/workflows/go.yml/badge.svg)](https://github.com/vincent119/images-filters/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/vincent119/images-filters/badge.svg?branch=main)](https://coveralls.io/github/vincent119/images-filters?branch=main)
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

For more details, please refer to the [Documentation](docs/README.md).

---

## 📚 Documentation

### Core Docs

- [Architecture](docs/architecture.md)
- [API Specification](docs/api.md)
- [Security Design](docs/security.md)
- [Configuration](docs/configuration.md)
- [Blind Watermark](docs/blind-watermark.md)

### Advanced Guides

- [Image Pipeline](docs/image-pipeline.md)
- [Cache Strategy](docs/cache-strategy.md)
- [Observability](docs/observability.md)
- [Performance](docs/performance.md)

### Ops & Dev

- [Deployment](docs/deployment.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Developer Guide](docs/dev-guide.md)
- [Contributing](docs/contributing.md)

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

MIT License
