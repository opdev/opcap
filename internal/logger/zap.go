package logger

import (
	"encoding/json"

	"go.uber.org/zap"
)

var logger *zap.Logger
var Sugar *zap.SugaredLogger

func init() {

	var err error
	rawJSON := []byte(`{
		"level": "debug",
		"encoding": "json",
		"outputPaths": ["stdout", "Logs.json"],
		"errorOutputPaths": ["stderr"],
		"encoderConfig": {
		  "messageKey": "message",
		  "levelKey": "level",
		  "levelEncoder": "lowercase"
		}
	  }`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}
	Sugar = logger.Sugar()
}
