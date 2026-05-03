package domain

import (
	"context"
	"fmt"
	"goph-profile/internal/feature/avatar/domain/model"
	"io"

	"github.com/google/uuid"
)

type (
	GetAvatarKeyRepo interface {
		Avatar(
			ctx context.Context,
			avatarID uuid.UUID,
			format *model.FormatType,
			aspectRatio *model.AspectRatio,
		) (string, error)
	}
	FetchAvatarRepo interface {
		Fetch(ctx context.Context, name string) (io.ReadCloser, error)
	}
)

type UseCaseGetAvatar struct {
	keyRepo    GetAvatarKeyRepo
	avatarRepo FetchAvatarRepo
}

func NewUseCaseGetAvatar(keyRepo GetAvatarKeyRepo, avatarRepo FetchAvatarRepo) UseCaseGetAvatar {
	return UseCaseGetAvatar{keyRepo: keyRepo, avatarRepo: avatarRepo}
}

func (u UseCaseGetAvatar) Run(
	ctx context.Context,
	avatarID uuid.UUID,
	format *string,
	size *string,
) ([]byte, string, error) {
	var formatType *model.FormatType
	var aspectRatio *model.AspectRatio

	if format != nil {
		ft, err := model.NewFormatType(*format)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse format: %w", err)
		}
		formatType = &ft
	}

	if size != nil {
		ar, err := model.NewAspectRatio(*size)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse size: %w", err)
		}
		aspectRatio = &ar
	}

	key, err := u.keyRepo.Avatar(ctx, avatarID, formatType, aspectRatio)
	if err != nil {
		return nil, "", fmt.Errorf("getting avatar s3 key from db: %w", err)
	}

	obj, err := u.avatarRepo.Fetch(ctx, key)
	if err != nil {
		return nil, "", fmt.Errorf("downloading avatar: %w", err)
	}
	defer func() { _ = obj.Close() }()

	img, err := io.ReadAll(obj)
	if err != nil {
		return nil, "", fmt.Errorf("reading avatar: %w", err)
	}

	return img, key, nil
}
