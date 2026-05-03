package app

import (
	"context"
	"fmt"
	"goph-profile/internal/config"
	"goph-profile/internal/feature/avatar"
	"goph-profile/internal/feature/s3"
	"goph-profile/internal/transport"
	"goph-profile/migration"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
)

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

	if err := i.setupDBPool(ctx); err != nil {
		return fmt.Errorf("setup db pool failed: %w", err)
	}

	if err := i.setupLogger(); err != nil {
		return fmt.Errorf("setup logger failed: %w", err)
	}

	if err := i.setupServer(); err != nil {
		return fmt.Errorf("setup server failed: %w", err)
	}

	if err := migration.RollMigration(i.db); err != nil {
		return fmt.Errorf("roll migration failed: %w", err)
	}

	return nil
}

func (i *Instance) setupLogger() error {
	var err error
	var logger *zap.Logger

	switch i.config.Env {
	case config.Dev:
		logger, err = zap.NewDevelopment()
	case config.Prod:
		logger, err = zap.NewProduction()
	default:
		return fmt.Errorf("unknown env type: %s", i.config.Env)
	}

	if err != nil {
		return err
	}

	i.logger = logger

	return nil
}

func (i *Instance) setupDBPool(ctx context.Context) error {
	url := i.config.DB.Conn
	if url == "" {
		return fmt.Errorf("expected correct uri, found empty string instead")
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return fmt.Errorf("unable to parse uri: %w", err)
	}

	cfg.MaxConnIdleTime = 5 * time.Second
	cfg.MaxConnLifetime = 10 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("unable to connect to db: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping db: %w", err)
	}

	i.db = pool

	return nil
}

func (i *Instance) setupServer() error {
	s3Service, err := s3.NewService(
		i.config.S3.Addr,
		i.config.S3.AccessKeyID,
		i.config.S3.SecretAccessKey,
		i.config.S3.BucketName,
	)
	if err != nil {
		return err
	}

	avatarHandler, finish := avatar.New(i.db, s3Service, i.config.Kafka.Addr, i.config.Kafka.Topic)
	i.finishCallbacks = append(i.finishCallbacks, finish)

	i.server = &http.Server{
		Addr:         i.config.HTTP.Addr,
		Handler:      transport.NewHandler(i.config.Env, i.logger, avatarHandler),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return nil
}
