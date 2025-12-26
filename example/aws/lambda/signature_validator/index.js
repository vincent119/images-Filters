'use strict';

const crypto = require('crypto');

/**
 * Lambda@Edge: HMAC 簽名驗證
 *
 * 功能：
 * 1. 完整 HMAC-SHA256 簽名驗證
 * 2. 支援 /unsafe/ 開發模式（可透過環境變數控制）
 * 3. 在邊緣拒絕無效請求，減少 Origin 負擔
 *
 * 部署位置：Viewer Request 或 Origin Request
 * 執行時間限制：5 秒（Viewer Request）/ 30 秒（Origin Request）
 * 區域：必須部署在 us-east-1
 */

// 安全金鑰（生產環境應使用 Secrets Manager 或 Parameter Store）
// Lambda@Edge 不支援環境變數，需在程式碼中設定或使用 Secrets Manager
const SECURITY_KEY = 'your-security-key-here';
const ALLOW_UNSAFE = false; // 生產環境應設為 false

/**
 * 生成 HMAC-SHA256 簽名
 */
function generateSignature(path, key) {
    const hmac = crypto.createHmac('sha256', key);
    hmac.update(path);
    return hmac.digest('base64')
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=+$/, '');
}

/**
 * 驗證簽名
 */
function verifySignature(uri, key) {
    const parts = uri.split('/').filter(p => p.length > 0);

    if (parts.length < 2) {
        return { valid: false, reason: 'URL too short' };
    }

    const providedSignature = parts[0];

    // 檢查是否為 unsafe 模式
    if (providedSignature === 'unsafe') {
        if (ALLOW_UNSAFE) {
            return { valid: true, unsafe: true };
        }
        return { valid: false, reason: 'Unsafe mode is disabled' };
    }

    // 取得簽名後的路徑
    const pathWithoutSignature = '/' + parts.slice(1).join('/');

    // 計算期望的簽名
    const expectedSignature = generateSignature(pathWithoutSignature, key);

    // 時序安全比對
    if (providedSignature.length !== expectedSignature.length) {
        return { valid: false, reason: 'Signature mismatch' };
    }

    let result = 0;
    for (let i = 0; i < providedSignature.length; i++) {
        result |= providedSignature.charCodeAt(i) ^ expectedSignature.charCodeAt(i);
    }

    if (result !== 0) {
        return { valid: false, reason: 'Signature mismatch' };
    }

    return { valid: true };
}

/**
 * Lambda Handler
 */
exports.handler = async (event) => {
    const request = event.Records[0].cf.request;
    const uri = request.uri;

    // 跳過健康檢查和指標端點
    if (uri === '/healthz' || uri === '/metrics') {
        return request;
    }

    // 跳過 Swagger 文件
    if (uri.startsWith('/swagger')) {
        return request;
    }

    // 驗證簽名
    const result = verifySignature(uri, SECURITY_KEY);

    if (!result.valid) {
        console.log(`Signature validation failed: ${result.reason}, URI: ${uri}`);

        return {
            status: '403',
            statusDescription: 'Forbidden',
            headers: {
                'content-type': [{ key: 'Content-Type', value: 'application/json' }],
                'cache-control': [{ key: 'Cache-Control', value: 'no-store' }]
            },
            body: JSON.stringify({
                error: 'INVALID_SIGNATURE',
                message: result.reason
            })
        };
    }

    // 如果是 unsafe 模式，添加 header 標記
    if (result.unsafe) {
        request.headers['x-unsafe-mode'] = [{ key: 'X-Unsafe-Mode', value: 'true' }];
    }

    return request;
};
