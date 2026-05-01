package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goph-profile/internal/feature/thumbnail/domain"
	"goph-profile/internal/feature/thumbnail/domain/model"
)

type mockDeleteThumbnailAvatarRepo struct {
	deleteFn func(ctx context.Context, key string) error
}

func (m *mockDeleteThumbnailAvatarRepo) Delete(ctx context.Context, key string) error {
	return m.deleteFn(ctx, key)
}

type mockDeleteThumbnailMetadataRepo struct {
	fetchThumbnailKeysFn func(ctx context.Context, avatarID uuid.UUID) ([]model.Thumbnail, error)
}

func (m *mockDeleteThumbnailMetadataRepo) FetchThumbnailKeys(ctx context.Context, avatarID uuid.UUID) ([]model.Thumbnail, error) {
	return m.fetchThumbnailKeysFn(ctx, avatarID)
}

func TestUseCaseDeleteThumbnail_Run(t *testing.T) {
	ctx := context.Background()
	avatarID := uuid.New()

	thumbnails := []model.Thumbnail{
		{Size: "100x100", URL: "avatars/key_100x100"},
		{Size: "300x300", URL: "avatars/key_300x300"},
	}

	t.Run("success", func(t *testing.T) {
		var deletedKeys []string
		metadataRepo := &mockDeleteThumbnailMetadataRepo{
			fetchThumbnailKeysFn: func(_ context.Context, id uuid.UUID) ([]model.Thumbnail, error) {
				assert.Equal(t, avatarID, id)
				return thumbnails, nil
			},
		}
		avatarRepo := &mockDeleteThumbnailAvatarRepo{
			deleteFn: func(_ context.Context, key string) error {
				deletedKeys = append(deletedKeys, key)
				return nil
			},
		}

		uc := domain.NewDeleteThumbnail(metadataRepo, avatarRepo)
		err := uc.Run(ctx, avatarID)

		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"avatars/key_100x100", "avatars/key_300x300"}, deletedKeys)
	})

	t.Run("no thumbnails", func(t *testing.T) {
		metadataRepo := &mockDeleteThumbnailMetadataRepo{
			fetchThumbnailKeysFn: func(_ context.Context, _ uuid.UUID) ([]model.Thumbnail, error) {
				return nil, nil
			},
		}
		avatarRepo := &mockDeleteThumbnailAvatarRepo{}

		uc := domain.NewDeleteThumbnail(metadataRepo, avatarRepo)
		err := uc.Run(ctx, avatarID)

		require.NoError(t, err)
	})

	t.Run("fetch thumbnail keys fails", func(t *testing.T) {
		fetchErr := errors.New("db unavailable")
		metadataRepo := &mockDeleteThumbnailMetadataRepo{
			fetchThumbnailKeysFn: func(_ context.Context, _ uuid.UUID) ([]model.Thumbnail, error) {
				return nil, fetchErr
			},
		}
		avatarRepo := &mockDeleteThumbnailAvatarRepo{}

		uc := domain.NewDeleteThumbnail(metadataRepo, avatarRepo)
		err := uc.Run(ctx, avatarID)

		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to fetch thumbnail keys")
		assert.ErrorIs(t, err, fetchErr)
	})

	t.Run("s3 delete fails", func(t *testing.T) {
		deleteErr := errors.New("s3 delete error")
		metadataRepo := &mockDeleteThumbnailMetadataRepo{
			fetchThumbnailKeysFn: func(_ context.Context, _ uuid.UUID) ([]model.Thumbnail, error) {
				return thumbnails, nil
			},
		}
		avatarRepo := &mockDeleteThumbnailAvatarRepo{
			deleteFn: func(_ context.Context, _ string) error {
				return deleteErr
			},
		}

		uc := domain.NewDeleteThumbnail(metadataRepo, avatarRepo)
		err := uc.Run(ctx, avatarID)

		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to delete thumbnail")
		assert.ErrorIs(t, err, deleteErr)
	})

	t.Run("second thumbnail delete fails", func(t *testing.T) {
		deleteErr := errors.New("s3 delete error")
		callCount := 0
		metadataRepo := &mockDeleteThumbnailMetadataRepo{
			fetchThumbnailKeysFn: func(_ context.Context, _ uuid.UUID) ([]model.Thumbnail, error) {
				return thumbnails, nil
			},
		}
		avatarRepo := &mockDeleteThumbnailAvatarRepo{
			deleteFn: func(_ context.Context, _ string) error {
				callCount++
				if callCount == 2 {
					return deleteErr
				}
				return nil
			},
		}

		uc := domain.NewDeleteThumbnail(metadataRepo, avatarRepo)
		err := uc.Run(ctx, avatarID)

		require.Error(t, err)
		assert.ErrorIs(t, err, deleteErr)
		assert.Equal(t, 2, callCount)
	})
}
