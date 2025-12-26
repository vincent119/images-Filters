# Error Handling

[繁體中文](TW/error-handling.md)

## Response Structure

All API errors return a consistent JSON structure with an appropriate HTTP status code (`4xx` or `5xx`).

```json
{
  "error": "error_code_string",
  "message": "Human readable description",
  "request_id": "req-123456"
}
```

### HTTP Status Codes

- **400 Bad Request**: Client side error (Invalid params, Bad Signature).
- **401 Unauthorized**: Missing authentication (if applicable).
- **403 Forbidden**: Valid signature but not allowed resource (e.g. domain not whitelisted).
- **404 Not Found**: Image source not found.
- **405 Method Not Allowed**: Using POST instead of GET.
- **429 Too Many Requests**: Rate limit exceeded.
- **500 Internal Server Error**: Server crash or unhandled exception.
- **502 Bad Gateway**: Upstream source unreachable.
- **503 Service Unavailable**: Server overloaded.

### Error Codes List

| Code | Description |
| ------ | ------------- |
| `invalid_signature` | HMAC signature verification failed. |
| `invalid_options` | Processing options (width/height) are invalid. |
| `invalid_filter` | Filter syntax error or unknown filter. |
| `image_load_failed` | Failed to download/load source image. |
| `image_process_failed` | Libvips/Imaging processing error. |
| `remote_source_error` | Upstream server returned an error. |
