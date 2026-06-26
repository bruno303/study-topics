package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/outbox"
)

type OutboxMemDbRepository struct {
	data *sync.Map
}

var _ outbox.OutboxRepository = (*OutboxMemDbRepository)(nil)

func NewOutboxMemDbRepository() *OutboxMemDbRepository {
	return &OutboxMemDbRepository{data: &sync.Map{}}
}

func (r *OutboxMemDbRepository) Insert(_ context.Context, msg *outbox.OutboxMessage) error {
	r.data.Store(msg.ID, *msg)
	return nil
}

func (r *OutboxMemDbRepository) FetchPendingBatch(_ context.Context, limit int) ([]*outbox.OutboxMessage, error) {
	var pending []*outbox.OutboxMessage

	r.data.Range(func(_, value any) bool {
		m, ok := value.(outbox.OutboxMessage)
		if !ok {
			return true
		}
		if m.Status == outbox.StatusPending && m.DeletedAt == nil {
			cp := m
			pending = append(pending, &cp)
		}
		return true
	})

	sort.Slice(pending, func(i, j int) bool {
		return pending[i].CreatedAt.Before(pending[j].CreatedAt)
	})

	if limit > 0 && len(pending) > limit {
		pending = pending[:limit]
	}

	if pending == nil {
		pending = make([]*outbox.OutboxMessage, 0)
	}

	return pending, nil
}

func (r *OutboxMemDbRepository) MarkSent(_ context.Context, msgID string) error {
	val, ok := r.data.Load(msgID)
	if !ok {
		return nil
	}

	m, ok := val.(outbox.OutboxMessage)
	if !ok {
		return nil
	}

	now := time.Now()
	m.Status = outbox.StatusSent
	m.SentAt = &now
	m.UpdatedAt = now
	m.LastError = nil

	r.data.Store(msgID, m)
	return nil
}

func (r *OutboxMemDbRepository) MarkFailed(_ context.Context, msgID string, lastError string, maxAttempts int) error {
	val, ok := r.data.Load(msgID)
	if !ok {
		return nil
	}

	m, ok := val.(outbox.OutboxMessage)
	if !ok {
		return nil
	}

	m.AttemptCounter++
	m.LastError = &lastError
	m.UpdatedAt = time.Now()
	if m.AttemptCounter >= maxAttempts {
		m.Status = outbox.StatusError
	}

	r.data.Store(msgID, m)
	return nil
}
