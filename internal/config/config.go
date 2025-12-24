// Package config 提供應用程式設定管理功能
// 使用 Viper 支援 YAML 設定檔和環境變數覆蓋
// 使用 go-playground/validator 進行結構體驗證
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config 應用程式完整設定結構
type Config struct {
	Server     ServerConfig     `mapstructure:"server" validate:"required"`
	Processing ProcessingConfig `mapstructure:"processing" validate:"required"`
	Security   SecurityConfig   `mapstructure:"security"`
	Storage    StorageConfig    `mapstructure:"storage" validate:"required"`
	Cache      CacheConfig      `mapstructure:"cache"`
	Logging    LoggingConfig    `mapstructure:"logging" validate:"required"`
	Metrics    MetricsConfig    `mapstructure:"metrics"`
	Swagger    SwaggerConfig    `mapstructure:"swagger"`
}

// SwaggerConfig Swagger UI 設定
type SwaggerConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Path     string `mapstructure:"path" validate:"omitempty,startswith=/"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// MetricsConfig Prometheus 指標設定
type MetricsConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Namespace string `mapstructure:"namespace" validate:"omitempty,alphanum"`
	Path      string `mapstructure:"path" validate:"omitempty,startswith=/"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
}

// ServerConfig HTTP 服務器設定
type ServerConfig struct {
	Host           string        `mapstructure:"host" validate:"required,ip|hostname"`
	Port           int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout" validate:"required,min=1s"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout" validate:"required,min=1s"`
	MaxRequestSize int64         `mapstructure:"max_request_size" validate:"required,min=1"`
}

// ProcessingConfig 圖片處理設定
type ProcessingConfig struct {
	DefaultQuality int    `mapstructure:"default_quality" validate:"required,min=1,max=100"`
	MaxWidth       int    `mapstructure:"max_width" validate:"required,min=1,max=16384"`
	MaxHeight      int    `mapstructure:"max_height" validate:"required,min=1,max=16384"`
	Workers        int    `mapstructure:"workers" validate:"required,min=1,max=128"`
	DefaultFormat  string `mapstructure:"default_format" validate:"required,oneof=jpeg jpg png gif webp avif jxl"`
}

// SecurityConfig 安全設定
type SecurityConfig struct {
	Enabled        bool     `mapstructure:"enabled"`
	SecurityKey    string   `mapstructure:"security_key" validate:"required_if=Enabled true,omitempty,min=16"`
	AllowUnsafe    bool     `mapstructure:"allow_unsafe"`
	AllowedSources []string `mapstructure:"allowed_sources"`
	MaxWidth       int      `mapstructure:"max_width" validate:"omitempty,min=1,max=16384"`
	MaxHeight      int      `mapstructure:"max_height" validate:"omitempty,min=1,max=16384"`
}

// StorageConfig 儲存設定
type StorageConfig struct {
	Type  string             `mapstructure:"type" validate:"required,oneof=local s3 no_storage mixed"`
	Local LocalStorageConfig `mapstructure:"local"`
	S3    S3StorageConfig    `mapstructure:"s3"`
	Mixed MixedStorageConfig `mapstructure:"mixed"`
}

// LocalStorageConfig 本地儲存設定
type LocalStorageConfig struct {
	RootPath string `mapstructure:"root_path" validate:"required_if=Type local"`
}

// S3StorageConfig AWS S3 儲存設定
type S3StorageConfig struct {
	Bucket    string `mapstructure:"bucket" validate:"required_if=Type s3"`
	Region    string `mapstructure:"region" validate:"required_if=Type s3"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Endpoint  string `mapstructure:"endpoint"`
}

// MixedStorageConfig 混合儲存設定
type MixedStorageConfig struct {
	SourceStorage string `mapstructure:"source_storage" validate:"required_if=Type mixed,omitempty,oneof=local s3"`
	ResultStorage string `mapstructure:"result_storage" validate:"required_if=Type mixed,omitempty,oneof=local s3"`
}

// CacheConfig 快取設定
type CacheConfig struct {
	Enabled bool              `mapstructure:"enabled"`
	Type    string            `mapstructure:"type" validate:"required_if=Enabled true,omitempty,oneof=redis memory"`
	Redis   RedisCacheConfig  `mapstructure:"redis"`
	Memory  MemoryCacheConfig `mapstructure:"memory"`
}

// RedisCacheConfig Redis 快取設定
type RedisCacheConfig struct {
	Host     string `mapstructure:"host" validate:"required_if=Type redis,omitempty,hostname|ip"`
	Port     int    `mapstructure:"port" validate:"required_if=Type redis,omitempty,min=1,max=65535"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db" validate:"omitempty,min=0,max=15"`
	TTL      int    `mapstructure:"ttl" validate:"omitempty,min=1"`
}

// MemoryCacheConfig 記憶體快取設定
type MemoryCacheConfig struct {
	MaxSize int64 `mapstructure:"max_size" validate:"required_if=Type memory,omitempty,min=1048576"`
	TTL     int   `mapstructure:"ttl" validate:"omitempty,min=1"`
}

// LoggingConfig 日誌設定
type LoggingConfig struct {
	Level    string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	Format   string `mapstructure:"format" validate:"required,oneof=json text console"`
	Output   string `mapstructure:"output" validate:"required,oneof=stdout file"`
	FilePath string `mapstructure:"file_path" validate:"required_if=Output file"`
}

// 全域驗證器實例
var validate *validator.Validate

