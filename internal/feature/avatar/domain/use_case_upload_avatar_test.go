package domain_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/jpeg"
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

func makeJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, nil))
	return buf.Bytes()
}

func TestUseCaseUploadAvatarMetadata_Run(t *testing.T) {
	ctx := context.Background()
	userID := "user-1"
	fileName := "avatar.jpg"
	s3Key := "avatars/abc123"

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		imgBytes := makeJPEG(t, 50, 50)
		expectedMeta := model.Metadata{ID: uuid.New(), UserID: userID, S3Key: s3Key}

		avatarRepo := mock.NewMockUploadAvatarRepo(ctrl)
		avatarRepo.EXPECT().
			Upload(ctx, gomock.Any(), "jpeg", len(imgBytes), gomock.Any()).
			Return(s3Key, nil)

		metadataRepo := mock.NewMockUploadAvatarMetadataRepo(ctrl)
		metadataRepo.EXPECT().
			Add(ctx, gomock.Any(), userID, fileName, model.JPEG, 50, 50, len(imgBytes), s3Key).
			Return(expectedMeta, nil)

		dispatcher := mock.NewMockUploadAvatarEventDispatcher(ctrl)
		dispatcher.EXPECT().
			AvatarUploaded(ctx, expectedMeta.ID).
			Return(nil)

		uc := domain.NewUseCaseUploadAvatarMetadata(avatarRepo, metadataRepo, dispatcher)
		got, err := uc.Run(ctx, userID, fileName, bytes.NewReader(imgBytes))

		require.NoError(t, err)
		assert.Equal(t, expectedMeta, got)
	})

	t.Run("image too large", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		avatarRepo := mock.NewMockUploadAvatarRepo(ctrl)
		metadataRepo := mock.NewMockUploadAvatarMetadataRepo(ctrl)
		dispatcher := mock.NewMockUploadAvatarEventDispatcher(ctrl)

		oversized := bytes.NewReader(make([]byte, model.AvatarMaxSizeBytes+1))

		uc := domain.NewUseCaseUploadAvatarMetadata(avatarRepo, metadataRepo, dispatcher)
		_, err := uc.Run(ctx, userID, fileName, oversized)

		require.ErrorIs(t, err, model.ErrUploadAvatarTooLarge)
	})

	t.Run("unrecognised image format", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		avatarRepo := mock.NewMockUploadAvatarRepo(ctrl)
		metadataRepo := mock.NewMockUploadAvatarMetadataRepo(ctrl)
		dispatcher := mock.NewMockUploadAvatarEventDispatcher(ctrl)

		uc := domain.NewUseCaseUploadAvatarMetadata(avatarRepo, metadataRepo, dispatcher)
		_, err := uc.Run(ctx, userID, fileName, bytes.NewReader([]byte("not-an-image")))

		require.Error(t, err)
	})

	t.Run("upload to s3 fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		uploadErr := errors.New("s3 unavailable")
		imgBytes := makeJPEG(t, 50, 50)

		avatarRepo := mock.NewMockUploadAvatarRepo(ctrl)
		avatarRepo.EXPECT().
			Upload(ctx, gomock.Any(), "jpeg", len(imgBytes), gomock.Any()).
			Return("", uploadErr)

		metadataRepo := mock.NewMockUploadAvatarMetadataRepo(ctrl)
		dispatcher := mock.NewMockUploadAvatarEventDispatcher(ctrl)

		uc := domain.NewUseCaseUploadAvatarMetadata(avatarRepo, metadataRepo, dispatcher)
		_, err := uc.Run(ctx, userID, fileName, bytes.NewReader(imgBytes))

		require.Error(t, err)
		assert.ErrorIs(t, err, uploadErr)
	})

	t.Run("save metadata fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		addErr := errors.New("db write error")
		imgBytes := makeJPEG(t, 50, 50)

		avatarRepo := mock.NewMockUploadAvatarRepo(ctrl)
		avatarRepo.EXPECT().
			Upload(ctx, gomock.Any(), "jpeg", len(imgBytes), gomock.Any()).
			Return(s3Key, nil)

		metadataRepo := mock.NewMockUploadAvatarMetadataRepo(ctrl)
		metadataRepo.EXPECT().
			Add(ctx, gomock.Any(), userID, fileName, model.JPEG, 50, 50, len(imgBytes), s3Key).
			Return(model.Metadata{}, addErr)

		dispatcher := mock.NewMockUploadAvatarEventDispatcher(ctrl)

		uc := domain.NewUseCaseUploadAvatarMetadata(avatarRepo, metadataRepo, dispatcher)
		_, err := uc.Run(ctx, userID, fileName, bytes.NewReader(imgBytes))

		require.Error(t, err)
		assert.ErrorIs(t, err, addErr)
	})

	t.Run("dispatch event fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		dispatchErr := errors.New("kafka unavailable")
		imgBytes := makeJPEG(t, 50, 50)
		meta := model.Metadata{ID: uuid.New(), UserID: userID, S3Key: s3Key}

		avatarRepo := mock.NewMockUploadAvatarRepo(ctrl)
		avatarRepo.EXPECT().
			Upload(ctx, gomock.Any(), "jpeg", len(imgBytes), gomock.Any()).
			Return(s3Key, nil)

		metadataRepo := mock.NewMockUploadAvatarMetadataRepo(ctrl)
		metadataRepo.EXPECT().
			Add(ctx, gomock.Any(), userID, fileName, model.JPEG, 50, 50, len(imgBytes), s3Key).
			Return(meta, nil)

		dispatcher := mock.NewMockUploadAvatarEventDispatcher(ctrl)
		dispatcher.EXPECT().
			AvatarUploaded(ctx, meta.ID).
			Return(dispatchErr)

		uc := domain.NewUseCaseUploadAvatarMetadata(avatarRepo, metadataRepo, dispatcher)
		_, err := uc.Run(ctx, userID, fileName, bytes.NewReader(imgBytes))

		require.Error(t, err)
		assert.ErrorIs(t, err, dispatchErr)
	})

	t.Run("read image fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		avatarRepo := mock.NewMockUploadAvatarRepo(ctrl)
		metadataRepo := mock.NewMockUploadAvatarMetadataRepo(ctrl)
		dispatcher := mock.NewMockUploadAvatarEventDispatcher(ctrl)

		readErr := errors.New("read error")
		uc := domain.NewUseCaseUploadAvatarMetadata(avatarRepo, metadataRepo, dispatcher)
		_, err := uc.Run(ctx, userID, fileName, &errReader{err: readErr})

		require.Error(t, err)
		assert.ErrorIs(t, err, readErr)
	})
}

// errReader is an io.Reader that always returns an error.
type errReader struct{ err error }

func (r *errReader) Read(_ []byte) (int, error) { return 0, r.err }
func (r *errReader) Close() error               { return nil }

// compile-time check
var _ io.ReadCloser = &errReader{}
