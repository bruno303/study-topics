package repository

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
)

type memDbOutboxStorage struct {
	mu   sync.Mutex
	data map[string]applicationRepository.OutboxMessage
}

type OutboxMemDbRepository struct {
	storage *memDbOutboxStorage
}

func newMemDbOutboxStorage() *memDbOutboxStorage {
	return &memDbOutboxStorage{
		data: map[string]applicationRepository.OutboxMessage{},
	}
}

func NewOutboxMemDbRepository(storage *memDbOutboxStorage) applicationRepository.OutboxRepository {
	if storage == nil {
		storage = newMemDbOutboxStorage()
	}
	return &OutboxMemDbRepository{storage: storage}
}

func (r *OutboxMemDbRepository) Enqueue(ctx context.Context, msg applicationRepository.OutboxMessage) error {
	msg = applyOutboxDefaults(msg)

	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()
	r.storage.data[msg.ID] = msg
	return nil
}

func (r *OutboxMemDbRepository) ClaimNext(ctx context.Context, maxAttempts int) (*applicationRepository.OutboxMessage, error) {
	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()

	keys := make([]string, 0, len(r.storage.data))
	for key := range r.storage.data {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left := r.storage.data[keys[i]]
		right := r.storage.data[keys[j]]
		if left.CreatedAt.Equal(right.CreatedAt) {
			return left.ID < right.ID
		}
		return left.CreatedAt.Before(right.CreatedAt)
	})

	now := time.Now()
	for _, key := range keys {
		msg := r.storage.data[key]
		if msg.Status != applicationRepository.OutboxStatusPending {
			continue
		}
		if msg.AvailableAt.After(now) {
			continue
		}
		if msg.Attempts >= maxAttempts {
			continue
		}
		copyMsg := msg
		return &copyMsg, nil
	}

	return nil, nil
}

func (r *OutboxMemDbRepository) Update(ctx context.Context, msg applicationRepository.OutboxMessage) error {
	msg = applyOutboxDefaults(msg)

	r.storage.mu.Lock()
	defer r.storage.mu.Unlock()
	if _, ok := r.storage.data[msg.ID]; !ok {
		return errors.New("outbox message not found")
	}
	r.storage.data[msg.ID] = msg
	return nil
}
