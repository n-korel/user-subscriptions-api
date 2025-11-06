package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerInterface interface {
	Info(message string, fields map[string]any)
	Error(message string, fields map[string]any)
	Warn(message string, fields map[string]any)
	Debug(message string, fields map[string]any)
	Fatal(message string, fields map[string]any)
	Sync() error
}

type Logger struct {
	*zap.Logger
}

var _ LoggerInterface = (*Logger)(nil)

func New(level string) (*Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Encoding:          "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "timestamp",
			NameKey:        "logger",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	zapLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{zapLogger}, nil
}

func (l *Logger) Info(message string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	l.Logger.Info(message, l.mapToFields(fields)...)
}

func (l *Logger) Error(message string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	l.Logger.Error(message, l.mapToFields(fields)...)
}

func (l *Logger) Warn(message string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	l.Logger.Warn(message, l.mapToFields(fields)...)
}

func (l *Logger) Debug(message string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	l.Logger.Debug(message, l.mapToFields(fields)...)
}

func (l *Logger) Fatal(message string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	l.Logger.Fatal(message, l.mapToFields(fields)...)
}

func (l *Logger) mapToFields(m map[string]any) []zap.Field {
	fields := make([]zap.Field, 0, len(m))
	for k, v := range m {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}
