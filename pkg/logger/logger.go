package logger

import (
	"github.com/villageFlower/paypilot_dev_session_service/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Initialize initializes the logger with the given configuration
func Initialize(cfg *config.LogConfig) error {
	var zapConfig zap.Config

	if cfg.Encoding == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Set log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return err
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Set encoding
	zapConfig.Encoding = cfg.Encoding

	// Set output paths
	if len(cfg.OutputPaths) > 0 {
		zapConfig.OutputPaths = cfg.OutputPaths
	}
	if len(cfg.ErrorOutputPaths) > 0 {
		zapConfig.ErrorOutputPaths = cfg.ErrorOutputPaths
	}

	// Build logger
	logger, err := zapConfig.Build()
	if err != nil {
		return err
	}

	Log = logger
	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
