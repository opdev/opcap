package logger

import (
	"encoding/json"
	"log"
	"os"

	"go.uber.org/zap"
)

var (
	setLog string
	envLog = os.Getenv("OPCAP_LOG_LEVEL")
	cfg    zap.Config
)

// logJSON implements the structure of Zap Logger
type LogJSON struct {
	EncoderConfig *EncoderCfg `json:"encoderConfig"`
	Level         string      `json:"level"`
	Encoding      string      `json:"encoding"`
	OutputPaths   []string    `json:"outputPaths"`
}

// encoderCfg implements the structure for EncoderConfig field
type EncoderCfg struct {
	MessageKey   string `json:"messageKey"`
	LevelKey     string `json:"levelKey"`
	LevelEncoder string `json:"levelEncoder"`
}

var (
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
)

// initLogger creates the Sugared Zap Logger
// logLever could be set either through the flag "--log-level" or environment variable OPCAP_LOG_LEVEL
// which take precedence respectively
func InitLogger(logLevel string) {
	setLog = envLog
	if logLevel != envLog {
		setLog = logLevel
	}
	data := &LogJSON{
		Level:       setLog,
		Encoding:    "json",
		OutputPaths: []string{"stdout"},
		EncoderConfig: &EncoderCfg{
			MessageKey:   "message",
			LevelKey:     "level",
			LevelEncoder: "lowercase",
		},
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("failed to build zap logger: %v", err)
	}

	if err := json.Unmarshal([]byte(dataJSON), &cfg); err != nil {
		panic(err)
	}
	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}
	sugarLogger = logger.Sugar()
}

// info exports Info Suggared Loglevel
func Info(message string, fields ...interface{}) {
	sugarLogger.Info(message, fields)
}

// debugw exports Suggared Loglevel
func Debugw(message string, fields ...interface{}) {
	sugarLogger.Debugw(message, fields)
}

// debugf exports Suggared Loglevel
func Debugf(message string, fields ...interface{}) {
	sugarLogger.Debugf(message, fields)
}

// errorf exports Suggared Loglevel
func Errorf(message string, fields ...interface{}) {
	sugarLogger.Errorf(message, fields)
}

// fatal exports Suggared Loglevel
func Fatal(message string, fields ...interface{}) {
	sugarLogger.Fatal(message, fields)
}

// fatal exports Suggared Loglevel
func Fatalf(message string, fields ...interface{}) {
	sugarLogger.Fatalf(message, fields)
}

// panic exports Suggared Loglevel
func Panic(message string, fields ...interface{}) {
	sugarLogger.Panic(message, fields)
}
