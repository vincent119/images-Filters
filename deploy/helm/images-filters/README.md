# Images Filters Helm Chart

高效能圖片處理服務 Helm Chart。

## 安裝

```bash
# 使用預設值安裝
helm install images-filters ./deploy/helm/images-filters

# 使用生產環境設定
helm install images-filters ./deploy/helm/images-filters -f ./deploy/helm/images-filters/values-prod.yaml

# 指定命名空間
helm install images-filters ./deploy/helm/images-filters -n images-filters --create-namespace
```

## 設定參數

| 參數 | 說明 | 預設值 |
|------|------|--------|
| `replicaCount` | 副本數 | `2` |
| `image.repository` | 映像倉庫 | `images-filters` |
| `image.tag` | 映像標籤 | `""` (使用 appVersion) |
| `service.type` | Service 類型 | `ClusterIP` |
| `service.port` | Service 端口 | `80` |
| `ingress.enabled` | 啟用 Ingress | `false` |
| `autoscaling.enabled` | 啟用 HPA | `true` |
| `autoscaling.minReplicas` | 最小副本數 | `2` |
| `autoscaling.maxReplicas` | 最大副本數 | `10` |
| `config.cache.type` | 快取類型 | `memory` |
| `config.security.enabled` | 啟用安全驗證 | `false` |

## 升級

```bash
helm upgrade images-filters ./deploy/helm/images-filters
```

## 解除安裝

```bash
helm uninstall images-filters
```
