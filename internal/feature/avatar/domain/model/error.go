package model

import "errors"

var (
	ErrDeleteAvatarNotFound      = errors.New("delete avatar: entry with provided avatarID & userID does not exist")
	ErrGetAvatarNotFound         = errors.New("get avatar: entry with provided avatarID does not exist")
	ErrGetAvatarMetadataNotFound = errors.New("get avatar metadata: entry with provided avatarID does not exist")
	ErrUploadAvatarUnknown       = errors.New("upload avatar: unknown type")
	ErrUploadAvatarTooLarge      = errors.New("upload avatar: file too large")
)
