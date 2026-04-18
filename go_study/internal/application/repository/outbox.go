package repository

import (
	"context"
	"time"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"
	OutboxStatusPublished OutboxStatus = "published"
	OutboxStatusFailed    OutboxStatus = "failed"

	DefaultOutboxMaxAttempts = 5
)

type OutboxMessage struct {
	ID          string
	Topic       string
	Payload     string
	Status      OutboxStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	AvailableAt time.Time
	PublishedAt time.Time
	Attempts    int
	LastError   string
}

func (m OutboxMessage) Key() string {
	return m.ID
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, msg OutboxMessage) error
	ClaimNext(ctx context.Context, maxAttempts int) (*OutboxMessage, error)
	Update(ctx context.Context, msg OutboxMessage) error
}
