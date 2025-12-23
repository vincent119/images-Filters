package security

import (
	"testing"
)

func TestNewSigner(t *testing.T) {
	signer := NewSigner("test-secret-key-1234")
	if signer == nil {
		t.Fatal("NewSigner should not return nil")
	}
	if len(signer.key) == 0 {
		t.Error("Signer key should not be empty")
	}
}

func TestSign(t *testing.T) {
	signer := NewSigner("my-secret-key-for-testing")

	// 測試基本簽名
	sig1 := signer.Sign("300x200/test.jpg")
	if sig1 == "" {
		t.Error("Sign should return non-empty signature")
	}

	// 簽名長度應為 44（Base64 編碼的 32 bytes）
	if len(sig1) != 44 {
		t.Errorf("Signature length = %d; want 44", len(sig1))
	}

	// 相同路徑應產生相同簽名
	sig2 := signer.Sign("300x200/test.jpg")
	if sig1 != sig2 {
		t.Error("Same path should produce same signature")
	}

	// 不同路徑應產生不同簽名
	sig3 := signer.Sign("400x300/test.jpg")
	if sig1 == sig3 {
		t.Error("Different paths should produce different signatures")
	}
}

func TestSign_NormalizePath(t *testing.T) {
	signer := NewSigner("my-secret-key-for-testing")

	// 有無開頭斜線應產生相同簽名
	sig1 := signer.Sign("/300x200/test.jpg")
	sig2 := signer.Sign("300x200/test.jpg")

	if sig1 != sig2 {
		t.Error("Leading slash should be normalized")
	}
}

func TestVerify(t *testing.T) {
	signer := NewSigner("my-secret-key-for-testing")
	path := "300x200/filters:blur(5)/test.jpg"

	// 產生正確簽名
	signature := signer.Sign(path)

	// 驗證正確簽名
	if !signer.Verify(signature, path) {
		t.Error("Verify should return true for valid signature")
	}

	// 驗證錯誤簽名
	if signer.Verify("invalid-signature", path) {
		t.Error("Verify should return false for invalid signature")
	}

	// 驗證被篡改的路徑
	if signer.Verify(signature, "400x300/test.jpg") {
		t.Error("Verify should return false for tampered path")
	}
}

func TestVerify_DifferentKey(t *testing.T) {
	signer1 := NewSigner("secret-key-one")
	signer2 := NewSigner("secret-key-two")

	path := "300x200/test.jpg"
	signature := signer1.Sign(path)

	// 不同金鑰應驗證失敗
	if signer2.Verify(signature, path) {
		t.Error("Different key should fail verification")
	}
}

func TestSignURL(t *testing.T) {
	signer := NewSigner("my-secret-key-for-testing")

	tests := []struct {
		name string
		path string
	}{
		{"基本路徑", "300x200/test.jpg"},
		{"帶斜線", "/300x200/test.jpg"},
		{"含濾鏡", "300x200/filters:blur(5)/test.jpg"},
		{"HTTP URL", "300x200/https://example.com/image.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signedURL := signer.SignURL(tt.path)

			// 應以斜線開頭
			if signedURL[0] != '/' {
				t.Error("Signed URL should start with /")
			}

			// 應包含原始路徑
			// 提取簽名後驗證
			sig, extractedPath, ok := ExtractSignatureAndPath(signedURL)
			if !ok {
				t.Error("Should be able to extract signature and path")
			}

			if !signer.Verify(sig, extractedPath) {
				t.Error("Extracted signature should be valid")
			}
		})
	}
}

func TestExtractSignatureAndPath(t *testing.T) {
	signer := NewSigner("my-secret-key-for-testing")
	originalPath := "300x200/test.jpg"
	signedURL := signer.SignURL(originalPath)

	sig, path, ok := ExtractSignatureAndPath(signedURL)
	if !ok {
		t.Fatal("ExtractSignatureAndPath should succeed")
	}

	if len(sig) != 44 {
		t.Errorf("Signature length = %d; want 44", len(sig))
	}

	if path != originalPath {
		t.Errorf("Path = %s; want %s", path, originalPath)
	}
}

func TestExtractSignatureAndPath_Invalid(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"無斜線", "noseparator"},
		{"短簽名", "short/300x200/test.jpg"},
		{"空路徑", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, ok := ExtractSignatureAndPath(tt.path)
			if ok {
				t.Error("Should return false for invalid path")
			}
		})
	}
}

func TestIsUnsafePath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/unsafe/300x200/test.jpg", true},
		{"unsafe/300x200/test.jpg", true},
		{"/unsafe/", true},
		{"/300x200/test.jpg", false},
		{"/signature/300x200/test.jpg", false},
		{"/unsafeX/300x200/test.jpg", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsUnsafePath(tt.path); got != tt.expected {
				t.Errorf("IsUnsafePath(%s) = %v; want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestGetPathWithoutUnsafe(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/unsafe/300x200/test.jpg", "300x200/test.jpg"},
		{"unsafe/300x200/test.jpg", "300x200/test.jpg"},
		{"/300x200/test.jpg", "300x200/test.jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := GetPathWithoutUnsafe(tt.input); got != tt.expected {
				t.Errorf("GetPathWithoutUnsafe(%s) = %s; want %s", tt.input, got, tt.expected)
			}
		})
	}
}
