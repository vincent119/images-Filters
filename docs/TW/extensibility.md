# Extensibility Guide

[English](../extensibility.md)

## 新增濾鏡 (Filters)

濾鏡實作位於 `internal/processor`。新增濾鏡步驟：

1. **介面**: 確保您的圖片處理器實作了該操作（例如特定的對比度演算法）。
2. **註冊**: 在濾鏡鏈邏輯中註冊濾鏡名稱與參數解析器。
3. **文件**: 更新 `api.md` 加入新濾鏡名稱與用法。

### 自訂載入器 (Loaders)

若要支援從新來源獲取圖片 (例如 Google Cloud Storage, FTP)：

1. **實作介面**: 建立一個實作 `loader.ImageLoader` 的結構體。

   ```go
   type ImageLoader interface {
       Load(ctx context.Context, path string) ([]byte, error)
   }
   ```

2. **註冊**: 在 `main.go` 的載入器工廠中註冊該載入器。

3. **設定**: 在 `config.yaml` 加入必要的設定鍵值。
