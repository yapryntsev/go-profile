package model

import "errors"

type FormatType string

const (
	JPEG FormatType = "application/jpeg"
	PNG  FormatType = "application/png"
	WebP FormatType = "application/webp"
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
