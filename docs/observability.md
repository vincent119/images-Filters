# Observability Guide

[繁體中文](TW/observability.md)

## Overview

Images Filters provides comprehensive observability through structured logging and Prometheus metrics.

### Prometheus Metrics

Metrics are exposed at `GET /metrics`.

#### Key Metrics

| Metric Name | Type | Description |
| ------------- | ------ | ------------- |
| `imgfilter_http_requests_total` | Counter | Total HTTP requests by status, method, path. |
| `imgfilter_http_request_duration_seconds` | Histogram | Request latency distribution. |
| `imgfilter_processing_duration_seconds` | Histogram | Time spent in image processing (decode, transform, encode). |
| `imgfilter_cache_ops_total` | Counter | Cache hits and misses. |
| `imgfilter_storage_ops_total` | Counter | Storage backend read operations. |

### Grafana Dashboard

A sample dashboard is provided in `example/grafana-dashboard.json`.

**Panels:**

- **Traffic**: RPS, Status Codes.
- **Latency**: P99, P95, Avg response times.
- **Cache**: Hit Rate %, Operation counts.
- **Resources**: CPU, Memory, Goroutines.

### Logging

Logs are structured in JSON format (production) or Console format (development).

**Example Log:**

```json
{
  "level": "info",
  "ts": "2024-03-20T10:00:00.000Z",
  "caller": "api/middleware.go:45",
  "msg": "request completed",
  "method": "GET",
  "path": "/image.jpg",
  "status": 200,
  "latency": 0.123
}
```
