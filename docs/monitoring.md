# Monitoring Guide

[繁體中文](TW/monitoring.md)

## Overview

This guide focuses on operational monitoring, setting up Prometheus scraping, and defining critical alerts for the Images Filters service. For a list of available metrics, please refer to the [Observability Guide](observability.md).

## Prometheus Setup

### Scrape Configuration

Add the following job to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'images-filters'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics
    # If Basic Auth is enabled in config.yaml
    # basic_auth:
    #   username: your_username
    #   password: your_password
```

## Service Level Indicators (SLIs)

Monitor these 4 Golden Signals to ensure service health:

### 1. Latency
- **Metric**: `imgfilter_http_request_duration_seconds_bucket`
- **Goal**: 99% of requests < 500ms (P99).
- **Query**:
  ```promql
  histogram_quantile(0.99, sum(rate(imgfilter_http_request_duration_seconds_bucket[5m])) by (le))
  ```

### 2. Traffic
- **Metric**: `imgfilter_http_requests_total`
- **Goal**: Monitor Request Per Second (RPS) trends.
- **Query**:
  ```promql
  sum(rate(imgfilter_http_requests_total[5m]))
  ```

### 3. Errors (Availability)
- **Metric**: `imgfilter_http_requests_total{status=~"5.."}`
- **Goal**: Error rate < 0.1%.
- **Query**:
  ```promql
  sum(rate(imgfilter_http_requests_total{status=~"5.."}[5m]))
  /
  sum(rate(imgfilter_http_requests_total[5m]))
  ```

### 4. Saturation
- **Metric**: Memory Usage, CPU Usage, Goroutines.
- **Query (Goroutines)**: `go_goroutines`

## Alerting Rules

Recommended Prometheus alerting rules:

```yaml
groups:
  - name: images-filters
    rules:
      - alert: HighErrorRate
        expr: |
          (sum(rate(imgfilter_http_requests_total{status=~"5.."}[5m]))
          /
          sum(rate(imgfilter_http_requests_total[5m]))) > 0.01
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is above 1% for 5 minutes."

      - alert: HighLatency
        expr: |
          histogram_quantile(0.99, sum(rate(imgfilter_http_request_duration_seconds_bucket[5m])) by (le)) > 1.0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High P99 latency"
          description: "P99 latency is above 1s for 5 minutes."
```

## Grafana Dashboard

A sample Grafana dashboard JSON is available in `examples/grafana-dashboard.json`. It visualizes the SLIs mentioned above.
