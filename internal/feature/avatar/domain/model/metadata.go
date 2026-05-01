package model

import (
	"time"

	"github.com/google/uuid"
)

type ProcessingStatus string

const (
	ProcessingPending   ProcessingStatus = "pending"
	ProcessingActive    ProcessingStatus = "active"
	ProcessingCompleted ProcessingStatus = "completed"
)

type Metadata struct {
	ID     uuid.UUID
	UserID string

	FileName  string
	MimeType  string
	SizeBytes int
	Height    int
	Width     int

	ProcessingStatus ProcessingStatus

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	S3Key           string
	S3ThumbnailKeys *S3ThumbnailKeys
}

type S3ThumbnailKeys = []struct {
	Size string
	URL  string
}

func NewMetadata(
	id uuid.UUID,
	userID string,
	s3key string,
	fileName string,
	mimeType string,
	size int,
	s3thumbnailKeys *S3ThumbnailKeys,
) Metadata {
	return Metadata{
		ID:               id,
		UserID:           userID,
		FileName:         fileName,
		MimeType:         mimeType,
		SizeBytes:        size,
		ProcessingStatus: ProcessingPending,
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
		DeletedAt:        nil,
		S3Key:            s3key,
		S3ThumbnailKeys:  s3thumbnailKeys,
	}
}
