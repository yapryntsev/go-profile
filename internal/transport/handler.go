package transport

import (
	"context"
	"goph-profile/internal/api"
	"goph-profile/internal/config"
	avatar "goph-profile/internal/feature/avatar/transport"
	"goph-profile/internal/transport/middleware"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.uber.org/zap"
)

type HealthChecker interface {
	CheckDB(ctx context.Context) error
	CheckS3(ctx context.Context) error
	CheckBroker(ctx context.Context) error
}

type Handler struct {
	avatar.Handler
	health HealthChecker
}

func (a Handler) GetHealth(ctx context.Context, request api.GetHealthRequestObject) (
	api.GetHealthResponseObject,
	error,
) {
	return api.GetHealth200JSONResponse{
		Db:     healthStatus(a.health.CheckDB(ctx)),
		S3:     healthStatus(a.health.CheckS3(ctx)),
		Broker: healthStatus(a.health.CheckBroker(ctx)),
	}, nil
}

func healthStatus(err error) string {
	if err != nil {
		return "unavailable"
	}
	return "ok"
}

func NewHandler(
	env config.Env,
	logger *zap.Logger,
	avatar avatar.Handler,
	health HealthChecker,
) http.Handler {
	httpHandler := echo.New()
	appHandler := Handler{
		Handler: avatar,
		health:  health,
	}

	httpHandler.Use(otelecho.Middleware(reflect.TypeOf(appHandler).PkgPath()))
	httpHandler.File("/", "web/static/index.html")

	api.RegisterHandlers(
		httpHandler,
		api.NewStrictHandler(
			appHandler,
			[]api.StrictMiddlewareFunc{
				middleware.Logger(env, logger),
			},
		),
	)

	return httpHandler
}
