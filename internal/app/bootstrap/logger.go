package bootstrap

import (
	"fmt"
	"goph-profile/internal/config"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg config.App, loggerProvider *log.LoggerProvider) (*zap.Logger, error) {
	otelCore := otelzap.NewCore(cfg.Name, otelzap.WithLoggerProvider(loggerProvider))
	coreWrapper := zap.WrapCore(
		func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, otelCore)
		},
	)

	switch cfg.Env {
	case config.Dev:
		return zap.NewDevelopment(coreWrapper)
	case config.Prod:
		return zap.NewProduction(coreWrapper)
	default:
		return nil, fmt.Errorf("unknown env type: %s", cfg.Env)
	}
}