func init() {
	validate = validator.New()

	// 註冊自訂驗證規則（如有需要）
	// validate.RegisterValidation("custom_rule", customValidationFunc)
}

// Load 載入設定檔
// configPath 為設定檔路徑，若為空則使用預設路徑
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 設定預設值
	setDefaults(v)

	// 設定設定檔名稱和路徑
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	// 啟用環境變數覆蓋
	v.SetEnvPrefix("IMG")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 讀取設定檔
	if err := v.ReadInConfig(); err != nil {
		// 如果找不到設定檔，使用預設值
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// 解析設定到結構體
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// 使用 validator 驗證設定結構體
	if err := ValidateConfig(&cfg); err != nil {
		return nil, err
	}

	f, _ := os.OpenFile("/tmp/config_debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		fmt.Fprintf(f, "DEBUG CONFIG: Security=%+v\n", cfg.Security)
		val, ok := os.LookupEnv("IMG_SECURITY_ALLOW_UNSAFE")
		fmt.Fprintf(f, "DEBUG ENV: IMG_SECURITY_ALLOW_UNSAFE=%s (exists=%v)\n", val, ok)
		val2, ok2 := os.LookupEnv("IMG_SECURITY_ALLOWED_SOURCES")
		fmt.Fprintf(f, "DEBUG ENV: IMG_SECURITY_ALLOWED_SOURCES=%s (exists=%v)\n", val2, ok2)
		fmt.Fprintf(f, "DEBUG ALLOWED SOURCES: %v (len=%d)\n", cfg.Security.AllowedSources, len(cfg.Security.AllowedSources))
		f.Close()
	}

	return &cfg, nil
}

// ValidateConfig 使用 validator 驗證設定結構體
func ValidateConfig(cfg *Config) error {
	if err := validate.Struct(cfg); err != nil {
		// 轉換驗證錯誤為更友善的格式
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errMsgs []string
			for _, e := range validationErrors {
				errMsgs = append(errMsgs, formatValidationError(e))
			}
			return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errMsgs, "\n  - "))
		}
		return fmt.Errorf("config validation failed: %w", err)
	}
	return nil
}

// formatValidationError 格式化單個驗證錯誤
func formatValidationError(e validator.FieldError) string {
	field := e.Namespace()
	tag := e.Tag()
	param := e.Param()
	value := e.Value()

	switch tag {
	case "required":
		return fmt.Sprintf("%s 為必填欄位", field)
	case "required_if":
		return fmt.Sprintf("%s 在特定條件下為必填", field)
	case "min":
		return fmt.Sprintf("%s 最小值為 %s（目前值: %v）", field, param, value)
	case "max":
		return fmt.Sprintf("%s 最大值為 %s（目前值: %v）", field, param, value)
	case "oneof":
		return fmt.Sprintf("%s 必須是 [%s] 其中之一（目前值: %v）", field, param, value)
	case "ip":
		return fmt.Sprintf("%s 必須是有效的 IP 地址（目前值: %v）", field, value)
	case "hostname":
		return fmt.Sprintf("%s 必須是有效的主機名稱（目前值: %v）", field, value)
	default:
		return fmt.Sprintf("%s 驗證失敗 (%s: %s)", field, tag, param)
	}
}

// setDefaults 設定預設值
func setDefaults(v *viper.Viper) {
	// Server 預設值
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.max_request_size", 10485760) // 10MB

	// Processing 預設值
	v.SetDefault("processing.default_quality", 85)
	v.SetDefault("processing.max_width", 4096)
	v.SetDefault("processing.max_height", 4096)
	v.SetDefault("processing.workers", 4)
	v.SetDefault("processing.default_format", "jpeg")

	// Security 預設值
	v.SetDefault("security.enabled", false)
	v.SetDefault("security.allow_unsafe", true)
	v.SetDefault("security.security_key", "")
	v.SetDefault("security.allowed_sources", []string{})
	v.SetDefault("security.max_width", 4096)
	v.SetDefault("security.max_height", 4096)

	// Storage 預設值
	v.SetDefault("storage.type", "local")
	v.SetDefault("storage.local.root_path", "./data/images")

	// Cache 預設值
	v.SetDefault("cache.enabled", false)
	v.SetDefault("cache.type", "memory")
	v.SetDefault("cache.redis.host", "localhost")
	v.SetDefault("cache.redis.port", 6379)
	v.SetDefault("cache.redis.db", 0)
	v.SetDefault("cache.redis.ttl", 3600)
	v.SetDefault("cache.memory.max_size", 536870912) // 512MB
	v.SetDefault("cache.memory.ttl", 3600)

	// Logging 預設值
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")

	// Metrics 預設值
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.namespace", "imgfilter")
	v.SetDefault("metrics.path", "/metrics")
	v.SetDefault("metrics.username", "")
	v.SetDefault("metrics.password", "")

	// Swagger 預設值
	v.SetDefault("swagger.enabled", true)
	v.SetDefault("swagger.path", "/swagger")
	v.SetDefault("swagger.username", "")
	v.SetDefault("swagger.password", "")
}

// GetAddress 取得服務器監聽地址
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsConsoleOutput 檢查是否為 console 輸出模式
func (c *Config) IsConsoleOutput() bool {
	return c.Logging.Output == "stdout" && c.Logging.Format == "console"
}

// GetLogLevel 取得 zlogger 對應的日誌等級
func (c *Config) GetLogLevel() string {
	return c.Logging.Level
}
