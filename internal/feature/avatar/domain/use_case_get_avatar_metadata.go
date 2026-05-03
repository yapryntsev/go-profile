package domain

import (
	"context"
	"goph-profile/internal/feature/avatar/domain/model"

	"github.com/google/uuid"
)

type GetAvatarMetadataRepo interface {
	Metadata(ctx context.Context, avatarID uuid.UUID) (model.Metadata, error)
}

type UseCaseGetAvatarMetadata struct {
	repo GetAvatarMetadataRepo
}

func NewUseCaseGetAvatarMetadata(repo GetAvatarMetadataRepo) UseCaseGetAvatarMetadata {
	return UseCaseGetAvatarMetadata{repo: repo}
}

func (u UseCaseGetAvatarMetadata) Run(ctx context.Context, avatarID uuid.UUID) (model.Metadata, error) {
	return u.repo.Metadata(ctx, avatarID)
}
