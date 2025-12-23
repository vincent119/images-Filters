package security

import (
	"testing"
)

func TestNewSourceValidator(t *testing.T) {
	// 空白名單
	v1 := NewSourceValidator(nil)
	if v1.IsEnabled() {
		t.Error("Empty whitelist should not be enabled")
	}

	// 有白名單
	v2 := NewSourceValidator([]string{"example.com"})
	if !v2.IsEnabled() {
		t.Error("Non-empty whitelist should be enabled")
	}
}

func TestSourceValidator_IsAllowed_Disabled(t *testing.T) {
	v := NewSourceValidator(nil)

	// 未啟用時，所有來源都允許
	tests := []string{
		"example.com",
		"https://example.com/image.jpg",
		"http://any.domain.com",
	}

	for _, source := range tests {
		if !v.IsAllowed(source) {
			t.Errorf("Disabled validator should allow %s", source)
		}
	}
}

func TestSourceValidator_IsAllowed_ExactMatch(t *testing.T) {
	v := NewSourceValidator([]string{"example.com", "images.cdn.com"})

	tests := []struct {
		source   string
		expected bool
	}{
		{"example.com", true},
		{"https://example.com/image.jpg", true},
		{"http://example.com/path/to/image.png", true},
		{"images.cdn.com", true},
		{"https://images.cdn.com/img.webp", true},
		{"notallowed.com", false},
		{"sub.example.com", false}, // 完全匹配不包含子域名
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			if got := v.IsAllowed(tt.source); got != tt.expected {
				t.Errorf("IsAllowed(%s) = %v; want %v", tt.source, got, tt.expected)
			}
		})
	}
}

func TestSourceValidator_IsAllowed_Wildcard(t *testing.T) {
	v := NewSourceValidator([]string{"*.example.com", "*.cdn.net"})

	tests := []struct {
		source   string
		expected bool
	}{
		// 萬用字元匹配子域名
		{"sub.example.com", true},
		{"https://images.example.com/img.jpg", true},
		{"api.cdn.net", true},
		{"static.cdn.net", true},

		// 萬用字元也匹配主域名
		{"example.com", true},
		{"cdn.net", true},

		// 不匹配多層子域名（只匹配一層）
		{"a.b.example.com", false},

		// 不匹配其他域名
		{"notexample.com", false},
		{"exampleXcom", false},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			if got := v.IsAllowed(tt.source); got != tt.expected {
				t.Errorf("IsAllowed(%s) = %v; want %v", tt.source, got, tt.expected)
			}
		})
	}
}

func TestSourceValidator_IsAllowed_MixedPatterns(t *testing.T) {
	v := NewSourceValidator([]string{
		"static.example.com",
		"*.cdn.example.com",
		"trusted.net",
	})

	tests := []struct {
		source   string
		expected bool
	}{
		{"static.example.com", true},
		{"images.cdn.example.com", true},
		{"cdn.example.com", true},
		{"trusted.net", true},
		{"example.com", false},
		{"untrusted.net", false},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			if got := v.IsAllowed(tt.source); got != tt.expected {
				t.Errorf("IsAllowed(%s) = %v; want %v", tt.source, got, tt.expected)
			}
		})
	}
}

func TestSourceValidator_IsAllowed_EmptySource(t *testing.T) {
	v := NewSourceValidator([]string{"example.com"})

	if v.IsAllowed("") {
		t.Error("Empty source should not be allowed")
	}
}

func TestExtractHost(t *testing.T) {
	tests := []struct {
		source   string
		expected string
	}{
		{"example.com", "example.com"},
		{"example.com:8080", "example.com"},
		{"example.com/path/to/image.jpg", "example.com"},
		{"https://example.com", "example.com"},
		{"https://example.com:443/path", "example.com"},
		{"http://SUB.Example.COM/image.jpg", "sub.example.com"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			if got := extractHost(tt.source); got != tt.expected {
				t.Errorf("extractHost(%s) = %s; want %s", tt.source, got, tt.expected)
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern  string
		host     string
		expected bool
	}{
		// 完全匹配
		{"example.com", "example.com", true},
		{"example.com", "Example.COM", true},
		{"example.com", "other.com", false},

		// 萬用字元
		{"*.example.com", "sub.example.com", true},
		{"*.example.com", "example.com", true},
		{"*.example.com", "a.b.example.com", false},
		{"*.example.com", "notexample.com", false},
	}

	for _, tt := range tests {
		name := tt.pattern + "_" + tt.host
		t.Run(name, func(t *testing.T) {
			if got := matchPattern(tt.pattern, tt.host); got != tt.expected {
				t.Errorf("matchPattern(%s, %s) = %v; want %v",
					tt.pattern, tt.host, got, tt.expected)
			}
		})
	}
}
