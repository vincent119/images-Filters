package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoad 測試設定載入
func TestLoad(t *testing.T) {
	// 建立臨時設定檔
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  host: "127.0.0.1"
  port: 9090
  read_timeout: 60s
  write_timeout: 60s

processing:
  default_quality: 90
  max_width: 2048
  max_height: 2048
  workers: 8
  default_format: "webp"

logging:
  level: "debug"
  format: "text"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("無法建立測試設定檔: %v", err)
	}

	// 載入設定
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("載入設定失敗: %v", err)
	}

	// 驗證設定值
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %s; want 127.0.0.1", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d; want 9090", cfg.Server.Port)
	}
	if cfg.Processing.DefaultQuality != 90 {
		t.Errorf("Processing.DefaultQuality = %d; want 90", cfg.Processing.DefaultQuality)
	}
	if cfg.Processing.DefaultFormat != "webp" {
		t.Errorf("Processing.DefaultFormat = %s; want webp", cfg.Processing.DefaultFormat)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %s; want debug", cfg.Logging.Level)
	}
}

// TestLoadDefaults 測試預設值載入
func TestLoadDefaults(t *testing.T) {
	// 建立空的臨時設定檔
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("無法建立測試設定檔: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("載入設定失敗: %v", err)
	}

	// 驗證預設值
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %s; want 0.0.0.0", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d; want 8080", cfg.Server.Port)
	}
	if cfg.Processing.DefaultQuality != 85 {
		t.Errorf("Processing.DefaultQuality = %d; want 85", cfg.Processing.DefaultQuality)
	}
	if cfg.Storage.Type != "local" {
		t.Errorf("Storage.Type = %s; want local", cfg.Storage.Type)
	}
}

// TestValidation 測試設定驗證
func TestValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		wantError bool
	}{
		{
			name: "valid config",
			config: `
server:
  port: 8080
processing:
  default_quality: 85
logging:
  level: "info"
`,
			wantError: false,
		},
		{
			name: "invalid port",
			config: `
server:
  port: 70000
`,
			wantError: true,
		},
		{
			name: "invalid quality",
			config: `
processing:
  default_quality: 150
`,
			wantError: true,
		},
		{
			name: "invalid log level",
			config: `
logging:
  level: "invalid"
`,
			wantError: true,
		},
		{
			name: "invalid storage type",
			config: `
storage:
  type: "invalid"
`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configPath, []byte(tt.config), 0644); err != nil {
				t.Fatalf("無法建立測試設定檔: %v", err)
			}

			_, err := Load(configPath)
			if (err != nil) != tt.wantError {
				t.Errorf("Load() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestGetAddress 測試取得服務器地址
func TestGetAddress(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
		},
	}

	expected := "127.0.0.1:8080"
	if addr := cfg.GetAddress(); addr != expected {
		t.Errorf("GetAddress() = %s; want %s", addr, expected)
	}
}
