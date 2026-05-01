package domain

import (
	"context"
	"io"
)

type ServiceRepo interface {
	UploadRepo
	FetchRepo
}
type Service struct {
	upload UseCaseUpload
	fetch  UseCaseFetch
}

func NewService(repo ServiceRepo) Service {
	return Service{
		NewUseCaseUpload(repo),
		NewUseCaseFetch(repo),
	}
}

func (s Service) Upload(ctx context.Context, name string, format string, size int, img io.Reader) (string, error) {
	return s.upload.Run(ctx, name, format, size, img)
}

func (s Service) Fetch(ctx context.Context, name string) (io.ReadCloser, error) {
	return s.fetch.Run(ctx, name)
}
