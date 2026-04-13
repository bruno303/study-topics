package repository

import (
	"context"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/model"
)

//go:generate go tool mockgen -source=outbox.go -destination=outbox_mocks.go -package repository

type OutboxRepository interface {
	Enqueue(ctx context.Context, message *model.OutboxMessage) (*model.OutboxMessage, error)
	ListPending(ctx context.Context, limit int, maxAttempts int, now time.Time) ([]model.OutboxMessage, error)
	MarkAsPublished(ctx context.Context, id string, publishedAt time.Time) error
	MarkAsError(ctx context.Context, id string, lastError string, nextAttempt time.Time) error
}
