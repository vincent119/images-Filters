# Observability Guide

[English](../observability.md)

## 概述

Images Filters 透過結構化日誌與 Prometheus 指標提供全面的可觀測性。

### Prometheus 指標

指標可於 `GET /metrics` 存取。

#### 關鍵指標

| 指標名稱 | 類型 | 說明 |
| ------------- | ------ | ------------- |
| `imgfilter_http_requests_total` | Counter | HTTP 請求總數 (依狀態、方法、路徑分類)。 |
| `imgfilter_http_request_duration_seconds` | Histogram | 請求延遲分佈。 |
| `imgfilter_processing_duration_seconds` | Histogram | 圖片處理耗時 (解碼、轉換、編碼)。 |
| `imgfilter_cache_ops_total` | Counter | 快取命中與未命中次數。 |
| `imgfilter_storage_ops_total` | Counter | 儲存後端讀取操作次數。 |

### Grafana 儀表板

範例儀表板提供於 `example/grafana-dashboard.json`。

**面板:**

- **流量 (Traffic)**: 每秒請求數 (RPS)、狀態碼分佈。
- **延遲 (Latency)**: P99, P95, 平均回應時間。
- **快取 (Cache)**: 命中率 %、操作次數。
- **資源 (Resources)**: CPU、記憶體、Goroutines 數量。

### 日誌 (Logging)

日誌採用 JSON 格式 (生產環境) 或 Console 格式 (開發環境)。

**日誌範例:**

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
