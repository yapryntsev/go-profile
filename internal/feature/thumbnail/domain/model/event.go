package model

import "github.com/google/uuid"

type Event interface{}
type UploadEvent struct {
	AvatarID uuid.UUID
}

type DeleteEvent struct {
	AvatarID uuid.UUID
}
