# Contributing Guide

[English](../contributing.md)

## 工作流程

1. **Fork** 此專案庫。
2. 建立 **功能分支** (`git checkout -b feat/new-filter`)。
3. 提交您的變更。
4. 推送至分支。
5. 開啟 **Pull Request**。

### 程式風格

- 遵循標準 **Go Code Review Comments**。
- 確保通過 `make lint` 檢查。
- 註解建議使用 **繁體中文** (針對文件說明) 或 英文 (針對程式邏輯)。
  *註：專案偏好文件使用繁體中文。*

### 提交訊息 (Commit Messages)

我們遵循 **Conventional Commits** 規範。

格式: `<type>(<scope>): <subject>`

**類型 (Types):**

- `feat`: 新增功能
- `fix`: 修復錯誤
- `docs`: 文件變更
- `chore`: 建置過程或輔助工具變更
- `refactor`: 重構 (既不修復錯誤也不新增功能的程式碼變更)
- `test`: 新增缺少的測試

**範例:**
`feat(processor): 新增 avif 格式支援`
