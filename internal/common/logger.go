package common

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() (*zap.Logger, error) {
	level := zap.NewAtomicLevel()
	level.SetLevel(zap.DebugLevel)
	productionConfig := zap.NewProductionConfig()
	productionConfig.Encoding = "console"
	productionConfig.Level = level
	productionConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return productionConfig.Build()
}
