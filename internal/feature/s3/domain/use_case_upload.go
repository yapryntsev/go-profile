package domain

import (
	"context"
	"goph-profile/internal/feature/s3/domain/model"
	"io"
)

type UploadRepo interface {
	Upload(ctx context.Context, name string, format model.FormatType, size int, img io.Reader) (string, error)
}

type UseCaseUpload struct {
	repo UploadRepo
}

func NewUseCaseUpload(repo UploadRepo) UseCaseUpload {
	return UseCaseUpload{repo: repo}
}

func (u UseCaseUpload) Run(ctx context.Context, name string, format string, size int, img io.Reader) (string, error) {
	formatType, err := model.NewFormatType(format)
	if err != nil {
		return "", model.ErrUploadUnknownMIME
	}

	return u.repo.Upload(ctx, name, formatType, size, img)
}
