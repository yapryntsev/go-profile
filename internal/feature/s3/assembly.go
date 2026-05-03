package s3

import (
	"goph-profile/internal/feature/s3/domain"
	"goph-profile/internal/feature/s3/infra"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewService(
	endpoint string,
	accessKeyID string,
	secretAccessKey string,
	bucketName string,
) (domain.Service, error) {
	client, err := minio.New(
		endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
			Secure: false,
		},
	)
	if err != nil {
		return domain.Service{}, err
	}

	repo, err := infra.NewRepo(bucketName, client)
	if err != nil {
		return domain.Service{}, err
	}

	return domain.NewService(repo), nil
}
