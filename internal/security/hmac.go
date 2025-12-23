// Package security 提供 HMAC 簽名驗證功能
// 用於保護圖片處理 API 的安全存取
package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// Signer HMAC 簽名器
type Signer struct {
	key []byte
}

// NewSigner 建立新的簽名器
func NewSigner(secretKey string) *Signer {
	return &Signer{
		key: []byte(secretKey),
	}
}

// Sign 對路徑進行 HMAC-SHA256 簽名
// 回傳 Base64 URL-safe 編碼的簽名
func (s *Signer) Sign(path string) string {
	// 正規化路徑：移除開頭的斜線
	path = strings.TrimPrefix(path, "/")

	// 計算 HMAC-SHA256
	h := hmac.New(sha256.New, s.key)
	h.Write([]byte(path))
	signature := h.Sum(nil)

	// Base64 URL-safe 編碼
	return base64.URLEncoding.EncodeToString(signature)
}

// Verify 驗證簽名是否有效
func (s *Signer) Verify(signature, path string) bool {
	// 正規化路徑
	path = strings.TrimPrefix(path, "/")

	// 計算預期的簽名
	expected := s.Sign(path)

	// 使用常數時間比較防止時間攻擊
	return hmac.Equal([]byte(signature), []byte(expected))
}

// SignURL 產生帶簽名的 URL 路徑
// 輸入：300x200/filters:blur(5)/test.jpg
// 輸出：/{signature}/300x200/filters:blur(5)/test.jpg
func (s *Signer) SignURL(path string) string {
	// 正規化路徑
	path = strings.TrimPrefix(path, "/")

	// 產生簽名
	signature := s.Sign(path)

	// 組合完整路徑
	return "/" + signature + "/" + path
}

// ExtractSignatureAndPath 從 URL 路徑提取簽名和實際路徑
// 輸入：/abc123xyz=/300x200/test.jpg
// 輸出：signature="abc123xyz=", path="300x200/test.jpg"
func ExtractSignatureAndPath(fullPath string) (signature, path string, ok bool) {
	// 移除開頭斜線
	fullPath = strings.TrimPrefix(fullPath, "/")

	// 找到第一個斜線位置
	idx := strings.Index(fullPath, "/")
	if idx == -1 {
		return "", "", false
	}

	signature = fullPath[:idx]
	path = fullPath[idx+1:]

	// 簽名應該是 Base64 URL-safe 編碼，長度固定為 44（SHA256 = 32 bytes -> Base64 = 44 chars）
	if len(signature) != 44 {
		return "", "", false
	}

	return signature, path, true
}

// IsUnsafePath 檢查是否為 unsafe 路徑
func IsUnsafePath(path string) bool {
	path = strings.TrimPrefix(path, "/")
	return strings.HasPrefix(path, "unsafe/")
}

// GetPathWithoutUnsafe 移除 unsafe 前綴
// 輸入：/unsafe/300x200/test.jpg
// 輸出：300x200/test.jpg
func GetPathWithoutUnsafe(path string) string {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, "unsafe/")
	return path
}
