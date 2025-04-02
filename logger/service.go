package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"viot/models"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// zapLogger 是基於 zap 的日誌實現
type zapLogger struct {
	logger *zap.SugaredLogger
	level  string
}

// Debug 實現 Logger 接口
func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	if len(fields) > 0 {
		l.logger.Desugar().Debug(msg, fields...)
	} else {
		l.logger.Debug(msg)
	}
}

// Info 實現 Logger 接口
func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	if len(fields) > 0 {
		l.logger.Desugar().Info(msg, fields...)
	} else {
		l.logger.Info(msg)
	}
}

// Warn 實現 Logger 接口
func (l *zapLogger) Warn(msg string, fields ...zap.Field) {
	if len(fields) > 0 {
		l.logger.Desugar().Warn(msg, fields...)
	} else {
		l.logger.Warn(msg)
	}
}

// Error 實現 Logger 接口
func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	if len(fields) > 0 {
		l.logger.Desugar().Error(msg, fields...)
	} else {
		l.logger.Error(msg)
	}
}

// With 實現 Logger 接口
func (l *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{
		logger: l.logger.Desugar().With(fields...).Sugar(),
		level:  l.level,
	}
}

// Named 實現 Logger 接口
func (l *zapLogger) Named(name string) Logger {
	return &zapLogger{
		logger: l.logger.Named(name),
		level:  l.level,
	}
}

// Clone 實現 Logger 接口
func (l *zapLogger) Clone() Logger {
	return &zapLogger{
		logger: l.logger,
		level:  l.level,
	}
}

// IsDebugMode 實現 Logger 接口
func (l *zapLogger) IsDebugMode() bool {
	return strings.ToLower(l.level) == "debug"
}

// Sync 實現 Logger 接口
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// Close 實現 Logger 接口
func (l *zapLogger) Close() error {
	return l.logger.Sync()
}

// GetLogger 實現 Logger 接口
func (l *zapLogger) GetLogger() *zap.Logger {
	return l.logger.Desugar()
}

// fieldsToArgs 將 Field 轉換為 zap 的參數
func fieldsToArgs(fields []Field) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for _, field := range fields {
		args = append(args, field.Key, field.Value)
	}
	return args
}

// ZapLoggerFactory 是基於 zap 的日誌工廠
type ZapLoggerFactory struct {
	config LoggerConfig
	core   zapcore.Core
	level  string
}

// NewZapLoggerFactory 創建一個新的 zap 日誌工廠
func NewZapLoggerFactory() *ZapLoggerFactory {
	return &ZapLoggerFactory{
		level: "info", // 默認級別
	}
}

// Configure 配置日誌工廠
func (f *ZapLoggerFactory) Configure(config LoggerConfig) error {
	f.config = config
	f.level = config.Level

	// 創建日誌目錄
	if config.OutputPath != "" && config.OutputPath != "stdout" && config.OutputPath != "stderr" {
		dir := filepath.Dir(config.OutputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("創建日誌目錄時出錯: %w", err)
		}
	}

	// 設置日誌級別
	level := zapcore.InfoLevel
	switch strings.ToLower(config.Level) {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	}

	// 設置編碼器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if strings.ToLower(config.Format) == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 設置輸出
	var writeSyncer zapcore.WriteSyncer
	if config.OutputPath == "stdout" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else if config.OutputPath == "stderr" {
		writeSyncer = zapcore.AddSync(os.Stderr)
	} else {
		// 使用 lumberjack 進行日誌輪轉
		ljLogger := &lumberjack.Logger{
			Filename:   config.OutputPath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}
		writeSyncer = zapcore.AddSync(ljLogger)
	}

	// 創建核心
	f.core = zapcore.NewCore(encoder, writeSyncer, zap.NewAtomicLevelAt(level))

	return nil
}

// NewLogger 創建一個新的日誌記錄器
func (f *ZapLoggerFactory) NewLogger(name string) Logger {
	if f.core == nil {
		// 如果未配置，使用默認配置
		if err := f.Configure(LoggerConfig{
			Level:      "info",
			Format:     "console",
			OutputPath: "stdout",
		}); err != nil {
			// 如果配置失敗，使用最小配置
			encoderConfig := zapcore.EncoderConfig{
				MessageKey: "msg",
				LevelKey:   "level",
				TimeKey:    "time",
				EncodeTime: zapcore.ISO8601TimeEncoder,
			}
			f.core = zapcore.NewCore(
				zapcore.NewConsoleEncoder(encoderConfig),
				zapcore.AddSync(os.Stdout),
				zap.NewAtomicLevelAt(zapcore.InfoLevel),
			)
		}
	}

	logger := zap.New(
		f.core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)

	if name != "" {
		logger = logger.Named(name)
	}

	return &zapLogger{
		logger: logger.Sugar(),
		level:  f.level,
	}
}

// DefaultLoggerFactory 是默認的日誌工廠
var DefaultLoggerFactory = NewZapLoggerFactory()

// DefaultLogger 是默認的日誌記錄器
var DefaultLogger = DefaultLoggerFactory.NewLogger("viot")

// Debug 使用默認日誌記錄器記錄調試級別的日誌
func Debug(msg string, fields ...zap.Field) {
	DefaultLogger.Debug(msg, fields...)
}

// Info 使用默認日誌記錄器記錄信息級別的日誌
func Info(msg string, fields ...zap.Field) {
	DefaultLogger.Info(msg, fields...)
}

// Warn 使用默認日誌記錄器記錄警告級別的日誌
func Warn(msg string, fields ...zap.Field) {
	DefaultLogger.Warn(msg, fields...)
}

// Error 使用默認日誌記錄器記錄錯誤級別的日誌
func Error(msg string, fields ...zap.Field) {
	DefaultLogger.Error(msg, fields...)
}

// With 使用默認日誌記錄器創建一個帶有指定字段的新日誌記錄器
func With(fields ...zap.Field) Logger {
	return DefaultLogger.With(fields...)
}

// Named 使用默認日誌記錄器創建一個帶有指定名稱的新日誌記錄器
func Named(name string) Logger {
	return DefaultLogger.Named(name)
}

// Configure 配置默認日誌工廠
func Configure(config models.LoggingConfig) error {
	loggerConfig := LoggerConfig{
		Level:      config.Level,
		Format:     config.Format,
		OutputPath: config.Output,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}
	return DefaultLoggerFactory.Configure(loggerConfig)
}
