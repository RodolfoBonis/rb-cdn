package config

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func ZapConfig() *zap.Logger {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		zap.InfoLevel,
	)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
}

func ZapTestConfig() *zap.Logger {
	logger, err := zap.NewProduction()

	if err != nil {
		panic(err.Error())
	}

	return logger
}
