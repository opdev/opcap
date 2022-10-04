package logger

import (
	"os"

	"go.uber.org/zap"
)

var (
	setLog string
	envLog = os.Getenv("OPCAP_LOG_LEVEL")
	cfg    zap.Config
)

var sugarLogger *zap.SugaredLogger

// InitLogger creates the Sugared Zap Logger
// logLevel could be set either through the flag "--log-level" or environment variable OPCAP_LOG_LEVEL
// which takes precedence respectively
func InitLogger(logLevel string) error {
	setLog = envLog
	if logLevel != envLog {
		setLog = logLevel
	}

	atomicLevel, err := zap.ParseAtomicLevel(setLog)
	if err != nil {
		return err
	}

	cfg = zap.NewProductionConfig()
	cfg.Level = atomicLevel
	cfg.EncoderConfig.MessageKey = "message"

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	sugarLogger = logger.Sugar()

	return nil
}

// info exports Info Suggared Loglevel
func Infow(message string, fields ...interface{}) {
	sugarLogger.Infow(message, fields...)
}

// debugw exports Suggared Loglevel
func Debugw(message string, fields ...interface{}) {
	sugarLogger.Debugw(message, fields...)
}

// debugf exports Suggared Loglevel
func Debugf(message string, fields ...interface{}) {
	sugarLogger.Debugf(message, fields...)
}

// errorf exports Suggared Loglevel
func Errorf(message string, fields ...interface{}) {
	sugarLogger.Errorf(message, fields...)
}

// Errorw exports sugared LogLevel Error
func Errorw(message string, fields ...interface{}) {
	sugarLogger.Errorw(message, fields...)
}
