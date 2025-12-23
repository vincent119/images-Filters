// Package security 提供安全驗證功能
package security

import (
	"net/url"
	"strings"
)

// SourceValidator 來源驗證器
type SourceValidator struct {
	allowedSources []string
	enabled        bool
}

// NewSourceValidator 建立新的來源驗證器
func NewSourceValidator(allowedSources []string) *SourceValidator {
	return &SourceValidator{
		allowedSources: allowedSources,
		enabled:        len(allowedSources) > 0,
	}
}

// IsAllowed 檢查來源是否允許
// 支援萬用字元格式：*.example.com
func (v *SourceValidator) IsAllowed(source string) bool {
	// 未啟用白名單時，允許所有來源
	if !v.enabled {
		return true
	}

	// 空來源不允許
	if source == "" {
		return false
	}

	// 解析 URL 取得主機名稱
	host := extractHost(source)
	if host == "" {
		return false
	}

	// 檢查是否在白名單中
	for _, pattern := range v.allowedSources {
		if matchPattern(pattern, host) {
			return true
		}
	}

	return false
}

// IsEnabled 檢查白名單是否啟用
func (v *SourceValidator) IsEnabled() bool {
	return v.enabled
}

// extractHost 從 URL 或主機名稱中提取主機
func extractHost(source string) string {
	// 如果已經是主機名稱（不含 scheme）
	if !strings.Contains(source, "://") {
		// 移除路徑
		if idx := strings.Index(source, "/"); idx != -1 {
			source = source[:idx]
		}
		// 移除埠號
		if idx := strings.Index(source, ":"); idx != -1 {
			source = source[:idx]
		}
		return strings.ToLower(source)
	}

	// 解析完整 URL
	u, err := url.Parse(source)
	if err != nil {
		return ""
	}

	host := u.Hostname()
	return strings.ToLower(host)
}

// matchPattern 檢查主機是否匹配模式
// 支援萬用字元：*.example.com 匹配 sub.example.com
func matchPattern(pattern, host string) bool {
	pattern = strings.ToLower(pattern)
	host = strings.ToLower(host)

	// 完全匹配
	if pattern == host {
		return true
	}

	// 萬用字元匹配：*.example.com
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // .example.com
		// 匹配子域名：sub.example.com
		if strings.HasSuffix(host, suffix) {
			// 確保前面還有內容
			prefix := host[:len(host)-len(suffix)]
			if prefix != "" && !strings.Contains(prefix, ".") {
				return true
			}
		}
		// 也匹配主域名：example.com
		if host == pattern[2:] {
			return true
		}
	}

	return false
}
