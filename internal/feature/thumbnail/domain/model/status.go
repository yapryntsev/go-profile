package model

type ProcessingStatus string

const (
	ProcessingPending   ProcessingStatus = "pending"
	ProcessingActive    ProcessingStatus = "active"
	ProcessingCompleted ProcessingStatus = "completed"
)
