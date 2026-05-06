package infra

import (
	"context"
	"errors"
	"fmt"
	"goph-profile/internal/feature/avatar/domain/model"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type Repo struct {
	pool   *pgxpool.Pool
	tracer trace.Tracer
}

func NewRepo(pool *pgxpool.Pool, tracer trace.Tracer) Repo {
	return Repo{pool: pool, tracer: tracer}
}

func (r Repo) Avatar(
	ctx context.Context,
	avatarID uuid.UUID,
	format *model.FormatType,
	aspectRatio *model.AspectRatio,
) (string, error) {
	ctx, span := r.tracer.Start(
		ctx,
		"db.select.avatar",
		trace.WithAttributes(
			attribute.String("avatar id", avatarID.String()),
		),
	)
	defer span.End()

	var sel string
	if aspectRatio != nil {
		span.SetAttributes(attribute.String("aspect ratio", string(*aspectRatio)))
		sel = fmt.Sprintf("thumbnail_s3_keys->>'%s'", string(*aspectRatio))
	} else {
		sel = "s3_key"
	}

	wr := sq.Eq{"id": avatarID}
	if format != nil {
		span.SetAttributes(attribute.String("format", string(*format)))
		wr["mime_type"] = string(*format)
	}

	query, args, err := builder.
		Select(sel).
		From("metadata").
		Where(wr).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", fmt.Errorf("building avatar query: %w", err)
	}

	var url string
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&url); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", model.ErrGetAvatarNotFound
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", fmt.Errorf("running get avatar query: %w", err)
	}

	return url, nil
}

func (r Repo) Metadata(ctx context.Context, avatarID uuid.UUID) (model.Metadata, error) {
	ctx, span := r.tracer.Start(
		ctx,
		"db.select.avatar_metadata",
		trace.WithAttributes(attribute.String("avatar id", avatarID.String())),
	)
	defer span.End()

	var metadata model.Metadata

	query, args, err := builder.
		Select("*").
		From("metadata").
		Where(sq.Eq{"id": avatarID}).
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return model.Metadata{}, fmt.Errorf("building metadata query: %w", err)
	}

	err = r.pool.
		QueryRow(ctx, query, args...).
		Scan(
			&metadata.ID,
			&metadata.UserID,
			&metadata.FileName,
			&metadata.MimeType,
			&metadata.Width,
			&metadata.Height,
			&metadata.SizeBytes,
			&metadata.S3Key,
			&metadata.S3ThumbnailKeys,
			&metadata.ProcessingStatus,
			&metadata.CreatedAt,
			&metadata.UpdatedAt,
			&metadata.DeletedAt,
		)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return metadata, model.ErrGetAvatarMetadataNotFound
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return metadata, fmt.Errorf("running get metadata query: %w", err)
	}

	return metadata, nil
}

func (r Repo) Delete(ctx context.Context, userID string, avatarID uuid.UUID) (model.Metadata, error) {
	ctx, span := r.tracer.Start(
		ctx,
		"db.delete.avatar",
		trace.WithAttributes(
			attribute.String("avatar id", avatarID.String()),
			attribute.String("user id", userID),
		),
	)
	defer span.End()

	var metadata model.Metadata

	query, args, err := builder.
		Update("metadata").
		Set("deleted_at", time.Now()).
		Where(sq.Eq{"id": avatarID, "user_id": userID}).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return metadata, fmt.Errorf("building delete query: %w", err)
	}

	err = r.pool.
		QueryRow(ctx, query, args...).
		Scan(
			&metadata.ID,
			&metadata.UserID,
			&metadata.FileName,
			&metadata.MimeType,
			&metadata.Width,
			&metadata.Height,
			&metadata.SizeBytes,
			&metadata.S3Key,
			&metadata.S3ThumbnailKeys,
			&metadata.ProcessingStatus,
			&metadata.CreatedAt,
			&metadata.UpdatedAt,
			&metadata.DeletedAt,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return metadata, model.ErrDeleteAvatarNotFound
		}

		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return metadata, fmt.Errorf("running delete avatar query: %w", err)
	}

	return metadata, nil
}

func (r Repo) Add(
	ctx context.Context,
	avatarID uuid.UUID,
	userID string,
	fileName string,
	mimeType model.FormatType,
	width int,
	height int,
	size int,
	s3key string,
) (model.Metadata, error) {
	ctx, span := r.tracer.Start(
		ctx,
		"db.insert.avatar",
		trace.WithAttributes(
			attribute.String("avatar id", avatarID.String()),
			attribute.String("user id", userID),
		),
	)
	defer span.End()

	var metadata model.Metadata

	query, args, err := builder.
		Insert("metadata").
		SetMap(
			map[string]interface{}{
				"id":         avatarID,
				"user_id":    userID,
				"file_name":  fileName,
				"mime_type":  string(mimeType),
				"width":      width,
				"height":     height,
				"size_bytes": size,
				"s3_key":     s3key,
			},
		).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return metadata, fmt.Errorf("building add query: %w", err)
	}

	err = r.pool.
		QueryRow(ctx, query, args...).
		Scan(
			&metadata.ID,
			&metadata.UserID,
			&metadata.FileName,
			&metadata.MimeType,
			&metadata.Width,
			&metadata.Height,
			&metadata.SizeBytes,
			&metadata.S3Key,
			&metadata.S3ThumbnailKeys,
			&metadata.ProcessingStatus,
			&metadata.CreatedAt,
			&metadata.UpdatedAt,
			&metadata.DeletedAt,
		)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return metadata, fmt.Errorf("running add avatar metadata query: %w", err)
	}

	return metadata, nil
}
