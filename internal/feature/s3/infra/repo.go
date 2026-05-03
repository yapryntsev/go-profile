package infra

import (
	"context"
	"goph-profile/internal/feature/s3/domain/model"
	"io"

	"github.com/minio/minio-go/v7"
)

type Repo struct {
	bucketName string
	client     *minio.Client
}

func NewRepo(bucketName string, client *minio.Client) (Repo, error) {
	repo := Repo{bucketName: bucketName, client: client}

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
	info, err := r.client.PutObject(
		ctx,
		r.bucketName,
		name,
		img,
		int64(size),
		minio.PutObjectOptions{ContentType: string(format)},
	)
	if err != nil {
		return "", err
	}

	return info.Key, nil
}

func (r Repo) Fetch(ctx context.Context, name string) (io.ReadCloser, error) {
	return r.client.GetObject(ctx, r.bucketName, name, minio.GetObjectOptions{})
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
