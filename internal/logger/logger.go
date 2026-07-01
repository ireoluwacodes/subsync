package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Init builds a zap logger from app environment and log level, registers it globally,
// and returns the configured instance.
func Init(appEnv, logLevel string) (*zap.Logger, error) {
	var cfg zap.Config
	if strings.EqualFold(appEnv, "development") {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}

	level, err := parseLevel(logLevel)
	if err != nil {
		return nil, err
	}
	cfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := cfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		return nil, fmt.Errorf("build zap logger: %w", err)
	}

	zap.ReplaceGlobals(logger)
	return logger, nil
}

// MustInit panics if logger initialization fails.
func MustInit(appEnv, logLevel string) *zap.Logger {
	logger, err := Init(appEnv, logLevel)
	if err != nil {
		panic(err)
	}
	return logger
}

// Sync flushes any buffered log entries. Safe to call on shutdown.
func Sync(logger *zap.Logger) {
	_ = logger.Sync()
}

func parseLevel(level string) (zapcore.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info", "":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown log level: %s", level)
	}
}
