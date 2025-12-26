# Error Handling

[English](../error-handling.md)

## 回應結構

所有 API 錯誤皆回傳一致的 JSON 結構，並搭配適當的 HTTP 狀態碼 (`4xx` 或 `5xx`)。

```json
{
  "error": "error_code_string",
  "message": "人類可讀的描述",
  "request_id": "req-123456"
}
```

### HTTP 狀態碼

- **400 Bad Request**: 用戶端錯誤 (無效參數, 簽名錯誤)。
- **401 Unauthorized**: 缺少認證 (若適用)。
- **403 Forbidden**: 簽名有效但不允許存取該資源 (如網域不在白名單)。
- **404 Not Found**: 找不到原始圖片。
- **405 Method Not Allowed**: 使用 POST 而非 GET。
- **429 Too Many Requests**: 超出請求限制。
- **500 Internal Server Error**: 伺服器崩潰或未處理異常。
- **502 Bad Gateway**: 上游來源無法連接。
- **503 Service Unavailable**: 伺服器過載。

### 錯誤代碼列表

| 代碼 | 說明 |
| ------ | ------------- |
| `invalid_signature` | HMAC 簽名驗證失敗。 |
| `invalid_options` | 處理選項 (寬/高) 無效。 |
| `invalid_filter` | 濾鏡語法錯誤或未知濾鏡。 |
| `image_load_failed` | 下載/載入原始圖片失敗。 |
| `image_process_failed` | 圖片處理引擎錯誤。 |
| `remote_source_error` | 上游伺服器回傳錯誤。 |
