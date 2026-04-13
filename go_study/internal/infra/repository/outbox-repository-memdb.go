package repository

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/model"
	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/infra/database"
)

type memDbOutboxRecord struct {
	Message model.OutboxMessage
}

func (r memDbOutboxRecord) Key() string {
	return r.Message.Id
}

type OutboxMemDbRepository struct {
	db *database.MemDbRepository[memDbOutboxRecord]
}

const outboxProcessingRecoveryWindow = 1 * time.Minute

var _ applicationRepository.OutboxRepository = (*OutboxMemDbRepository)(nil)

func NewOutboxMemDbRepository(db *database.MemDbRepository[memDbOutboxRecord]) OutboxMemDbRepository {
	return OutboxMemDbRepository{db: db}
}

func (r OutboxMemDbRepository) Enqueue(ctx context.Context, message *model.OutboxMessage) (*model.OutboxMessage, error) {
	if message == nil {
		return nil, fmt.Errorf("enqueue outbox message: message is nil")
	}

	now := time.Now().UTC()
	if message.Status == "" {
		message.Status = model.OutboxStatusPending
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = now
	}
	if message.UpdatedAt.IsZero() {
		message.UpdatedAt = now
	}
	if message.NextAttempt.IsZero() {
		message.NextAttempt = now
	}

	record := memDbOutboxRecord{Message: cloneOutboxMessage(*message)}
	saved, err := r.db.Save(ctx, &record)
	if err != nil {
		return nil, err
	}

	result := cloneOutboxMessage(saved.Message)
	return &result, nil
}

func (r OutboxMemDbRepository) ListPending(ctx context.Context, limit int, maxAttempts int, now time.Time) ([]model.OutboxMessage, error) {
	if limit <= 0 {
		return []model.OutboxMessage{}, nil
	}

	all := r.db.ListAll(ctx)
	pending := make([]model.OutboxMessage, 0, len(all))
	for _, record := range all {
		if !isPendingForClaim(record.Message, maxAttempts, now) {
			continue
		}
		pending = append(pending, cloneOutboxMessage(record.Message))
	}

	sort.Slice(pending, func(i, j int) bool {
		if pending[i].CreatedAt.Equal(pending[j].CreatedAt) {
			return pending[i].Id < pending[j].Id
		}
		return pending[i].CreatedAt.Before(pending[j].CreatedAt)
	})

	if len(pending) > limit {
		pending = pending[:limit]
	}

	claimedAt := time.Now().UTC()
	claimed := make([]model.OutboxMessage, 0, len(pending))
	for _, message := range pending {
		stored, err := r.db.FindById(ctx, message.Id)
		if err != nil {
			return nil, err
		}

		record := memDbOutboxRecord{Message: cloneOutboxMessage(stored.Message)}
		if !isPendingForClaim(record.Message, maxAttempts, now) {
			continue
		}

		record.Message.Status = model.OutboxStatusProcessing
		record.Message.Attempt++
		record.Message.UpdatedAt = claimedAt

		saved, err := r.db.Save(ctx, &record)
		if err != nil {
			return nil, err
		}

		claimed = append(claimed, cloneOutboxMessage(saved.Message))
	}

	return claimed, nil
}

func (r OutboxMemDbRepository) MarkAsPublished(ctx context.Context, id string, publishedAt time.Time) error {
	stored, err := r.db.FindById(ctx, id)
	if err != nil {
		return err
	}

	record := memDbOutboxRecord{Message: cloneOutboxMessage(stored.Message)}
	if record.Message.DeletedAt != nil || record.Message.PublishedAt != nil || record.Message.Status != model.OutboxStatusProcessing {
		return nil
	}

	record.Message.Status = model.OutboxStatusPublished
	record.Message.PublishedAt = ptrTime(publishedAt)
	record.Message.LastError = nil
	record.Message.UpdatedAt = publishedAt

	_, err = r.db.Save(ctx, &record)
	return err
}

func (r OutboxMemDbRepository) MarkAsError(ctx context.Context, id string, lastError string, nextAttempt time.Time) error {
	stored, err := r.db.FindById(ctx, id)
	if err != nil {
		return err
	}

	record := memDbOutboxRecord{Message: cloneOutboxMessage(stored.Message)}
	if record.Message.DeletedAt != nil || record.Message.PublishedAt != nil || record.Message.Status != model.OutboxStatusProcessing {
		return nil
	}

	status := model.OutboxStatusError
	if nextAttempt.IsZero() {
		status = model.OutboxStatusFailed
	}

	record.Message.Status = status
	record.Message.NextAttempt = nextAttempt
	record.Message.LastError = ptrString(lastError)
	record.Message.UpdatedAt = time.Now().UTC()

	_, err = r.db.Save(ctx, &record)
	return err
}

func isPendingForClaim(message model.OutboxMessage, maxAttempts int, now time.Time) bool {
	if message.DeletedAt != nil || message.PublishedAt != nil {
		return false
	}
	if maxAttempts > 0 && message.Attempt >= maxAttempts {
		return false
	}

	switch message.Status {
	case model.OutboxStatusPending, model.OutboxStatusError:
		return !message.NextAttempt.After(now)
	case model.OutboxStatusProcessing:
		if message.UpdatedAt.IsZero() {
			return true
		}

		reclaimBefore := now.Add(-outboxProcessingRecoveryWindow)
		return !message.UpdatedAt.After(reclaimBefore)
	default:
		return false
	}
}

func cloneOutboxMessage(message model.OutboxMessage) model.OutboxMessage {
	cloned := message
	if message.Payload != nil {
		cloned.Payload = append([]byte(nil), message.Payload...)
	}
	if message.Headers != nil {
		cloned.Headers = make(map[string]string, len(message.Headers))
		for key, value := range message.Headers {
			cloned.Headers[key] = value
		}
	}
	if message.PublishedAt != nil {
		cloned.PublishedAt = ptrTime(*message.PublishedAt)
	}
	if message.LastError != nil {
		cloned.LastError = ptrString(*message.LastError)
	}
	if message.DeletedAt != nil {
		cloned.DeletedAt = ptrTime(*message.DeletedAt)
	}

	return cloned
}

func ptrTime(value time.Time) *time.Time {
	v := value
	return &v
}

func ptrString(value string) *string {
	v := value
	return &v
}
