package domain_test

import (
	"context"
	"errors"
	"image"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goph-profile/internal/feature/thumbnail/domain"
	"goph-profile/internal/feature/thumbnail/domain/model"
)

type mockMakeThumbnailAvatarRepo struct {
	downloadFn func(ctx context.Context, key string) (image.Image, error)
	uploadFn   func(ctx context.Context, key string, img image.Image) (string, error)
}

func (m *mockMakeThumbnailAvatarRepo) Download(ctx context.Context, key string) (image.Image, error) {
	return m.downloadFn(ctx, key)
}

func (m *mockMakeThumbnailAvatarRepo) Upload(ctx context.Context, key string, img image.Image) (string, error) {
	return m.uploadFn(ctx, key, img)
}

type mockMakeThumbnailMetadataRepo struct {
	fetchS3KeyFn func(ctx context.Context, avatarID uuid.UUID) (string, error)
	updateFn     func(ctx context.Context, avatarID uuid.UUID, thumbnails []model.Thumbnail) error
}

func (m *mockMakeThumbnailMetadataRepo) FetchS3Key(ctx context.Context, avatarID uuid.UUID) (string, error) {
	return m.fetchS3KeyFn(ctx, avatarID)
}

func (m *mockMakeThumbnailMetadataRepo) Update(ctx context.Context, avatarID uuid.UUID, thumbnails []model.Thumbnail) error {
	return m.updateFn(ctx, avatarID, thumbnails)
}

