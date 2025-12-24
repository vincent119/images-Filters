# System Architecture

[繁體中文](TW/architecture.md)

## High-Level Architecture

Images Filters follows a clean, layered architecture designed for high concurrency and extensibility.

```mermaid
graph TD
    User[Client / CDN] -->|HTTP Request| LB[Load Balancer]
    LB -->|Distribute| Server[Images Filters Server]

    subgraph "Images Filters Server"
        API[API Layer<br>(Gin Framework)]
        Middleware[Middleware<br>(Auth/Metrics/Log)]
        Service[Service Layer<br>(Business Logic)]
        Processor[Image Processor<br>(Libvips/Imaging)]
        Cache[Cache Layer<br>(Redis/Memory)]
        Loader[Loader Layer<br>(Source Fetcher)]
    end

    API --> Middleware
    Middleware --> Service
    Service --> Cache
    Cache -->|Miss| Loader
    Loader -->|Fetch| Source[Image Source<br>(S3/Local)]
    Service -->|Raw Image| Processor
    Processor -->|Processed Image| Service
    Service -->|Save| Cache
    Service -->|Response| API
```

### Components

#### 1. API Layer (`internal/api`)

- Handles HTTP requests and responses using Gin framework.
- Validates request parameters and payload.
- Entry point for all external traffic.

#### 2. Service Layer (`internal/service`)

- Orchestrates the image processing workflow.
- Coordinates between Cache, Loader, and Processor.
- Implements core business logic like "check cache -> load image -> process -> save cache".

#### 3. Image Processor (`internal/processor`)

- The core engine for image manipulation.
- Wraps libraries like `disintegration/imaging` or `libvips`.
- Handles Resize, Crop, Filter application, and Format conversion.

#### 4. Loader Layer (`internal/loader`)

- Responsible for fetching original images from various sources.
- Supports multiple backends: Local Filesystem, AWS S3, HTTP Remote.

#### 5. Cache Layer (`internal/cache`)

- Reduces processing load by storing processed results.
- Supports Multi-level caching: In-Memory (Ristretto) and Distributed (Redis).

#### 6. Security Layer (`internal/security`)

- Validates HMAC signatures to secure image URLs.
- Prevents unauthorized resource usage and DoS attacks.

### Data Flow

1. **Request Ingestion**: Request arrives at `/metrics` path.
2. **Security Check**: Middleware verifies the signature (if enabled).
3. **Cache Lookup**: Service checks if the processed image exists in Cache.
   - **Hit**: Returns cached image immediately.
4. **Image Loading**: If cache miss, Loader fetches the original image from Source.
5. **Processing**: Processor decodes the image, applies operations (Resize, Filters), and encodes to the target format.
6. **Cache Storage**: The result is stored in Cache for future requests.
7. **Response**: The processed image is streamed back to the Client.
