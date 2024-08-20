package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new logger with the given service name
func New() *zap.Logger {
	var config zap.Config
	var encoderConfig zapcore.EncoderConfig
	env := os.Getenv("ENV")
	config = zap.NewProductionConfig()
	config.Sampling = nil // disable sampling
	encoderConfig = zap.NewProductionEncoderConfig()
	switch env {
	case "stage", "real":
		// Do nothing for "stage" and "real"
	default:
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	encoderConfig.TimeKey = "logged_at"
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.FunctionKey = "func" // use full path function name
	// encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig = encoderConfig
	logger, _ := config.Build()
	defer logger.Sync()

	serviceName := os.Getenv("SERVICE")
	version := os.Getenv("VERSION")
	// Add service name field
	serviceNameField := zap.String("service", serviceName)
	versionField := zap.String("version", version)
	logVersion := zap.String("logVersion", "0")
	withDefaultFields := logger.With(serviceNameField, versionField, logVersion)
	return withDefaultFields
}
