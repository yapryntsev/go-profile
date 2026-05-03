package domain

import (
	"context"
	"fmt"
	"goph-profile/internal/feature/thumbnail/domain/model"

	"github.com/google/uuid"
)

type (
	DeleteThumbnailAvatarRepo interface {
		Delete(ctx context.Context, key string) error
	}
	DeleteThumbnailMetadataRepo interface {
		FetchThumbnailKeys(ctx context.Context, avatarID uuid.UUID) ([]model.Thumbnail, error)
	}
)

type UseCaseDeleteThumbnail struct {
	metadataRepo DeleteThumbnailMetadataRepo
	avatarRepo   DeleteThumbnailAvatarRepo
}

func NewDeleteThumbnail(
	metadataRepo DeleteThumbnailMetadataRepo,
	avatarRepo DeleteThumbnailAvatarRepo,
) UseCaseDeleteThumbnail {
	return UseCaseDeleteThumbnail{metadataRepo: metadataRepo, avatarRepo: avatarRepo}
}

func (u UseCaseDeleteThumbnail) Run(ctx context.Context, avatarID uuid.UUID) error {
	thumbnails, err := u.metadataRepo.FetchThumbnailKeys(ctx, avatarID)
	if err != nil {
		return fmt.Errorf("failed to fetch thumbnail keys: %w", err)
	}

	for _, t := range thumbnails {
		if err := u.avatarRepo.Delete(ctx, t.URL); err != nil {
			return fmt.Errorf("failed to delete thumbnail %s: %w", t.URL, err)
		}
	}

	return nil
}
