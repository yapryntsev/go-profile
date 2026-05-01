package domain_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"goph-profile/internal/feature/avatar/domain"
	"goph-profile/internal/feature/avatar/domain/mock"
	"goph-profile/internal/feature/avatar/domain/model"
)

func ptr[T any](v T) *T { return &v }

func TestUseCaseGetAvatar_Run(t *testing.T) {
	ctx := context.Background()
	avatarID := uuid.New()
	s3Key := "avatars/original"
	imgData := []byte("fake-image-bytes")

	t.Run("success without filters", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		keyRepo := mock.NewMockGetAvatarKeyRepo(ctrl)
		keyRepo.EXPECT().
			Avatar(ctx, avatarID, (*model.FormatType)(nil), (*model.AspectRatio)(nil)).
			Return(s3Key, nil)

		fetchRepo := mock.NewMockFetchAvatarRepo(ctrl)
		fetchRepo.EXPECT().
			Fetch(ctx, s3Key).
			Return(io.NopCloser(bytes.NewReader(imgData)), nil)

		uc := domain.NewUseCaseGetAvatar(keyRepo, fetchRepo)
		gotImg, gotKey, err := uc.Run(ctx, avatarID, nil, nil)

		require.NoError(t, err)
		assert.Equal(t, imgData, gotImg)
		assert.Equal(t, s3Key, gotKey)
	})

	t.Run("success with format filter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ft := model.JPEG

		keyRepo := mock.NewMockGetAvatarKeyRepo(ctrl)
		keyRepo.EXPECT().
			Avatar(ctx, avatarID, &ft, (*model.AspectRatio)(nil)).
			Return(s3Key, nil)

		fetchRepo := mock.NewMockFetchAvatarRepo(ctrl)
		fetchRepo.EXPECT().
			Fetch(ctx, s3Key).
			Return(io.NopCloser(bytes.NewReader(imgData)), nil)

		uc := domain.NewUseCaseGetAvatar(keyRepo, fetchRepo)
		gotImg, gotKey, err := uc.Run(ctx, avatarID, ptr("jpeg"), nil)

		require.NoError(t, err)
		assert.Equal(t, imgData, gotImg)
		assert.Equal(t, s3Key, gotKey)
	})

	t.Run("success with size filter", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ar := model.S100x100

		keyRepo := mock.NewMockGetAvatarKeyRepo(ctrl)
		keyRepo.EXPECT().
			Avatar(ctx, avatarID, (*model.FormatType)(nil), &ar).
			Return(s3Key, nil)

		fetchRepo := mock.NewMockFetchAvatarRepo(ctrl)
		fetchRepo.EXPECT().
			Fetch(ctx, s3Key).
			Return(io.NopCloser(bytes.NewReader(imgData)), nil)

		uc := domain.NewUseCaseGetAvatar(keyRepo, fetchRepo)
		gotImg, gotKey, err := uc.Run(ctx, avatarID, nil, ptr("100x100"))

		require.NoError(t, err)
		assert.Equal(t, imgData, gotImg)
		assert.Equal(t, s3Key, gotKey)
	})

	t.Run("invalid format string", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		keyRepo := mock.NewMockGetAvatarKeyRepo(ctrl)
		fetchRepo := mock.NewMockFetchAvatarRepo(ctrl)

		uc := domain.NewUseCaseGetAvatar(keyRepo, fetchRepo)
		_, _, err := uc.Run(ctx, avatarID, ptr("bmp"), nil)

		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to parse format")
	})

	t.Run("invalid size string", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		keyRepo := mock.NewMockGetAvatarKeyRepo(ctrl)
		fetchRepo := mock.NewMockFetchAvatarRepo(ctrl)

		uc := domain.NewUseCaseGetAvatar(keyRepo, fetchRepo)
		_, _, err := uc.Run(ctx, avatarID, nil, ptr("9999x9999"))

		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to parse size")
	})

	t.Run("key repo returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repoErr := errors.New("db error")

		keyRepo := mock.NewMockGetAvatarKeyRepo(ctrl)
		keyRepo.EXPECT().
			Avatar(ctx, avatarID, gomock.Any(), gomock.Any()).
			Return("", repoErr)

		fetchRepo := mock.NewMockFetchAvatarRepo(ctrl)

		uc := domain.NewUseCaseGetAvatar(keyRepo, fetchRepo)
		_, _, err := uc.Run(ctx, avatarID, nil, nil)

		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("fetch repo returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		fetchErr := errors.New("s3 error")

		keyRepo := mock.NewMockGetAvatarKeyRepo(ctrl)
		keyRepo.EXPECT().
			Avatar(ctx, avatarID, gomock.Any(), gomock.Any()).
			Return(s3Key, nil)

		fetchRepo := mock.NewMockFetchAvatarRepo(ctrl)
		fetchRepo.EXPECT().
			Fetch(ctx, s3Key).
			Return(nil, fetchErr)

		uc := domain.NewUseCaseGetAvatar(keyRepo, fetchRepo)
		_, _, err := uc.Run(ctx, avatarID, nil, nil)

		require.Error(t, err)
		assert.ErrorIs(t, err, fetchErr)
	})
}
