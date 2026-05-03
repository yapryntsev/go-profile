package model

import "errors"

type FormatType string

const (
	JPEG FormatType = "image/jpeg"
	PNG  FormatType = "image/png"
	WebP FormatType = "image/webp"
)

func NewFormatType(s string) (FormatType, error) {
	switch s {
	case "jpeg":
		return JPEG, nil
	case "png":
		return PNG, nil
	case "webp":
		return WebP, nil
	default:
		return "", errors.New("unknown format type")
	}
}
