package common

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() (*zap.Logger, error) {
	productionConfig := zap.NewProductionConfig()
	productionConfig.Encoding = "console"
	productionConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return productionConfig.Build()
}
