package s3

import (
	"goph-profile/internal/feature/s3/domain"
	"goph-profile/internal/feature/s3/infra"
	"reflect"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.opentelemetry.io/otel/trace"
)

type Feature struct {
	S3Service domain.Service
}

func New(
	endpoint string,
	accessKeyID string,
	secretAccessKey string,
	bucketName string,
	tracerProvider trace.TracerProvider,
) (Feature, error) {
	client, err := minio.New(
		endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
			Secure: false,
		},
	)
	if err != nil {
		return Feature{}, err
	}
	repo, err := infra.NewRepo(bucketName, tracerProvider.Tracer(reflect.TypeOf(Feature{}).PkgPath()), client)
	if err != nil {
		return Feature{}, err
	}

	return Feature{S3Service: domain.NewService(repo)}, nil
}
