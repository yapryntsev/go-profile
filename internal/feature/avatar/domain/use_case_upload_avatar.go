package domain

import (
	"bytes"
	"context"
	"fmt"
	"goph-profile/internal/feature/avatar/domain/model"
	"image"
	"io"

	"github.com/google/uuid"
)

type (
	UploadAvatarMetadataRepo interface {
		Add(
			ctx context.Context,
			avatarID uuid.UUID,
			userID string,
			fileName string,
			mimeType model.FormatType,
			width int,
			height int,
			size int,
			s3key string,
		) (model.Metadata, error)
	}
	UploadAvatarRepo interface {
		Upload(ctx context.Context, name string, format string, size int, img io.Reader) (string, error)
	}
	UploadAvatarEventDispatcher interface {
		AvatarUploaded(ctx context.Context, avatarID uuid.UUID) error
	}
)

type UseCaseUploadAvatarMetadata struct {
	avatarRepo   UploadAvatarRepo
	metadataRepo UploadAvatarMetadataRepo
	dispatcher   UploadAvatarEventDispatcher
}

func NewUseCaseUploadAvatarMetadata(
	avatarRepo UploadAvatarRepo,
	metadataRepo UploadAvatarMetadataRepo,
	dispatcher UploadAvatarEventDispatcher,
) UseCaseUploadAvatarMetadata {
	return UseCaseUploadAvatarMetadata{avatarRepo: avatarRepo, metadataRepo: metadataRepo, dispatcher: dispatcher}
}

func (u UseCaseUploadAvatarMetadata) Run(
	ctx context.Context,
	userID string,
	name string,
	img io.Reader,
) (model.Metadata, error) {
	var metadata model.Metadata
	avatarID := uuid.New()

	imgData, err := io.ReadAll(img)
	if err != nil {
		return metadata, fmt.Errorf("failed to read avatar image: %w", err)
	}

	imgSize := len(imgData)
	if imgSize > model.AvatarMaxSizeBytes {
		return metadata, model.ErrUploadAvatarTooLarge
	}

	cfg, rawFormatType, err := image.DecodeConfig(bytes.NewReader(imgData))
	if err != nil {
		return metadata, err
	}

	formatType, err := model.NewFormatType(rawFormatType)
	if err != nil {
		return metadata, model.ErrUploadAvatarUnknown
	}

	s3Key, err := u.avatarRepo.Upload(ctx, avatarID.String(), rawFormatType, imgSize, bytes.NewReader(imgData))
	if err != nil {
		return metadata, fmt.Errorf("field to upload avatar to s3: %w", err)
	}

	metadata, err = u.metadataRepo.Add(ctx, avatarID, userID, name, formatType, cfg.Width, cfg.Height, imgSize, s3Key)
	if err != nil {
		return metadata, fmt.Errorf("failed to save avatar metadata: %w", err)
	}

	if err := u.dispatcher.AvatarUploaded(ctx, metadata.ID); err != nil {
		return metadata, fmt.Errorf("failed to publish message during avatar update: %w", err)
	}

	return metadata, nil
}
