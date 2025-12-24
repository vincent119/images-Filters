# Versioning Policy

[English](../versioning.md)

## API 版本控制

Images Filters 概念上採用 **URL 路徑版本控制**，儘管目前的實作直接暴露於根路徑。

**未來規劃:**

- V1: `http://host/v1/signature/...`
- V2: `http://host/v2/signature/...`

目前服務隱含運作為 **v1** 版本。任何針對 URL 結構或簽名演算法的破壞性變更，都將推進版本至 `v2`。

### 語意化版本 (Semantic Versioning)

應用程式執行檔與 Docker 映像檔遵循 [Semantic Versioning 2.0.0](https://semver.org/) 規範。

- **MAJOR (主版本)**: 不相容的 API 變更。
- **MINOR (次版本)**: 向下相容的功能新增 (新濾鏡、新載入器)。
- **PATCH (修訂版)**: 向下相容的錯誤修正。
