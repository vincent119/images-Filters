/**
 * CloudFront Function: URL 驗證與正規化
 *
 * 功能：
 * 1. 快速驗證簽名格式（拒絕明顯無效請求）
 * 2. URL 正規化（統一格式）
 * 3. 支援 /unsafe/ 開發模式
 *
 * 部署位置：Viewer Request
 * 執行時間限制：1ms
 * 記憶體限制：2MB
 */

function handler(event) {
    var request = event.request;
    var uri = request.uri;

    // 跳過健康檢查和指標端點
    if (uri === '/healthz' || uri === '/metrics') {
        return request;
    }

    // 跳過 Swagger 文件
    if (uri.startsWith('/swagger')) {
        return request;
    }

    // 解析 URI 路徑
    var parts = uri.split('/').filter(function(p) { return p.length > 0; });

    if (parts.length < 2) {
        return {
            statusCode: 400,
            statusDescription: 'Bad Request',
            headers: {
                'content-type': { value: 'application/json' }
            },
            body: JSON.stringify({
                error: 'INVALID_URL',
                message: 'URL format is invalid'
            })
        };
    }

    var firstSegment = parts[0];

    // 允許 unsafe 模式（開發環境）
    if (firstSegment === 'unsafe') {
        // 添加標記 header，讓 Origin 知道這是 unsafe 請求
        request.headers['x-unsafe-mode'] = { value: 'true' };
        return request;
    }

    // 驗證簽名格式
    // HMAC-SHA256 + Base64 URL-safe 編碼後約 43-44 字元
    // 這裡只做格式驗證，完整驗證由 Origin 或 Lambda@Edge 執行
    if (firstSegment.length < 20 || firstSegment.length > 50) {
        return {
            statusCode: 403,
            statusDescription: 'Forbidden',
            headers: {
                'content-type': { value: 'application/json' }
            },
            body: JSON.stringify({
                error: 'INVALID_SIGNATURE',
                message: 'Signature format is invalid'
            })
        };
    }

    // 檢查簽名是否只包含 Base64 URL-safe 字元
    var validChars = /^[A-Za-z0-9_-]+$/;
    if (!validChars.test(firstSegment)) {
        return {
            statusCode: 403,
            statusDescription: 'Forbidden',
            headers: {
                'content-type': { value: 'application/json' }
            },
            body: JSON.stringify({
                error: 'INVALID_SIGNATURE',
                message: 'Signature contains invalid characters'
            })
        };
    }

    // URL 正規化：移除重複的斜線
    var normalizedUri = '/' + parts.join('/');
    if (normalizedUri !== uri) {
        request.uri = normalizedUri;
    }

    return request;
}
