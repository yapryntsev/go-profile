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

func TestUseCaseGetAvatarMetadata_Run(t *testing.T) {
	ctx := context.Background()
	avatarID := uuid.New()

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		expected := model.Metadata{ID: avatarID, UserID: "user-1", S3Key: "avatars/key"}

		repo := mock.NewMockGetAvatarMetadataRepo(ctrl)
		repo.EXPECT().
			Metadata(ctx, avatarID).
			Return(expected, nil)

		uc := domain.NewUseCaseGetAvatarMetadata(repo)
		got, err := uc.Run(ctx, avatarID)

		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		repo := mock.NewMockGetAvatarMetadataRepo(ctrl)
		repo.EXPECT().
			Metadata(ctx, avatarID).
			Return(model.Metadata{}, model.ErrGetAvatarMetadataNotFound)

		uc := domain.NewUseCaseGetAvatarMetadata(repo)
		_, err := uc.Run(ctx, avatarID)

		require.ErrorIs(t, err, model.ErrGetAvatarMetadataNotFound)
	})

	t.Run("repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repoErr := errors.New("db unavailable")

		repo := mock.NewMockGetAvatarMetadataRepo(ctrl)
		repo.EXPECT().
			Metadata(ctx, avatarID).
			Return(model.Metadata{}, repoErr)

		uc := domain.NewUseCaseGetAvatarMetadata(repo)
		_, err := uc.Run(ctx, avatarID)

		require.ErrorIs(t, err, repoErr)
	})
}
