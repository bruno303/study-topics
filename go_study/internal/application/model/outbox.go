package model

import "time"

type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "pending"
	OutboxStatusProcessing OutboxStatus = "processing"
	OutboxStatusPublished  OutboxStatus = "published"
	OutboxStatusError      OutboxStatus = "error"
	OutboxStatusFailed     OutboxStatus = "failed"
)

type OutboxMessage struct {
	Id          string
	Topic       string
	MessageKey  string
	Payload     []byte
	Headers     map[string]string
	Status      OutboxStatus
	Attempt     int
	NextAttempt time.Time
	PublishedAt *time.Time
	LastError   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
