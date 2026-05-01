package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"goph-profile/internal/api"
	"goph-profile/internal/feature/avatar/domain"
	"goph-profile/internal/feature/avatar/domain/model"
	"io"
	"mime/multipart"
	"strconv"
)

type Handler struct {
	avatar   domain.UseCaseGetAvatar
	metadata domain.UseCaseGetAvatarMetadata
	delete   domain.UseCaseDeleteAvatar
	upload   domain.UseCaseUploadAvatarMetadata
}

func NewHandler(
	avatar domain.UseCaseGetAvatar,
	metadata domain.UseCaseGetAvatarMetadata,
	delete domain.UseCaseDeleteAvatar,
	upload domain.UseCaseUploadAvatarMetadata,
) Handler {
	return Handler{
		avatar:   avatar,
		metadata: metadata,
		delete:   delete,
		upload:   upload,
	}
}

func (h Handler) PostApiV1Avatars(
	ctx context.Context,
	request api.PostApiV1AvatarsRequestObject,
) (api.PostApiV1AvatarsResponseObject, error) {
	part, err := request.Body.NextPart()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("no file provided")
		}
		return nil, err
	}

	name := part.FileName()
	metadata, err := h.upload.Run(ctx, request.Params.XUserID, name, part)
	if err != nil {
		if errors.Is(err, model.ErrUploadAvatarTooLarge) {
			return api.PostApiV1Avatars400JSONResponse{
				Details: strconv.Itoa(model.AvatarMaxSizeBytes),
				Error:   "File too large",
			}, nil
		} else if errors.Is(err, model.ErrUploadAvatarUnknown) {
			return api.PostApiV1Avatars400JSONResponse{
				Details: "Supported formats: jpeg, png, webp",
				Error:   "Invalid file format",
			}, nil
		}

		return nil, fmt.Errorf("failed to get avatar metadata: %w", err)
	}

	return api.PostApiV1Avatars201JSONResponse{
		CreatedAt: metadata.CreatedAt,
		Id:        metadata.ID,
		Status:    string(metadata.ProcessingStatus),
		Url:       metadata.S3Key,
		UserId:    metadata.UserID,
	}, nil
}

func (h Handler) DeleteApiV1AvatarsId(
	ctx context.Context,
	request api.DeleteApiV1AvatarsIdRequestObject,
) (api.DeleteApiV1AvatarsIdResponseObject, error) {
	err := h.delete.Run(ctx, request.Params.XUserID, request.Id)
	if err != nil {
		if errors.Is(err, model.ErrDeleteAvatarNotFound) {
			return api.DeleteApiV1AvatarsId403JSONResponse{
				Details: "you can only delete your own avatars",
				Error:   "forbidden",
			}, nil
		}

		return nil, fmt.Errorf("failed to delete avatar: %w", err)
	}

	return api.DeleteApiV1AvatarsId200Response{}, nil
}

func (h Handler) GetApiV1AvatarsId(
	ctx context.Context,
	request api.GetApiV1AvatarsIdRequestObject,
) (api.GetApiV1AvatarsIdResponseObject, error) {
	img, name, err := h.avatar.Run(ctx, request.Id, request.Params.Format, request.Params.Size)
	if err != nil {
		if errors.Is(err, model.ErrGetAvatarNotFound) {
			return api.GetApiV1AvatarsId404JSONResponse{Error: "Avatar not found"}, nil
		}

		return nil, fmt.Errorf("failed to get avatar: %w", err)
	}

	return api.GetApiV1AvatarsId200MultipartResponse(
		func(writer *multipart.Writer) error {
			part, err := writer.CreateFormFile("avatar", fmt.Sprintf("%s.png", name))
			if err != nil {
				return err
			}

			_, err = io.Copy(part, bytes.NewReader(img))
			return err
		},
	), nil
}

func (h Handler) GetApiV1AvatarsIdMetadata(
	ctx context.Context,
	request api.GetApiV1AvatarsIdMetadataRequestObject,
) (api.GetApiV1AvatarsIdMetadataResponseObject, error) {
	metadata, err := h.metadata.Run(ctx, request.Id)
	if err != nil {
		if errors.Is(err, model.ErrGetAvatarMetadataNotFound) {
			return api.GetApiV1AvatarsIdMetadata404JSONResponse{
				Error: "forbidden",
			}, nil
		}

		return nil, fmt.Errorf("failed to get avatar metadata: %w", err)
	}

	var mappedThumbnails []struct {
		Size string `json:"size"`
		Url  string `json:"url"`
	}

	if metadata.S3ThumbnailKeys != nil {
		mappedThumbnails = make(
			[]struct {
				Size string `json:"size"`
				Url  string `json:"url"`
			}, len(*metadata.S3ThumbnailKeys),
		)

		for i, item := range *metadata.S3ThumbnailKeys {
			mappedThumbnails[i] = struct {
				Size string `json:"size"`
				Url  string `json:"url"`
			}{
				Size: item.Size,
				Url:  item.URL,
			}
		}
	}

	return api.GetApiV1AvatarsIdMetadata200JSONResponse{
		CreatedAt: metadata.CreatedAt,
		Dimensions: struct {
			Height int `json:"height"`
			Width  int `json:"width"`
		}{
			Height: metadata.Height,
			Width:  metadata.Width,
		},
		FileName:   metadata.FileName,
		Id:         metadata.ID,
		MimeType:   metadata.MimeType,
		Size:       metadata.SizeBytes,
		Thumbnails: &mappedThumbnails,
		UpdatedAt:  metadata.UpdatedAt,
		UserId:     metadata.UserID,
	}, nil
}
