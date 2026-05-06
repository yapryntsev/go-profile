package main

import (
	"context"
	"fmt"
	"goph-profile/internal/app/bootstrap"
	golog "log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	log "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"goph-profile/internal/config"
	"goph-profile/internal/feature/thumbnail/domain"
	"goph-profile/internal/feature/thumbnail/domain/model"
	"goph-profile/internal/feature/thumbnail/infra"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		golog.Fatal(fmt.Errorf("failed to parse config: %w", err))
	}

	telemetry, err := bootstrap.NewTelemetryStack(ctx, cfg)
	if err != nil {
		golog.Fatal(fmt.Errorf("create telemetry stack failed: %w", err))
	}

	logger, err := setupLogger(cfg, telemetry.Logger)
	if err != nil {
		golog.Fatal(fmt.Errorf("failed to setup logger: %w", err))
	}

	pool, err := setupDBPool(ctx, cfg.DB.Conn)
	if err != nil {
		logger.Fatal("failed to setup db pool", zap.Error(err))
	}
	defer pool.Close()

	minioClient, err := minio.New(
		cfg.S3.Addr, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, ""),
			Secure: false,
		},
	)
	if err != nil {
		logger.Fatal("failed to create s3 client", zap.Error(err))
	}

	tracer := telemetry.Tracer.Tracer("worker")
	dbRepo := infra.NewRepo(pool, tracer)
	s3Repo := infra.NewS3Repo(minioClient, tracer, cfg.S3.BucketName)

	logger.Info(
		"starting event-receiver",
		zap.String("addr", cfg.Kafka.Addr),
		zap.String("topic", cfg.Kafka.Topic),
	)

	receiver := infra.NewEventReceiver(tracer, logger, cfg.Kafka.Addr, cfg.Kafka.Topic)
	receiver.Observe(
		ctx, func(event model.Event) {
			switch e := event.(type) {
			case model.UploadEvent:
				if err := domain.NewMakeThumbnails(dbRepo, s3Repo).Run(ctx, e.AvatarID); err != nil {
					logger.Error(
						"failed to make thumbnails",
						zap.String("avatar_id", e.AvatarID.String()),
						zap.Error(err),
					)
				}
			case model.DeleteEvent:
				if err := domain.NewDeleteThumbnail(dbRepo, s3Repo).Run(ctx, e.AvatarID); err != nil {
					logger.Error(
						"failed to delete thumbnails",
						zap.String("avatar_id", e.AvatarID.String()),
						zap.Error(err),
					)
				}
			default:
				logger.Error("unknown event type", zap.Any("event", e))
			}
		},
	)
}

func setupLogger(cfg config.App, loggerProvider *log.LoggerProvider) (*zap.Logger, error) {
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

func setupDBPool(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	if connStr == "" {
		return nil, fmt.Errorf("db connection string is empty")
	}

	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse db uri: %w", err)
	}

	cfg.MaxConnIdleTime = 5 * time.Second
	cfg.MaxConnLifetime = 10 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to db: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping db: %w", err)
	}

	return pool, nil
}
