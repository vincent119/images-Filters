'use strict';

/**
 * Lambda@Edge: Origin Request - S3 優先 + API Server Failover
 *
 * 架構：
 * CloudFront → Lambda@Edge → 嘗試 S3 (已處理圖片)
 *                          → 若 S3 沒有 → 改打 API Server 處理
 *
 * 部署位置：Origin Request
 *
 * 優點：
 * - 已處理過的圖片直接從 S3 讀取，不經過 API Server
 * - API Server 只處理「首次」或「未快取」的請求
 * - 大幅降低 API Server 負載
 */

// 設定
const S3_DOMAIN = 'your-bucket.s3.ap-northeast-1.amazonaws.com';
const API_SERVER_DOMAIN = 'api.example.com';

// S3 中已處理圖片的存放路徑前綴
const S3_CACHE_PREFIX = 'processed/';

/**
 * 將 URL 轉換為 S3 快取路徑
 * 例如: /abc123/300x200/image.jpg → processed/abc123/300x200/image.jpg
 */
function getS3CachePath(uri) {
    // 移除開頭的斜線
    const path = uri.startsWith('/') ? uri.substring(1) : uri;
    return '/' + S3_CACHE_PREFIX + path;
}

/**
 * Lambda Handler
 */
exports.handler = async (event) => {
    const request = event.Records[0].cf.request;
    const uri = request.uri;

    // 跳過非圖片請求
    if (uri === '/healthz' || uri === '/metrics' || uri.startsWith('/swagger')) {
        // 這些請求直接打到 API Server
        request.origin = {
            custom: {
                domainName: API_SERVER_DOMAIN,
                port: 443,
                protocol: 'https',
                path: '',
                sslProtocols: ['TLSv1.2'],
                readTimeout: 30,
                keepaliveTimeout: 5
            }
        };
        request.headers['host'] = [{ key: 'host', value: API_SERVER_DOMAIN }];
        return request;
    }

    // 檢查是否為 unsafe 模式（開發用，永遠打 API Server）
    if (uri.startsWith('/unsafe/')) {
        request.origin = {
            custom: {
                domainName: API_SERVER_DOMAIN,
                port: 443,
                protocol: 'https',
                path: '',
                sslProtocols: ['TLSv1.2'],
                readTimeout: 30,
                keepaliveTimeout: 5
            }
        };
        request.headers['host'] = [{ key: 'host', value: API_SERVER_DOMAIN }];
        return request;
    }

    // 預設：嘗試從 S3 讀取已處理的圖片
    // 將 URI 轉換為 S3 快取路徑
    const s3CachePath = getS3CachePath(uri);

    request.origin = {
        s3: {
            domainName: S3_DOMAIN,
            region: 'ap-northeast-1',
            authMethod: 'origin-access-identity',
            path: '',
            customHeaders: {}
        }
    };

    // 更新 URI 為 S3 快取路徑
    request.uri = s3CachePath;
    request.headers['host'] = [{ key: 'host', value: S3_DOMAIN }];

    // 添加 header 標記原始 URI，供 Origin Response 使用
    request.headers['x-original-uri'] = [{ key: 'X-Original-Uri', value: uri }];

    return request;
};
