# System Limitations

[English](../limitations.md)

## 圖片格式

- **輸入**: JPEG, PNG, WebP, GIF, AVIF, HEIC (視函式庫支援而定)。
- **輸出**: JPEG, PNG, WebP。
- **注意**: 向量圖 (SVG) 會在載入時被點陣化。

### 尺寸限制

為確保系統穩定，系統強制執行以下預設限制 (可設定):

- **最大輸入尺寸**: 5000 x 5000 像素。
- **最大輸出尺寸**: 由 `processing.max_width/height` 定義。
- **最大檔案大小**: 取決於記憶體 (建議限制來源檔案在 50MB 以內)。

### 效能

- **動態 GIF**: 處理大型動態 GIF 極為消耗 CPU 資源。除非特別處理，某些濾鏡操作可能僅處理第一幀。
- **並發數**: 受限於 Worker 執行緒數量 (`processing.workers`)。在單核機器上設定過高會導致 Context Switching 開銷。

### 功能限制

- **智慧裁切 (Smart Crop)**: 依賴熵值計算，可能無法總是完美地將主體置中。
- **濾鏡**: 某些複雜濾鏡 (如卷積運算) 計算成本較高。
