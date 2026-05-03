package app

import (
	"context"
	"errors"
	"goph-profile/internal/config"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Instance struct {
	config config.App

	logger *zap.Logger
	server *http.Server
	db     *pgxpool.Pool

	finishCallbacks []func()
}

func NewInstance(config config.App) *Instance {
	return &Instance{config: config}
}

func NewContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func (i *Instance) Run(ctx context.Context) {
	serverFailed := make(chan struct{}, 1)
	defer close(serverFailed)

	go func() {
		if i.server == nil {
			i.logger.Fatal("trying to run app instance but server is not configured")
		}

		i.logger.Debug(
			"server started",
			zap.String("env", i.config.Env.String()),
			zap.String("addr", i.config.HTTP.Addr),
		)

		if err := i.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			i.logger.Error("server stopped", zap.Error(err))
			serverFailed <- struct{}{}
		}
	}()

	select {
	case <-serverFailed:
	case <-ctx.Done():
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for _, callback := range i.finishCallbacks {
		callback()
	}

	if err := i.server.Shutdown(ctx); err != nil {
		i.logger.Error("failed to shutdown server", zap.Error(err))
	}
	i.logger.Debug("server stopped")
}
