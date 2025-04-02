package logger

import (
	"go.uber.org/zap"
)

// Logger 定義日誌接口
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Named(name string) Logger
	IsDebugMode() bool
	Clone() Logger
	Sync() error
	Close() error
	GetLogger() *zap.Logger
}

// LoggerConfig 日誌配置
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// LoggerFactory 是日誌工廠接口
type LoggerFactory interface {
	// NewLogger 創建一個新的日誌記錄器
	NewLogger(name string) Logger
	// Configure 配置日誌工廠
	Configure(config LoggerConfig) error
}
