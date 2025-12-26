// Package logger 提供日誌功能
// 使用 zlogger (基於 zap) 實作高效能結構化日誌
package logger

import (
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/zlogger"
)

// Init 初始化日誌系統
func Init(cfg *config.LoggingConfig) {
	// 建立 zlogger 設定
	zloggerCfg := zlogger.DefaultConfig()

	// 設定日誌等級
	zloggerCfg.Level = cfg.Level

	// 設定輸出格式
	switch cfg.Format {
	case "console", "text":
		zloggerCfg.Format = "console"
		zloggerCfg.ColorEnabled = true
	default:
		zloggerCfg.Format = "json"
		zloggerCfg.ColorEnabled = false // JSON 格式禁用顏色，避免 ANSI 轉義序列污染輸出
	}

	// 設定輸出位置
	if cfg.Output == "file" && cfg.FilePath != "" {
		zloggerCfg.Outputs = []string{"file"}
		zloggerCfg.LogPath = cfg.FilePath
	} else {
		zloggerCfg.Outputs = []string{"console"}
	}

	// 開發模式設定
	if cfg.Level == "debug" {
		zloggerCfg.Development = true
		zloggerCfg.AddCaller = true
	}

	// 初始化 zlogger
	zlogger.Init(zloggerCfg)
}

// Debug 輸出 Debug 等級日誌
func Debug(msg string, fields ...zlogger.Field) {
	zlogger.Debug(msg, fields...)
}

// Info 輸出 Info 等級日誌
func Info(msg string, fields ...zlogger.Field) {
	zlogger.Info(msg, fields...)
}

// Warn 輸出 Warn 等級日誌
func Warn(msg string, fields ...zlogger.Field) {
	zlogger.Warn(msg, fields...)
}

// Error 輸出 Error 等級日誌
func Error(msg string, fields ...zlogger.Field) {
	zlogger.Error(msg, fields...)
}

// Fatal 輸出 Fatal 等級日誌並結束程式
func Fatal(msg string, fields ...zlogger.Field) {
	zlogger.Fatal(msg, fields...)
}

// Sync 同步日誌緩衝區
func Sync() error {
	return zlogger.Sync()
}

// 重新匯出常用的 Field 建構函數
var (
	String  = zlogger.String
	Int     = zlogger.Int
	Int64   = zlogger.Int64
	Float64 = zlogger.Float64
	Bool    = zlogger.Bool
	Any     = zlogger.Any
	Err     = zlogger.Err
)
