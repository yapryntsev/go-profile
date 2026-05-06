package infra

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"

	"github.com/minio/minio-go/v7"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	_ "golang.org/x/image/webp"
)

type S3Repo struct {
	client     *minio.Client
	tracer     trace.Tracer
	bucketName string
}

func NewS3Repo(client *minio.Client, tracer trace.Tracer, bucketName string) S3Repo {
	return S3Repo{client: client, tracer: tracer, bucketName: bucketName}
}

func (r S3Repo) Download(ctx context.Context, key string) (image.Image, error) {
	ctx, span := r.tracer.Start(
		ctx,
		"s3.fetch.avatar",
		trace.WithAttributes(
			attribute.String("object key", key),
		),
	)
	defer span.End()

	obj, err := r.client.GetObject(ctx, r.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, fmt.Errorf("getting object from s3: %w", err)
	}
	defer obj.Close()

	img, _, err := image.Decode(obj)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return nil, fmt.Errorf("decoding image: %w", err)
	}

	return img, nil
}

func (r S3Repo) Upload(ctx context.Context, key string, img image.Image) (string, error) {
	ctx, span := r.tracer.Start(
		ctx,
		"s3.upload.thumbnail",
		trace.WithAttributes(
			attribute.String("object key", key),
		),
	)
	defer span.End()

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", fmt.Errorf("encoding thumbnail as jpeg: %w", err)
	}

	data := buf.Bytes()
	_, err := r.client.PutObject(
		ctx,
		r.bucketName,
		key,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{ContentType: "image/jpeg"},
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return "", fmt.Errorf("uploading thumbnail to s3: %w", err)
	}

	return key, nil
}

func (r S3Repo) Delete(ctx context.Context, key string) error {
	ctx, span := r.tracer.Start(
		ctx,
		"s3.delete.avatar",
		trace.WithAttributes(
			attribute.String("object key", key),
		),
	)

	if err := r.client.RemoveObject(ctx, r.bucketName, key, minio.RemoveObjectOptions{}); err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)

		return fmt.Errorf("removing object from s3: %w", err)
	}
	return nil
}
