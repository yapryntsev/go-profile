package app

import (
	"context"
	"errors"
	"fmt"
	"goph-profile/internal/app/bootstrap"
	"goph-profile/internal/config"
	"goph-profile/migration"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
)

type Instance struct {
	config config.App

	logger *zap.Logger
	server *http.Server
	db     *pgxpool.Pool

	telemetry bootstrap.Telemetry

	finishCallbacks []func(ctx context.Context) error
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

	var err error
	for _, callback := range i.finishCallbacks {
		err = errors.Join(err, callback(ctx))
	}
	if err != nil {
		i.logger.Error("encounter error during shutdown: fail while executing finish callbacks", zap.Error(err))
	}

	if err := i.server.Shutdown(ctx); err != nil {
		i.logger.Error("encounter error during shutdown: http-server shutdown finished with error", zap.Error(err))
	}
	i.logger.Debug("server stopped")
}

func (i *Instance) Bootstrap(ctx context.Context) (err error) {
	defer func() {
		if err == nil {
			return
		}

		if i.logger != nil {
			i.logger.Error("app instance bootstrap failed", zap.Error(err))
		} else {
			log.Errorf("app instance bootstrap failed: %s", err.Error())
		}
	}()

	telemetry, err := bootstrap.NewTelemetryStack(ctx, i.config)
	if err != nil {
		return fmt.Errorf("create telemetry stack failed: %w", err)
	}
	i.telemetry = telemetry
	i.finishCallbacks = append(i.finishCallbacks, i.telemetry.Shutdown)

	logger, err := bootstrap.NewLogger(i.config, i.telemetry.Logger)
	if err != nil {
		return fmt.Errorf("setup logger failed: %w", err)
	}
	i.logger = logger

	db, err := bootstrap.NewDBPool(ctx, i.config)
	if err != nil {
		return fmt.Errorf("setup db pool failed: %w", err)
	}
	i.db = db

	server, finish, err := bootstrap.NewServer(i.config, i.db, i.logger, i.telemetry.Tracer, i.telemetry.Metric)
	if err != nil {
		return fmt.Errorf("setup server failed: %w", err)
	}
	i.server = server
	i.finishCallbacks = append(
		i.finishCallbacks,
		func(_ context.Context) error {
			finish()
			return nil
		},
	)

	if err := migration.RollMigration(i.db); err != nil {
		return fmt.Errorf("roll migration failed: %w", err)
	}

	return nil
}
