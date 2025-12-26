'use strict';

/**
 * Lambda@Edge: Origin Response - S3 Miss 時 Failover 到 API Server
 *
 * 當 S3 回傳 403/404 時，改從 API Server 取得圖片並回傳
 *
 * 部署位置：Origin Response
 */

const https = require('https');

// 設定
const API_SERVER_DOMAIN = 'api.example.com';
const S3_DOMAIN = 'your-bucket.s3.ap-northeast-1.amazonaws.com';
const S3_BUCKET = 'your-bucket';
const S3_REGION = 'ap-northeast-1';

/**
 * 從 API Server 取得圖片
 */
function fetchFromApiServer(uri) {
    return new Promise((resolve, reject) => {
        const options = {
            hostname: API_SERVER_DOMAIN,
            port: 443,
            path: uri,
            method: 'GET',
            timeout: 30000
        };

        const req = https.request(options, (res) => {
            const chunks = [];

            res.on('data', (chunk) => {
                chunks.push(chunk);
            });

            res.on('end', () => {
                const body = Buffer.concat(chunks);
                resolve({
                    status: res.statusCode.toString(),
                    statusDescription: res.statusMessage,
                    headers: res.headers,
                    body: body.toString('base64'),
                    bodyEncoding: 'base64'
                });
            });
        });

        req.on('error', (err) => {
            reject(err);
        });

        req.on('timeout', () => {
            req.destroy();
            reject(new Error('Request timeout'));
        });

        req.end();
    });
}

/**
 * Lambda Handler
 */
exports.handler = async (event) => {
    const response = event.Records[0].cf.response;
    const request = event.Records[0].cf.request;

    // 如果 S3 回傳成功，直接返回
    if (response.status === '200') {
        return response;
    }

    // 如果 S3 回傳 403 或 404，表示檔案不存在，嘗試從 API Server 取得
    if (response.status === '403' || response.status === '404') {
        // 取得原始 URI（由 Origin Request Lambda 設定）
        const originalUri = request.headers['x-original-uri']
            ? request.headers['x-original-uri'][0].value
            : request.uri;

        console.log(`S3 miss for ${request.uri}, fetching from API Server: ${originalUri}`);

        try {
            const apiResponse = await fetchFromApiServer(originalUri);

            // 如果 API Server 成功回傳，可以選擇性地將結果存回 S3
            // 這裡簡化處理，只回傳結果
            // 實際生產環境可以用 S3 PutObject 存回快取

            // 建構 CloudFront 回應格式
            const cfResponse = {
                status: apiResponse.status,
                statusDescription: apiResponse.statusDescription,
                headers: {
                    'content-type': [{
                        key: 'Content-Type',
                        value: apiResponse.headers['content-type'] || 'application/octet-stream'
                    }],
                    'cache-control': [{
                        key: 'Cache-Control',
                        value: apiResponse.headers['cache-control'] || 'max-age=86400'
                    }]
                },
                body: apiResponse.body,
                bodyEncoding: apiResponse.bodyEncoding
            };

            return cfResponse;

        } catch (error) {
            console.error('Failed to fetch from API Server:', error);

            // 回傳原始的 S3 錯誤
            return response;
        }
    }

    // 其他狀態碼直接返回
    return response;
};
