package outbox

import (
	"context"
)

type OutboxRepository interface {
	Insert(ctx context.Context, msg *OutboxMessage) error
	FetchPendingBatch(ctx context.Context, limit int) ([]*OutboxMessage, error)
	MarkSent(ctx context.Context, msgID string) error
	MarkFailed(ctx context.Context, msgID string, lastError string, maxAttempts int) error
}
