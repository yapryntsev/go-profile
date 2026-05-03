package model

import "errors"

type AspectRatio string

const (
	S100x100 AspectRatio = "100x100"
	S300x300 AspectRatio = "300x300"
	Origin   AspectRatio = "origin"
)

func NewAspectRatio(s string) (AspectRatio, error) {
	switch s {
	case string(S100x100):
		return S100x100, nil
	case string(S300x300):
		return S300x300, nil
	case string(Origin):
		return Origin, nil
	default:
		return "", errors.New("unknown size type")
	}
}
