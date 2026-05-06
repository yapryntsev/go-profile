package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"goph-profile/internal/feature/thumbnail/domain/model"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	builder     = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	ErrNotFound = errors.New("not found")
)

type Repo struct {
	pool   *pgxpool.Pool
	tracer trace.Tracer
}

func NewRepo(pool *pgxpool.Pool, tracer trace.Tracer) Repo {
	return Repo{pool: pool, tracer: tracer}
}

func (r Repo) FetchS3Key(ctx context.Context, avatarID uuid.UUID) (string, error) {
	ctx, span := r.tracer.Start(
		ctx,
		"db.select.s3_key",
		trace.WithAttributes(
			attribute.String("avatar id", avatarID.String()),
		),
	)
	defer span.End()

	query, args, err := builder.
		Select("s3_key", "deleted_at").
		From("metadata").
		Where(sq.Eq{"id": avatarID}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", fmt.Errorf("building fetch s3 key query: %w", err)
	}

	var s3Key string
	var deletedAt *time.Time
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&s3Key, &deletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", fmt.Errorf("running fetch s3 key query: %w", err)
	}

	if deletedAt != nil {
		return "", model.ErrAvatarDeleted
	}

	return s3Key, nil
}

func (r Repo) FetchThumbnailKeys(ctx context.Context, avatarID uuid.UUID) ([]model.Thumbnail, error) {
	ctx, span := r.tracer.Start(
		ctx,
		"db.select.thumbnail_keys",
		trace.WithAttributes(
			attribute.String("avatar id", avatarID.String()),
		),
	)
	defer span.End()

	query, args, err := builder.
		Select("thumbnail_s3_keys").
		From("metadata").
		Where(sq.Eq{"id": avatarID}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, fmt.Errorf("building fetch thumbnail keys query: %w", err)
	}

	var raw []byte
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, fmt.Errorf("running fetch thumbnail keys query: %w", err)
	}

	if raw == nil {
		return nil, nil
	}

	var thumbnails []model.Thumbnail
	if err := json.Unmarshal(raw, &thumbnails); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, fmt.Errorf("unmarshaling thumbnail keys: %w", err)
	}

	return thumbnails, nil
}

func (r Repo) Update(ctx context.Context, avatarID uuid.UUID, thumbnails []model.Thumbnail) error {
	ctx, span := r.tracer.Start(
		ctx,
		"db.update.thumbnail_keys",
		trace.WithAttributes(
			attribute.String("avatar id", avatarID.String()),
		),
	)
	defer span.End()

	data, err := json.Marshal(thumbnails)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return fmt.Errorf("marshaling thumbnails: %w", err)
	}

	query, args, err := builder.
		Update("metadata").
		Set("thumbnail_s3_keys", data).
		Set("processing_status", "completed").
		Where(sq.Eq{"id": avatarID}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return fmt.Errorf("building update query: %w", err)
	}

	if _, err = r.pool.Exec(ctx, query, args...); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return fmt.Errorf("running update thumbnail keys query: %w", err)
	}

	return nil
}
