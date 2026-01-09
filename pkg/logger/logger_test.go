package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vincent119/images-filters/internal/config"
	"github.com/vincent119/zlogger"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.LoggingConfig
	}{
		{
			name: "Console Debug",
			cfg: &config.LoggingConfig{
				Level:  "debug",
				Format: "console",
				Output: "console",
			},
		},
		{
			name: "JSON Info",
			cfg: &config.LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "console",
			},
		},
		{
			name: "File Output",
			cfg: &config.LoggingConfig{
				Level:    "info",
				Format:   "json",
				Output:   "file",
				FilePath: "test.log",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				Init(tt.cfg)
			})
		})
	}
}

func TestLoggerFunctions(t *testing.T) {
	// Initialize with console output for testing
	cfg := &config.LoggingConfig{
		Level:  "debug",
		Format: "console",
		Output: "console",
	}
	Init(cfg)

	// Since we can't easily capture stdout/stderr without reassigning os.Stdout/Stderr
	// (which zlogger might not support or might race), we just ensure these don't panic.
	assert.NotPanics(t, func() {
		Debug("debug message", String("key", "value"))
		Info("info message", Int("count", 1))
		Warn("warn message", Bool("active", true))
		Error("error message", Any("obj", map[string]string{"a": "b"}))
		// Fatal will exit, so we can't test it easily without execing a subprocess
	})
}

func TestFieldHelpers(t *testing.T) {
	// Verify that helper functions return valid zlogger.Field
	assert.IsType(t, zlogger.Field{}, String("k", "v"))
	assert.IsType(t, zlogger.Field{}, Int("k", 1))
	assert.IsType(t, zlogger.Field{}, Int64("k", 1))
	assert.IsType(t, zlogger.Field{}, Float64("k", 1.0))
	assert.IsType(t, zlogger.Field{}, Bool("k", true))
	assert.IsType(t, zlogger.Field{}, Any("k", "v"))
	assert.IsType(t, zlogger.Field{}, Err(nil))
}
