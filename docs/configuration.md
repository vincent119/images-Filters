# Configuration Guide

[繁體中文](TW/configuration.md)

## Overview

Images Filters uses a hierarchical configuration system controlled by `config/config.yaml`. Values can be overridden by environment variables.
The precedence is: Environment Variables > Config File > Default Values.

### Configuration File (`config.yaml`)

You can find a complete example in [`config/config.sample.yaml`](../config/config.sample.yaml).

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  max_request_size: 10485760 # 10MB

processing:
  default_quality: 85
  max_width: 4096
  max_height: 4096
  workers: 4
  default_format: "jpeg"

security:
  enabled: true
  security_key: "your-secret-key"
  allow_unsafe: false
  allowed_sources: []

storage:
  type: "local" # local, s3, mixed
  local:
    root_path: "./data/images"
  s3:
    bucket: "my-bucket"
    region: "us-east-1"
    access_key: ""
    secret_key: ""

cache:
  enabled: true
  type: "redis" # memory, redis
  memory:
    max_size: 536870912 # 512MB
    ttl: 3600
  redis:
    host: "localhost"
    port: 6379
    pool:
      size: 10
      timeout: 4

logging:
  level: "info" # debug, info, warn, error
  format: "json" # json, text, console
  output: "stdout" # stdout, file

metrics:
  enabled: true
  path: "/metrics"

swagger:
  enabled: true
  path: "/swagger"
```

### Environment Variables

All configuration keys map to environment variables. Arrays and nested objects use underscore `_` separators and the prefix `IMG_`.

| Variable | Description | Default |
| ---------- | ------------- | --------- |
| `IMG_SERVER_PORT` | Server listening port | `8080` |
| `IMG_LOGGING_LEVEL` | Log level | `info` |
| `IMG_CACHE_TYPE` | Cache backend | `memory` |
| `IMG_CACHE_REDIS_HOST` | Redis host | `localhost` |
| `IMG_SECURITY_ENABLED` | Enable HMAC check | `false` |
| `IMG_SECURITY_KEY` | Secret for HMAC | `""` |
| `IMG_STORAGE_TYPE` | Storage backend | `local` |
