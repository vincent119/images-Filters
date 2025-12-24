# Developer Guide

[繁體中文](TW/dev-guide.md)

## Prerequisites

- **Go**: 1.25.5 or later.
- **Docker**: For running dependencies and container builds.
- **Make**: For running build automation.
- **GolangCI-Lint**: For code quality checks.

### Environment Setup

1. **Clone the repo**:

   ```bash
   git clone https://github.com/vincent119/images-filters.git
   ```

2. **Install modules**:

   ```bash
   go mod download
   ```

### Running Locally

```bash
# Run with default config
make run

# Run with debug logs
LOG_LEVEL=debug make run
```

### Testing

```bash
# Run unit tests
make test

# Run tests with race detector
go test -race ./...

# View coverage
make coverage
```

### Linting

We use `golangci-lint` with strict settings.

```bash
make lint
```

### Dependency Injection

We use `uber-go/fx` for dependency injection. When adding a new component:

1. Define the constructor `NewComponent(...)`.
2. Register it in `cmd/server/main.go` using `fx.Provide(NewComponent)`.
