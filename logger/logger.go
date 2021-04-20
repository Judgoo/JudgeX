package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func initLogger() *zap.SugaredLogger {
	var cfg zap.Config = zap.NewProductionConfig()
	cfg.OutputPaths = []string{"judgex.log"}
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	logger, err := cfg.Build()
	if err != nil {

		panic(err)
	}
	sugar := logger.Sugar()
	defer sugar.Sync()
	return sugar
}

var Sugar = initLogger()
