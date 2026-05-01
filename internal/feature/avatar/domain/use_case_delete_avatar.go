package domain

import (
	"context"
	"fmt"
	"goph-profile/internal/feature/avatar/domain/model"

	"github.com/google/uuid"
)

type (
	DeleteAvatarRepo interface {
		Delete(ctx context.Context, userID string, avatarID uuid.UUID) (model.Metadata, error)
	}
	DeleteAvatarEventDispatcher interface {
		AvatarDeleted(ctx context.Context, avatarID uuid.UUID) error
	}
)

type UseCaseDeleteAvatar struct {
	repo       DeleteAvatarRepo
	dispatcher DeleteAvatarEventDispatcher
}

func NewUseCaseDeleteAvatar(
	repo DeleteAvatarRepo,
	dispatcher DeleteAvatarEventDispatcher,
) UseCaseDeleteAvatar {
	return UseCaseDeleteAvatar{repo: repo, dispatcher: dispatcher}
}

func (u UseCaseDeleteAvatar) Run(ctx context.Context, userID string, avatarID uuid.UUID) error {
	_, err := u.repo.Delete(ctx, userID, avatarID)
	if err != nil {
		return fmt.Errorf("failed to delete avatar: %w", err)
	}

	if err := u.dispatcher.AvatarDeleted(ctx, avatarID); err != nil {
		return fmt.Errorf("failed to publish message during avatar deletion: %w", err)
	}

	return nil
}
