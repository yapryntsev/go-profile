package domain

import (
	"context"
	"errors"
	"fmt"
	"goph-profile/internal/feature/thumbnail/domain/model"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/google/uuid"
	"golang.org/x/image/draw"

	_ "golang.org/x/image/webp"
)

type (
	MakeThumbnailAvatarRepo interface {
		Download(ctx context.Context, key string) (image.Image, error)
		Upload(ctx context.Context, key string, img image.Image) (string, error)
	}
	MakeThumbnailMetadataRepo interface {
		FetchS3Key(ctx context.Context, avatarID uuid.UUID) (string, error)
		Update(ctx context.Context, avatarID uuid.UUID, thumbnails []model.Thumbnail) error
	}
)

type UseCaseMakeThumbnails struct {
	metadataRepo MakeThumbnailMetadataRepo
	avatarRepo   MakeThumbnailAvatarRepo
}

func NewMakeThumbnails(
	metadataRepo MakeThumbnailMetadataRepo,
	avatarRepo MakeThumbnailAvatarRepo,
) UseCaseMakeThumbnails {
	return UseCaseMakeThumbnails{metadataRepo: metadataRepo, avatarRepo: avatarRepo}
}

func (u UseCaseMakeThumbnails) Run(ctx context.Context, avatarID uuid.UUID) error {
	s3key, err := u.metadataRepo.FetchS3Key(ctx, avatarID)
	if err != nil {
		if errors.Is(err, model.ErrAvatarDeleted) {
			return nil
		}
		return fmt.Errorf("failed to fetch s3 key: %w", err)
	}

	img, err := u.avatarRepo.Download(ctx, s3key)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}

	small := resizeImage(img, 100, 100)
	smallS3Key, err := u.avatarRepo.Upload(ctx, fmt.Sprintf("%s_%s", s3key, "100x100"), small)
	if err != nil {
		return fmt.Errorf("failed to upload small thumbnail: %w", err)
	}

	medium := resizeImage(img, 300, 300)
	mediumS3Key, err := u.avatarRepo.Upload(ctx, fmt.Sprintf("%s_%s", s3key, "300x300"), medium)
	if err != nil {
		return fmt.Errorf("failed to upload medium thumbnail: %w", err)
	}

	if err = u.metadataRepo.Update(
		ctx,
		avatarID,
		[]model.Thumbnail{
			{Size: "100x100", URL: smallS3Key},
			{Size: "300x300", URL: mediumS3Key},
		},
	); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}

func resizeImage(src image.Image, width, height int) image.Image {
	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(canvas, canvas.Bounds(), src, src.Bounds(), draw.Over, nil)
	return canvas
}
