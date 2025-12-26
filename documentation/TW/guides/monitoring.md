# 監控指南 (Monitoring Guide)

[English](../monitoring.md)

## 概述

本指南專注於維運監控，包含如何設定 Prometheus 抓取、定義關鍵指標 (SLIs) 與告警規則。關於所有可用指標的列表，請參考 [可觀測性指南](observability.md)。

## Prometheus 設定

### 抓取設定 (Scrape Config)

請將以下作業設定加入您的 `prometheus.yml`：

```yaml
scrape_configs:
  - job_name: 'images-filters'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics
    # 若在 config.yaml 中啟用了 Basic Auth
    # basic_auth:
    #   username: your_username
    #   password: your_password
```

## 服務層級指標 (SLIs)

監控以下四個黃金訊號 (4 Golden Signals) 以確保服務健康：

### 1. 延遲 (Latency)

- **指標**: `imgfilter_http_request_duration_seconds_bucket`
- **目標**: 99% 的請求小於 500ms (P99)。
- **查詢**:

  ```promql
  histogram_quantile(0.99, sum(rate(imgfilter_http_request_duration_seconds_bucket[5m])) by (le))
  ```

### 2. 流量 (Traffic)

- **指標**: `imgfilter_http_requests_total`
- **目標**: 監控每秒請求數 (RPS) 趨勢。
- **查詢**:

  ```promql
  sum(rate(imgfilter_http_requests_total[5m]))
  ```

### 3. 錯誤率 (Errors/Availability)

- **指標**: `imgfilter_http_requests_total{status=~"5.."}`
- **目標**: 錯誤率 < 0.1%。
- **查詢**:

  ```promql
  sum(rate(imgfilter_http_requests_total{status=~"5.."}[5m]))
  /
  sum(rate(imgfilter_http_requests_total[5m]))
  ```

### 4. 飽和度 (Saturation)

- **指標**: 記憶體使用量、CPU 使用率、Goroutine 數量。
- **查詢 (Goroutines)**: `go_goroutines`

## 告警規則 (Alerting Rules)

建議的 Prometheus 告警規則：

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
          summary: "偵測到高錯誤率"
          description: "錯誤率已持續 5 分鐘高於 1%。"

      - alert: HighLatency
        expr: |
          histogram_quantile(0.99, sum(rate(imgfilter_http_request_duration_seconds_bucket[5m])) by (le)) > 1.0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "P99 延遲過高"
          description: "P99 延遲已持續 5 分鐘高於 1 秒。"
```

## Grafana 儀表板

範例 Grafana 儀表板 JSON 檔位於 `examples/grafana-dashboard.json`。該儀表板視覺化了上述的 SLIs 指標。
