package domain

import (
	"context"
	"io"
)

type ServiceRepo interface {
	UploadRepo
	FetchRepo
	Ping(ctx context.Context) error
}
type Service struct {
	upload UseCaseUpload
	fetch  UseCaseFetch
	repo   ServiceRepo
}

func NewService(repo ServiceRepo) Service {
	return Service{
		upload: NewUseCaseUpload(repo),
		fetch:  NewUseCaseFetch(repo),
		repo:   repo,
	}
}

func (s Service) Upload(ctx context.Context, name string, format string, size int, img io.Reader) (string, error) {
	return s.upload.Run(ctx, name, format, size, img)
}

func (s Service) Fetch(ctx context.Context, name string) (io.ReadCloser, error) {
	return s.fetch.Run(ctx, name)
}

func (s Service) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
