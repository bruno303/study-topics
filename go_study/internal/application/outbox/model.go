package outbox

import "time"

type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusError   Status = "error"
)

type OutboxMessage struct {
	ID             string
	Payload        string
	Topic          string
	Headers        string
	Status         Status
	AttemptCounter int
	LastError      *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	SentAt         *time.Time
	DeletedAt      *time.Time
}
