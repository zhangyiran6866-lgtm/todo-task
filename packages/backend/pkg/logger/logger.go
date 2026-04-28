package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config defines logger initialization options.
type Config struct {
	Level         string
	Format        string
	AppPath       string
	ErrorPath     string
	AuditPath     string
	RetentionDays int
	Compress      bool
	Stdout        bool
}

// New initializes structured application/error channels and audit logger.
func New(cfg Config) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(cfg.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	var encoderConfig zapcore.EncoderConfig
	if cfg.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.LevelKey = "level"
	encoderConfig.MessageKey = "message"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewJSONEncoder(encoderConfig)
	if cfg.Format != "json" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	appSink, err := newDailyRotateWriter(cfg.AppPath, cfg.RetentionDays, cfg.Compress)
	if err != nil {
		return nil, fmt.Errorf("logger.New: init app sink failed: %w", err)
	}
	errorSink, err := newDailyRotateWriter(cfg.ErrorPath, cfg.RetentionDays, cfg.Compress)
	if err != nil {
		return nil, fmt.Errorf("logger.New: init error sink failed: %w", err)
	}
	auditSink, err := newDailyRotateWriter(cfg.AuditPath, cfg.RetentionDays, cfg.Compress)
	if err != nil {
		return nil, fmt.Errorf("logger.New: init audit sink failed: %w", err)
	}

	cores := []zapcore.Core{
		zapcore.NewCore(encoder, zapcore.AddSync(appSink), zap.NewAtomicLevelAt(zapLevel)),
		zapcore.NewCore(encoder, zapcore.AddSync(errorSink), zapcore.ErrorLevel),
	}

	if cfg.Stdout {
		consoleEncoderCfg := encoderConfig
		consoleEncoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cores = append(cores, zapcore.NewCore(
			zapcore.NewConsoleEncoder(consoleEncoderCfg),
			zapcore.AddSync(os.Stdout),
			zap.NewAtomicLevelAt(zapLevel),
		))
	}

	appLogger := zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	auditLogger := zap.New(
		zapcore.NewCore(encoder, zapcore.AddSync(auditSink), zapcore.InfoLevel),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	setAuditLogger(auditLogger)

	return appLogger, nil
}
