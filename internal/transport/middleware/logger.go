package middleware

import (
	"goph-profile/internal/api"
	"goph-profile/internal/config"

	"github.com/labstack/echo/v4"
	strictecho "github.com/oapi-codegen/runtime/strictmiddleware/echo"
	"go.uber.org/zap"
)

func Logger(env config.Env, logger *zap.Logger) api.StrictMiddlewareFunc {
	logger = logger.Named("transport")
	return func(f strictecho.StrictEchoHandlerFunc, operationID string) strictecho.StrictEchoHandlerFunc {
		return func(ctx echo.Context, request interface{}) (response interface{}, err error) {
			resp, err := f(ctx, request)

			if err != nil {
				logger.Error(
					"failed to serve request",
					zap.Error(err),
					zap.String("operation_id", operationID),
				)
			} else if env == config.Dev {
				logger.Debug(
					"request served",
					zap.String("operation_id", operationID),
				)
			}

			return resp, err
		}
	}
}
