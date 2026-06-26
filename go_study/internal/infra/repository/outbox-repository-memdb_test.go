package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/outbox"
)

func TestOutboxMemDbRepository_Insert(t *testing.T) {
	repo := NewOutboxMemDbRepository()
	ctx := context.Background()

	msg := &outbox.OutboxMessage{
		ID:        "msg-1",
		Payload:   `{"hello":"world"}`,
		Topic:     "test-topic",
		Status:    outbox.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Insert(ctx, msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify stored value is a copy (modifying original doesn't affect stored)
	msg.Payload = "modified"
	val, ok := repo.data.Load("msg-1")
	if !ok {
		t.Fatal("expected message to be stored")
	}
	stored := val.(outbox.OutboxMessage)
	if stored.Payload != `{"hello":"world"}` {
		t.Errorf("expected original payload, got %q", stored.Payload)
	}
	if stored.ID != "msg-1" {
		t.Errorf("expected id msg-1, got %q", stored.ID)
	}
}

func TestOutboxMemDbRepository_FetchPendingBatch_Empty(t *testing.T) {
	repo := NewOutboxMemDbRepository()
	ctx := context.Background()

	batch, err := repo.FetchPendingBatch(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if batch == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(batch) != 0 {
		t.Errorf("expected 0 messages, got %d", len(batch))
	}
}

func TestOutboxMemDbRepository_FetchPendingBatch_SortsByCreatedAt(t *testing.T) {
	repo := NewOutboxMemDbRepository()
	ctx := context.Background()

	now := time.Now()
	msgs := []outbox.OutboxMessage{
		{ID: "msg-3", Status: outbox.StatusPending, CreatedAt: now.Add(3 * time.Second)},
		{ID: "msg-1", Status: outbox.StatusPending, CreatedAt: now.Add(1 * time.Second)},
		{ID: "msg-2", Status: outbox.StatusPending, CreatedAt: now.Add(2 * time.Second)},
	}

	for _, m := range msgs {
		repo.data.Store(m.ID, m)
	}

	batch, err := repo.FetchPendingBatch(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(batch) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(batch))
	}

	expectedOrder := []string{"msg-1", "msg-2", "msg-3"}
	for i, m := range batch {
		if m.ID != expectedOrder[i] {
			t.Errorf("position %d: expected %s, got %s", i, expectedOrder[i], m.ID)
		}
	}
}

func TestOutboxMemDbRepository_FetchPendingBatch_FiltersNonPending(t *testing.T) {
	repo := NewOutboxMemDbRepository()
	ctx := context.Background()

	now := time.Now()
	repo.data.Store("sent-msg", outbox.OutboxMessage{
		ID: "sent-msg", Status: outbox.StatusSent, CreatedAt: now,
	})
	repo.data.Store("error-msg", outbox.OutboxMessage{
		ID: "error-msg", Status: outbox.StatusError, CreatedAt: now,
	})
	repo.data.Store("pending-msg", outbox.OutboxMessage{
		ID: "pending-msg", Status: outbox.StatusPending, CreatedAt: now,
	})

	batch, err := repo.FetchPendingBatch(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(batch) != 1 {
		t.Fatalf("expected 1 pending message, got %d", len(batch))
	}
	if batch[0].ID != "pending-msg" {
		t.Errorf("expected pending-msg, got %s", batch[0].ID)
	}
}

func TestOutboxMemDbRepository_FetchPendingBatch_FiltersDeleted(t *testing.T) {
	repo := NewOutboxMemDbRepository()
	ctx := context.Background()

	now := time.Now()
	deletedAt := now.Add(-1 * time.Hour)
	repo.data.Store("deleted-msg", outbox.OutboxMessage{
		ID: "deleted-msg", Status: outbox.StatusPending, CreatedAt: now, DeletedAt: &deletedAt,
	})
	repo.data.Store("active-msg", outbox.OutboxMessage{
		ID: "active-msg", Status: outbox.StatusPending, CreatedAt: now,
	})

	batch, err := repo.FetchPendingBatch(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(batch) != 1 {
		t.Fatalf("expected 1 active message, got %d", len(batch))
	}
	if batch[0].ID != "active-msg" {
		t.Errorf("expected active-msg, got %s", batch[0].ID)
	}
}

func TestOutboxMemDbRepository_FetchPendingBatch_RespectsLimit(t *testing.T) {
	repo := NewOutboxMemDbRepository()
	ctx := context.Background()

	now := time.Now()
	for i := range 5 {
		repo.data.Store(string(rune('a'+i)), outbox.OutboxMessage{
			ID: string(rune('a' + i)), Status: outbox.StatusPending, CreatedAt: now.Add(time.Duration(i) * time.Second),
		})
	}

	batch, err := repo.FetchPendingBatch(ctx, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(batch) != 2 {
		t.Fatalf("expected 2 messages (limit), got %d", len(batch))
	}
}

func TestOutboxMemDbRepository_MarkSent(t *testing.T) {
	repo := NewOutboxMemDbRepository()
	ctx := context.Background()

	now := time.Now()
	msg := outbox.OutboxMessage{
		ID:        "msg-1",
		Status:    outbox.StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	repo.data.Store("msg-1", msg)

	err := repo.MarkSent(ctx, "msg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, _ := repo.data.Load("msg-1")
	updated := val.(outbox.OutboxMessage)
	if updated.Status != outbox.StatusSent {
		t.Errorf("expected status sent, got %q", updated.Status)
	}
	if updated.SentAt == nil {
		t.Error("expected sent_at to be set")
	}
	if updated.LastError != nil {
		t.Errorf("expected last_error to be nil, got %v", *updated.LastError)
	}
}

func TestOutboxMemDbRepository_MarkSent_NonExistent(t *testing.T) {
	repo := NewOutboxMemDbRepository()
	ctx := context.Background()

	err := repo.MarkSent(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOutboxMemDbRepository_MarkFailed(t *testing.T) {
	repo := NewOutboxMemDbRepository()
	ctx := context.Background()

	now := time.Now()
	msg := outbox.OutboxMessage{
		ID:             "msg-1",
		Status:         outbox.StatusPending,
		AttemptCounter: 1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	repo.data.Store("msg-1", msg)

	err := repo.MarkFailed(ctx, "msg-1", "connection refused", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, _ := repo.data.Load("msg-1")
	updated := val.(outbox.OutboxMessage)
	if updated.AttemptCounter != 2 {
		t.Errorf("expected attempt counter 2, got %d", updated.AttemptCounter)
	}
	if updated.LastError == nil || *updated.LastError != "connection refused" {
		t.Errorf("expected last_error 'connection refused', got %v", updated.LastError)
	}
	if updated.Status != outbox.StatusPending {
		t.Errorf("expected status pending (under maxAttempts), got %q", updated.Status)
	}
}

func TestOutboxMemDbRepository_MarkFailed_ExceedsMaxAttempts(t *testing.T) {
	repo := NewOutboxMemDbRepository()
	ctx := context.Background()

	now := time.Now()
	msg := outbox.OutboxMessage{
		ID:             "msg-1",
		Status:         outbox.StatusPending,
		AttemptCounter: 2,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	repo.data.Store("msg-1", msg)

	err := repo.MarkFailed(ctx, "msg-1", "timeout", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, _ := repo.data.Load("msg-1")
	updated := val.(outbox.OutboxMessage)
	if updated.AttemptCounter != 3 {
		t.Errorf("expected attempt counter 3, got %d", updated.AttemptCounter)
	}
	if updated.Status != outbox.StatusError {
		t.Errorf("expected status error (reached maxAttempts), got %q", updated.Status)
	}
}

func TestOutboxMemDbRepository_MarkFailed_NonExistent(t *testing.T) {
	repo := NewOutboxMemDbRepository()
	ctx := context.Background()

	err := repo.MarkFailed(ctx, "nonexistent", "error", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOutboxMemDbRepository_ConcurrentAccess(t *testing.T) {
	repo := NewOutboxMemDbRepository()
	ctx := context.Background()

	now := time.Now()
	repo.data.Store("shared", outbox.OutboxMessage{
		ID: "shared", Status: outbox.StatusPending, CreatedAt: now, UpdatedAt: now,
	})

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = repo.MarkSent(ctx, "shared")
		}()
	}
	wg.Wait()

	val, _ := repo.data.Load("shared")
	final := val.(outbox.OutboxMessage)
	if final.Status != outbox.StatusSent {
		t.Errorf("expected status sent after concurrent marks, got %q", final.Status)
	}
}
