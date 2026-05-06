package infra

import (
	"context"
	"fmt"
	"goph-profile/internal/feature/s3/domain/model"
	"io"

	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Repo struct {
	bucketName string
	client     *minio.Client
	tracer     trace.Tracer
}

func NewRepo(bucketName string, tracer trace.Tracer, client *minio.Client) (Repo, error) {
	repo := Repo{bucketName: bucketName, tracer: tracer, client: client}

	if err := repo.initialConfiguration(); err != nil {
		return repo, err
	}

	return repo, nil
}

func (r Repo) Upload(
	ctx context.Context,
	name string,
	format model.FormatType,
	size int,
	img io.Reader,
) (string, error) {
	ctx, span := r.tracer.Start(
		ctx,
		"s3.upload.avatar",
		trace.WithAttributes(
			attribute.String("avatar name", name),
			attribute.String("format", string(format)),
			attribute.String("size", fmt.Sprint(size)),
		),
	)
	defer span.End()

	info, err := r.client.PutObject(
		ctx,
		r.bucketName,
		name,
		img,
		int64(size),
		minio.PutObjectOptions{ContentType: string(format)},
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", err
	}

	return info.Key, nil
}

func (r Repo) Fetch(ctx context.Context, name string) (io.ReadCloser, error) {
	ctx, span := r.tracer.Start(
		ctx,
		"s3.fetch.avatar",
		trace.WithAttributes(
			attribute.String("avatar name", name),
		),
	)
	defer span.End()

	obj, err := r.client.GetObject(ctx, r.bucketName, name, minio.GetObjectOptions{})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return nil, err
	}

	return obj, nil
}

func (r Repo) Ping(ctx context.Context) error {
	_, err := r.client.ListBuckets(ctx)
	return err
}

func (r Repo) initialConfiguration() error {
	ctx := context.Background()
	isExists, err := r.client.BucketExists(ctx, r.bucketName)
	if err != nil {
		return err
	}

	if isExists {
		return err
	}

	return r.client.MakeBucket(ctx, r.bucketName, minio.MakeBucketOptions{})
}
