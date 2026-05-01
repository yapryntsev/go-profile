package model

import (
	"errors"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

type FormatType string

const (
	JPEG FormatType = "jpeg"
	PNG  FormatType = "png"
	WebP FormatType = "webp"
)

func NewFormatType(s string) (FormatType, error) {
	switch s {
	case string(JPEG):
		return JPEG, nil
	case string(PNG):
		return PNG, nil
	case string(WebP):
		return WebP, nil
	default:
		return "", errors.New("unknown format type")
	}
}
