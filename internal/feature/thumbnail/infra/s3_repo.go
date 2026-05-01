package infra

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/jpeg"
	_ "image/png"

	"github.com/minio/minio-go/v7"
	_ "golang.org/x/image/webp"
)

type S3Repo struct {
	client     *minio.Client
	bucketName string
}

func NewS3Repo(client *minio.Client, bucketName string) S3Repo {
	return S3Repo{client: client, bucketName: bucketName}
}

func (r S3Repo) Download(ctx context.Context, key string) (image.Image, error) {
	obj, err := r.client.GetObject(ctx, r.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting object from s3: %w", err)
	}
	defer obj.Close()

	img, _, err := image.Decode(obj)
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}

	return img, nil
}

func (r S3Repo) Upload(ctx context.Context, key string, img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
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
		return "", fmt.Errorf("uploading thumbnail to s3: %w", err)
	}

	return key, nil
}

func (r S3Repo) Delete(ctx context.Context, key string) error {
	if err := r.client.RemoveObject(ctx, r.bucketName, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("removing object from s3: %w", err)
	}
	return nil
}
