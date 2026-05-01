package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"goph-profile/internal/feature/avatar/domain"
	"goph-profile/internal/feature/avatar/domain/mock"
	"goph-profile/internal/feature/avatar/domain/model"
)

func TestUseCaseDeleteAvatar_Run(t *testing.T) {
	ctx := context.Background()
	userID := "user-1"
	avatarID := uuid.New()

	deletedMeta := model.Metadata{ID: avatarID, UserID: userID, S3Key: "avatars/key"}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		repo := mock.NewMockDeleteAvatarRepo(ctrl)
		repo.EXPECT().
			Delete(ctx, userID, avatarID).
			Return(deletedMeta, nil)

		dispatcher := mock.NewMockDeleteAvatarEventDispatcher(ctrl)
		dispatcher.EXPECT().
			AvatarDeleted(ctx, avatarID).
			Return(nil)

		uc := domain.NewUseCaseDeleteAvatar(repo, dispatcher)
		err := uc.Run(ctx, userID, avatarID)

		require.NoError(t, err)
	})

	t.Run("repo delete fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		deleteErr := errors.New("db error")

		repo := mock.NewMockDeleteAvatarRepo(ctrl)
		repo.EXPECT().
			Delete(ctx, userID, avatarID).
			Return(model.Metadata{}, deleteErr)

		dispatcher := mock.NewMockDeleteAvatarEventDispatcher(ctrl)

		uc := domain.NewUseCaseDeleteAvatar(repo, dispatcher)
		err := uc.Run(ctx, userID, avatarID)

		require.Error(t, err)
		assert.ErrorIs(t, err, deleteErr)
		assert.ErrorContains(t, err, "failed to delete avatar")
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		repo := mock.NewMockDeleteAvatarRepo(ctrl)
		repo.EXPECT().
			Delete(ctx, userID, avatarID).
			Return(model.Metadata{}, model.ErrDeleteAvatarNotFound)

		dispatcher := mock.NewMockDeleteAvatarEventDispatcher(ctrl)

		uc := domain.NewUseCaseDeleteAvatar(repo, dispatcher)
		err := uc.Run(ctx, userID, avatarID)

		require.ErrorIs(t, err, model.ErrDeleteAvatarNotFound)
	})

	t.Run("dispatch event fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		dispatchErr := errors.New("kafka unavailable")

		repo := mock.NewMockDeleteAvatarRepo(ctrl)
		repo.EXPECT().
			Delete(ctx, userID, avatarID).
			Return(deletedMeta, nil)

		dispatcher := mock.NewMockDeleteAvatarEventDispatcher(ctrl)
		dispatcher.EXPECT().
			AvatarDeleted(ctx, avatarID).
			Return(dispatchErr)

		uc := domain.NewUseCaseDeleteAvatar(repo, dispatcher)
		err := uc.Run(ctx, userID, avatarID)

		require.Error(t, err)
		assert.ErrorIs(t, err, dispatchErr)
		assert.ErrorContains(t, err, "failed to publish message during avatar deletion")
	})
}
