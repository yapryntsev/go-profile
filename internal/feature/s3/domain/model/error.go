package model

import "errors"

var (
	ErrUploadUnknownMIME = errors.New("upload avatar: unknown mime type")
)
