package bootstrap

import (
	"context"
	"goph-profile/internal/config"
	"goph-profile/internal/feature/avatar"
	"goph-profile/internal/feature/s3"
	"goph-profile/internal/transport"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	kafka "github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

type healthChecker struct {
	db        *pgxpool.Pool
	s3        interface{ Ping(context.Context) error }
	kafkaAddr string
}

func (h healthChecker) CheckDB(ctx context.Context) error {
	return h.db.Ping(ctx)
}

func (h healthChecker) CheckS3(ctx context.Context) error {
	return h.s3.Ping(ctx)
}

func (h healthChecker) CheckBroker(ctx context.Context) error {
	conn, err := kafka.DialContext(ctx, "tcp", h.kafkaAddr)
	if err != nil {
		return err
	}
	return conn.Close()
}

func NewServer(
	cfg config.App,
	db *pgxpool.Pool,
	logger *zap.Logger,
	tracerProvider *trace.TracerProvider,
	meterProvider metric.MeterProvider,
) (*http.Server, func(), error) {
	s3Feature, err := s3.New(
		cfg.S3.Addr,
		cfg.S3.AccessKeyID,
		cfg.S3.SecretAccessKey,
		cfg.S3.BucketName,
		tracerProvider,
	)
	if err != nil {
		return nil, nil, err
	}

	avatarFeature, err := avatar.New(
		cfg.Kafka.Addr,
		cfg.Kafka.Topic,
		db,
		s3Feature.S3Service,
		tracerProvider,
		meterProvider,
	)
	if err != nil {
		return nil, nil, err
	}

	checker := healthChecker{
		db:        db,
		s3:        s3Feature.S3Service,
		kafkaAddr: cfg.Kafka.Addr,
	}

	server := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      transport.NewHandler(cfg.Env, logger, avatarFeature.Handler, checker),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return server, avatarFeature.Finish, nil
}