func TestUseCaseMakeThumbnails_Run(t *testing.T) {
	ctx := context.Background()
	avatarID := uuid.New()
	s3key := "avatars/original-key"

	srcImage := image.NewRGBA(image.Rect(0, 0, 400, 400))

	t.Run("success", func(t *testing.T) {
		uploadCount := 0
		metadataRepo := &mockMakeThumbnailMetadataRepo{
			fetchS3KeyFn: func(_ context.Context, id uuid.UUID) (string, error) {
				assert.Equal(t, avatarID, id)
				return s3key, nil
			},
			updateFn: func(_ context.Context, id uuid.UUID, thumbnails []model.Thumbnail) error {
				assert.Equal(t, avatarID, id)
				require.Len(t, thumbnails, 2)
				assert.Equal(t, "100x100", thumbnails[0].Size)
				assert.Equal(t, "300x300", thumbnails[1].Size)
				return nil
			},
		}
		avatarRepo := &mockMakeThumbnailAvatarRepo{
			downloadFn: func(_ context.Context, key string) (image.Image, error) {
				assert.Equal(t, s3key, key)
				return srcImage, nil
			},
			uploadFn: func(_ context.Context, key string, _ image.Image) (string, error) {
				uploadCount++
				return key + "_done", nil
			},
		}

		uc := domain.NewMakeThumbnails(metadataRepo, avatarRepo)
		err := uc.Run(ctx, avatarID)

		require.NoError(t, err)
		assert.Equal(t, 2, uploadCount)
	})

	t.Run("avatar deleted discards thumbnails", func(t *testing.T) {
		metadataRepo := &mockMakeThumbnailMetadataRepo{
			fetchS3KeyFn: func(_ context.Context, _ uuid.UUID) (string, error) {
				return "", model.ErrAvatarDeleted
			},
		}
		avatarRepo := &mockMakeThumbnailAvatarRepo{}

		uc := domain.NewMakeThumbnails(metadataRepo, avatarRepo)
		err := uc.Run(ctx, avatarID)

		require.NoError(t, err)
	})

	t.Run("fetch s3 key fails", func(t *testing.T) {
		fetchErr := errors.New("db unavailable")
		metadataRepo := &mockMakeThumbnailMetadataRepo{
			fetchS3KeyFn: func(_ context.Context, _ uuid.UUID) (string, error) {
				return "", fetchErr
			},
		}
		avatarRepo := &mockMakeThumbnailAvatarRepo{}

		uc := domain.NewMakeThumbnails(metadataRepo, avatarRepo)
		err := uc.Run(ctx, avatarID)

		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to fetch s3 key")
		assert.ErrorIs(t, err, fetchErr)
	})

	t.Run("download fails", func(t *testing.T) {
		downloadErr := errors.New("s3 read error")
		metadataRepo := &mockMakeThumbnailMetadataRepo{
			fetchS3KeyFn: func(_ context.Context, _ uuid.UUID) (string, error) {
				return s3key, nil
			},
		}
		avatarRepo := &mockMakeThumbnailAvatarRepo{
			downloadFn: func(_ context.Context, _ string) (image.Image, error) {
				return nil, downloadErr
			},
		}

		uc := domain.NewMakeThumbnails(metadataRepo, avatarRepo)
		err := uc.Run(ctx, avatarID)

		require.Error(t, err)
		assert.ErrorIs(t, err, downloadErr)
	})

	t.Run("upload small thumbnail fails", func(t *testing.T) {
		uploadErr := errors.New("s3 write error")
		metadataRepo := &mockMakeThumbnailMetadataRepo{
			fetchS3KeyFn: func(_ context.Context, _ uuid.UUID) (string, error) {
				return s3key, nil
			},
		}
		avatarRepo := &mockMakeThumbnailAvatarRepo{
			downloadFn: func(_ context.Context, _ string) (image.Image, error) {
				return srcImage, nil
			},
			uploadFn: func(_ context.Context, _ string, _ image.Image) (string, error) {
				return "", uploadErr
			},
		}

		uc := domain.NewMakeThumbnails(metadataRepo, avatarRepo)
		err := uc.Run(ctx, avatarID)

		require.Error(t, err)
		assert.ErrorIs(t, err, uploadErr)
		assert.ErrorContains(t, err, "failed to upload small thumbnail")
	})

	t.Run("upload medium thumbnail fails", func(t *testing.T) {
		uploadErr := errors.New("s3 write error")
		callCount := 0
		metadataRepo := &mockMakeThumbnailMetadataRepo{
			fetchS3KeyFn: func(_ context.Context, _ uuid.UUID) (string, error) {
				return s3key, nil
			},
		}
		avatarRepo := &mockMakeThumbnailAvatarRepo{
			downloadFn: func(_ context.Context, _ string) (image.Image, error) {
				return srcImage, nil
			},
			uploadFn: func(_ context.Context, _ string, _ image.Image) (string, error) {
				callCount++
				if callCount == 2 {
					return "", uploadErr
				}
				return "uploaded-key", nil
			},
		}

		uc := domain.NewMakeThumbnails(metadataRepo, avatarRepo)
		err := uc.Run(ctx, avatarID)

		require.Error(t, err)
		assert.ErrorIs(t, err, uploadErr)
		assert.ErrorContains(t, err, "failed to upload medium thumbnail")
	})

	t.Run("update metadata fails", func(t *testing.T) {
		updateErr := errors.New("db write error")
		metadataRepo := &mockMakeThumbnailMetadataRepo{
			fetchS3KeyFn: func(_ context.Context, _ uuid.UUID) (string, error) {
				return s3key, nil
			},
			updateFn: func(_ context.Context, _ uuid.UUID, _ []model.Thumbnail) error {
				return updateErr
			},
		}
		avatarRepo := &mockMakeThumbnailAvatarRepo{
			downloadFn: func(_ context.Context, _ string) (image.Image, error) {
				return srcImage, nil
			},
			uploadFn: func(_ context.Context, key string, _ image.Image) (string, error) {
				return key + "_done", nil
			},
		}

		uc := domain.NewMakeThumbnails(metadataRepo, avatarRepo)
		err := uc.Run(ctx, avatarID)

		require.Error(t, err)
		assert.ErrorIs(t, err, updateErr)
		assert.ErrorContains(t, err, "failed to update metadata")
	})
}
