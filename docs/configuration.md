# Configuration Guide

[繁體中文](TW/configuration.md)

## Overview

Images Filters uses a hierarchical configuration system controlled by `config/config.yaml`. Values can be overridden by environment variables.
The precedence is: Environment Variables > Config File > Default Values.

### Configuration File (`config.yaml`)

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug" # debug, release, test

logging:
  level: "info" # debug, info, warn, error
  format: "json" # json, text, console
  output: "stdout" # stdout, file
  file_path: "./logs/app.log"

processing:
  default_quality: 80
  max_width: 5000
  max_height: 5000
  workers: 4 # Processing worker pool size

cache:
  type: "redis" # memory, redis
  memory:
    max_size: 536870912 # 512MB
    ttl: 3600
  redis:
    host: "localhost"
    port: 6379
    username: ""
    password: ""
    db: 0
    ttl: 3600
    pool:
      size: 10
      timeout: 4
    tls:
      enabled: false

security:
  enabled: true
  security_key: "your-secret-key"
  allow_unsafe: false
```

### Environment Variables

All configuration keys map to environment variables. Arrays and nested objects use underscore `_` separators.

| Variable | Description | Default |
| ---------- | ------------- | --------- |
| `SERVER_PORT` | Server listening port | `8080` |
| `LOG_LEVEL` | Log level | `info` |
| `CACHE_TYPE` | Cache backend | `memory` |
| `CACHE_REDIS_HOST` | Redis host | `localhost` |
| `SECURITY_ENABLED` | Enable HMAC check | `false` |
| `SECURITY_KEY` | Secret for HMAC | `""` |
