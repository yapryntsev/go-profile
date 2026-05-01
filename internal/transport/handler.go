package transport

import (
	"context"
	"goph-profile/internal/api"
	"goph-profile/internal/config"
	avatar "goph-profile/internal/feature/avatar/transport"
	"goph-profile/internal/transport/middleware"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler struct {
	avatar.Handler
}

func (a Handler) GetHealth(ctx context.Context, request api.GetHealthRequestObject) (
	api.GetHealthResponseObject,
	error,
) {
	//TODO implement me
	panic("implement me")
}

func NewHandler(
	env config.Env,
	logger *zap.Logger,
	avatar avatar.Handler,
) http.Handler {
	httpHandler := echo.New()
	appHandler := Handler{
		avatar,
	}

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
