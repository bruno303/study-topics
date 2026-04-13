package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/model"
	"github.com/bruno303/study-topics/go-study/tests/integration"
)

func TestOutboxPgxRepository_ListPending_ClaimsByCreatedAtOrder(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	repo := NewOutboxPgxRepository(pool)
	now := time.Now().UTC()

	messages := []model.OutboxMessage{
		{Id: fmt.Sprintf("outbox-3-%d", now.UnixNano()), Topic: "topic", MessageKey: "k3", Payload: []byte("3"), Status: model.OutboxStatusPending, Attempt: 0, NextAttempt: now, CreatedAt: now.Add(3 * time.Second), UpdatedAt: now},
		{Id: fmt.Sprintf("outbox-1-%d", now.UnixNano()), Topic: "topic", MessageKey: "k1", Payload: []byte("1"), Status: model.OutboxStatusPending, Attempt: 1, NextAttempt: now, CreatedAt: now.Add(1 * time.Second), UpdatedAt: now},
		{Id: fmt.Sprintf("outbox-2-%d", now.UnixNano()), Topic: "topic", MessageKey: "k2", Payload: []byte("2"), Status: model.OutboxStatusError, Attempt: 2, NextAttempt: now, CreatedAt: now.Add(2 * time.Second), UpdatedAt: now},
	}

	for _, message := range messages {
		msg := message
		if _, err := repo.Enqueue(context.Background(), &msg); err != nil {
			t.Fatalf("failed to enqueue outbox message %s: %v", message.Id, err)
		}
	}

	claimed, err := repo.ListPending(context.Background(), 2, 5, now)
	if err != nil {
		t.Fatalf("expected claim without error, got %v", err)
	}
	if len(claimed) != 2 {
		t.Fatalf("expected two claimed messages, got %d", len(claimed))
	}
	if claimed[0].MessageKey != "k1" || claimed[1].MessageKey != "k2" {
		t.Fatalf("expected claim order [k1, k2], got [%s, %s]", claimed[0].MessageKey, claimed[1].MessageKey)
	}
	if claimed[0].Status != model.OutboxStatusProcessing || claimed[1].Status != model.OutboxStatusProcessing {
		t.Fatalf("expected claimed status processing, got [%s, %s]", claimed[0].Status, claimed[1].Status)
	}
	if claimed[0].Attempt != 2 || claimed[1].Attempt != 3 {
		t.Fatalf("expected incremented attempts [2,3], got [%d,%d]", claimed[0].Attempt, claimed[1].Attempt)
	}
}

func TestOutboxPgxRepository_ListPending_WhenClaimedByAnotherTransaction_SkipsLockedRows(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	now := time.Now().UTC()
	seedRepo := NewOutboxPgxRepository(pool)

	first := model.OutboxMessage{Id: fmt.Sprintf("claim-first-%d", now.UnixNano()), Topic: "topic", MessageKey: "first", Payload: []byte("a"), Status: model.OutboxStatusPending, NextAttempt: now, CreatedAt: now, UpdatedAt: now}
	second := model.OutboxMessage{Id: fmt.Sprintf("claim-second-%d", now.UnixNano()), Topic: "topic", MessageKey: "second", Payload: []byte("b"), Status: model.OutboxStatusPending, NextAttempt: now, CreatedAt: now.Add(time.Second), UpdatedAt: now}

	if _, err := seedRepo.Enqueue(context.Background(), &first); err != nil {
		t.Fatalf("failed to seed first outbox message: %v", err)
	}
	if _, err := seedRepo.Enqueue(context.Background(), &second); err != nil {
		t.Fatalf("failed to seed second outbox message: %v", err)
	}

	tx1, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("failed to begin first transaction: %v", err)
	}
	defer tx1.Rollback(context.Background())
	tx2, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("failed to begin second transaction: %v", err)
	}
	defer tx2.Rollback(context.Background())

	repo1 := newOutboxPgxRepository(pool, tx1)
	repo2 := newOutboxPgxRepository(pool, tx2)

	claimedByTx1, err := repo1.ListPending(context.Background(), 1, 5, now)
	if err != nil {
		t.Fatalf("expected tx1 claim without error, got %v", err)
	}
	if len(claimedByTx1) != 1 {
		t.Fatalf("expected tx1 to claim one message, got %d", len(claimedByTx1))
	}

	claimedByTx2, err := repo2.ListPending(context.Background(), 1, 5, now)
	if err != nil {
		t.Fatalf("expected tx2 claim without error, got %v", err)
	}
	if len(claimedByTx2) != 1 {
		t.Fatalf("expected tx2 to claim one message, got %d", len(claimedByTx2))
	}
	if claimedByTx1[0].Id == claimedByTx2[0].Id {
		t.Fatalf("expected different messages across transactions, both claimed %s", claimedByTx1[0].Id)
	}
}

func TestOutboxPgxRepository_Enqueue_WhenMinimalFieldsProvided_DefaultsAndClaimsAsPending(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	repo := NewOutboxPgxRepository(pool)
	message := model.OutboxMessage{
		Id:         fmt.Sprintf("minimal-pgx-%d", time.Now().UTC().UnixNano()),
		Topic:      "topic",
		MessageKey: "key",
		Payload:    []byte(`{"event":"minimal"}`),
	}

	saved, err := repo.Enqueue(context.Background(), &message)
	if err != nil {
		t.Fatalf("expected enqueue without error, got %v", err)
	}
	if saved.Status != model.OutboxStatusPending {
		t.Fatalf("expected default status pending, got %s", saved.Status)
	}
	if saved.CreatedAt.IsZero() || saved.UpdatedAt.IsZero() || saved.NextAttempt.IsZero() {
		t.Fatalf("expected created_at, updated_at and next_attempt to be defaulted, got %+v", *saved)
	}

	claimed, err := repo.ListPending(context.Background(), 1, 5, saved.NextAttempt)
	if err != nil {
		t.Fatalf("expected list pending without error, got %v", err)
	}
	if len(claimed) != 1 {
		t.Fatalf("expected one claimed message, got %d", len(claimed))
	}
	if claimed[0].Id != saved.Id {
		t.Fatalf("expected claimed id %s, got %s", saved.Id, claimed[0].Id)
	}
	if claimed[0].Status != model.OutboxStatusProcessing {
		t.Fatalf("expected claimed status processing, got %s", claimed[0].Status)
	}
	if claimed[0].Attempt != 1 {
		t.Fatalf("expected claimed attempt incremented to 1, got %d", claimed[0].Attempt)
	}
}

func TestOutboxPgxRepository_ListPending_WhenProcessingIsStale_ReclaimsOnlyStaleMessages(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	repo := NewOutboxPgxRepository(pool)
	now := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)
	maxAttempts := 3

	stale := model.OutboxMessage{
		Id:          fmt.Sprintf("stale-processing-%d", time.Now().UTC().UnixNano()),
		Topic:       "topic",
		MessageKey:  "stale",
		Payload:     []byte(`{"kind":"stale"}`),
		Status:      model.OutboxStatusProcessing,
		Attempt:     2,
		NextAttempt: now.Add(10 * time.Minute),
		CreatedAt:   now.Add(-2 * time.Minute),
		UpdatedAt:   now.Add(-outboxProcessingRecoveryWindow - time.Second),
	}
	exhausted := model.OutboxMessage{
		Id:          fmt.Sprintf("exhausted-processing-%d", time.Now().UTC().UnixNano()),
		Topic:       "topic",
		MessageKey:  "exhausted",
		Payload:     []byte(`{"kind":"exhausted"}`),
		Status:      model.OutboxStatusProcessing,
		Attempt:     maxAttempts,
		NextAttempt: now.Add(10 * time.Minute),
		CreatedAt:   now.Add(-90 * time.Second),
		UpdatedAt:   now.Add(-outboxProcessingRecoveryWindow - time.Second),
	}
	fresh := model.OutboxMessage{
		Id:          fmt.Sprintf("fresh-processing-%d", time.Now().UTC().UnixNano()),
		Topic:       "topic",
		MessageKey:  "fresh",
		Payload:     []byte(`{"kind":"fresh"}`),
		Status:      model.OutboxStatusProcessing,
		Attempt:     1,
		NextAttempt: now.Add(10 * time.Minute),
		CreatedAt:   now.Add(-time.Minute),
		UpdatedAt:   now.Add(-outboxProcessingRecoveryWindow + time.Second),
	}

	for _, message := range []model.OutboxMessage{stale, exhausted, fresh} {
		msg := message
		if _, err := repo.Enqueue(context.Background(), &msg); err != nil {
			t.Fatalf("failed to seed outbox message %s: %v", message.Id, err)
		}
	}

	claimed, err := repo.ListPending(context.Background(), 10, maxAttempts, now)
	if err != nil {
		t.Fatalf("expected list pending without error, got %v", err)
	}
	if len(claimed) != 1 {
		t.Fatalf("expected one stale message to be reclaimed, got %d", len(claimed))
	}
	if claimed[0].Id != stale.Id {
		t.Fatalf("expected stale message %s to be reclaimed, got %s", stale.Id, claimed[0].Id)
	}
	if claimed[0].Attempt != 3 {
		t.Fatalf("expected stale attempt incremented to 3, got %d", claimed[0].Attempt)
	}

	claimedAgain, err := repo.ListPending(context.Background(), 10, maxAttempts, now)
	if err != nil {
		t.Fatalf("expected second list pending without error, got %v", err)
	}
	if len(claimedAgain) != 0 {
		t.Fatalf("expected no additional claimable messages, got %d", len(claimedAgain))
	}
}

func TestOutboxPgxRepository_MarkAsError_WhenNextAttemptProvided_SetsRetryableStatus(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	repo := NewOutboxPgxRepository(pool)
	now := time.Now().UTC()
	id := fmt.Sprintf("retry-pgx-%d", now.UnixNano())

	_, err := repo.Enqueue(context.Background(), &model.OutboxMessage{
		Id:          id,
		Topic:       "topic",
		MessageKey:  "retry",
		Payload:     []byte(`{"kind":"retry"}`),
		Status:      model.OutboxStatusProcessing,
		Attempt:     1,
		NextAttempt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("failed to seed outbox message: %v", err)
	}

	nextAttempt := now.Add(time.Minute)
	if err := repo.MarkAsError(context.Background(), id, "temporary failure", nextAttempt); err != nil {
		t.Fatalf("expected mark as error without error, got %v", err)
	}

	var status string
	var dbNextAttempt time.Time
	var lastError *string
	if err := pool.QueryRow(context.Background(), "SELECT STATUS, NEXT_ATTEMPT, LAST_ERROR FROM OUTBOX WHERE ID = $1", id).Scan(&status, &dbNextAttempt, &lastError); err != nil {
		t.Fatalf("failed to query outbox row: %v", err)
	}
	if status != string(model.OutboxStatusError) {
		t.Fatalf("expected status error, got %s", status)
	}
	if !dbNextAttempt.Equal(nextAttempt) && !dbNextAttempt.Equal(nextAttempt.Truncate(time.Microsecond)) {
		t.Fatalf("expected next_attempt %v, got %v", nextAttempt, dbNextAttempt)
	}
	if lastError == nil || *lastError != "temporary failure" {
		t.Fatalf("expected last_error temporary failure, got %v", lastError)
	}

	claimed, err := repo.ListPending(context.Background(), 1, 5, nextAttempt)
	if err != nil {
		t.Fatalf("expected list pending without error, got %v", err)
	}
	if len(claimed) != 1 || claimed[0].Id != id {
		t.Fatalf("expected claimed retry row %s, got %+v", id, claimed)
	}
}

func TestOutboxPgxRepository_MarkAsError_WhenNextAttemptIsZero_MarksAsFailed(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	repo := NewOutboxPgxRepository(pool)
	now := time.Now().UTC()
	id := fmt.Sprintf("failed-pgx-%d", now.UnixNano())

	_, err := repo.Enqueue(context.Background(), &model.OutboxMessage{
		Id:          id,
		Topic:       "topic",
		MessageKey:  "failed",
		Payload:     []byte(`{"kind":"failed"}`),
		Status:      model.OutboxStatusProcessing,
		Attempt:     3,
		NextAttempt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("failed to seed outbox message: %v", err)
	}

	if err := repo.MarkAsError(context.Background(), id, "terminal failure", time.Time{}); err != nil {
		t.Fatalf("expected mark as error without error, got %v", err)
	}

	var status string
	if err := pool.QueryRow(context.Background(), "SELECT STATUS FROM OUTBOX WHERE ID = $1", id).Scan(&status); err != nil {
		t.Fatalf("failed to query outbox row: %v", err)
	}
	if status != string(model.OutboxStatusFailed) {
		t.Fatalf("expected status failed, got %s", status)
	}

	claimed, err := repo.ListPending(context.Background(), 1, 5, now.Add(time.Hour))
	if err != nil {
		t.Fatalf("expected list pending without error, got %v", err)
	}
	if len(claimed) != 0 {
		t.Fatalf("expected failed message not to be claimable, got %d", len(claimed))
	}
}

func TestOutboxPgxRepository_MarkAsError_WhenAlreadyPublished_DoesNotRevertToError(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	repo := NewOutboxPgxRepository(pool)
	now := time.Now().UTC()
	id := fmt.Sprintf("published-pgx-%d", now.UnixNano())
	publishedAt := now.Add(10 * time.Second)

	_, err := repo.Enqueue(context.Background(), &model.OutboxMessage{
		Id:          id,
		Topic:       "topic",
		MessageKey:  "published",
		Payload:     []byte(`{"kind":"published"}`),
		Status:      model.OutboxStatusProcessing,
		Attempt:     2,
		NextAttempt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("failed to seed outbox message: %v", err)
	}

	if err := repo.MarkAsPublished(context.Background(), id, publishedAt); err != nil {
		t.Fatalf("expected mark as published without error, got %v", err)
	}
	if err := repo.MarkAsError(context.Background(), id, "late failure", now.Add(time.Minute)); err != nil {
		t.Fatalf("expected mark as error to be no-op for published row, got %v", err)
	}

	var status string
	var dbPublishedAt *time.Time
	var lastError *string
	if err := pool.QueryRow(context.Background(), "SELECT STATUS, PUBLISHED_AT, LAST_ERROR FROM OUTBOX WHERE ID = $1", id).Scan(&status, &dbPublishedAt, &lastError); err != nil {
		t.Fatalf("failed to query outbox row: %v", err)
	}
	if status != string(model.OutboxStatusPublished) {
		t.Fatalf("expected status published, got %s", status)
	}
	if dbPublishedAt == nil || (!dbPublishedAt.Equal(publishedAt) && !dbPublishedAt.Equal(publishedAt.Truncate(time.Microsecond))) {
		t.Fatalf("expected published_at %v, got %v", publishedAt, dbPublishedAt)
	}
	if lastError != nil {
		t.Fatalf("expected last_error to remain nil, got %v", *lastError)
	}
}

func TestOutboxPgxRepository_MarkAsPublished_WhenStatusIsNotProcessing_IsNoOp(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	repo := NewOutboxPgxRepository(pool)
	now := time.Now().UTC()
	id := fmt.Sprintf("pending-published-noop-%d", now.UnixNano())

	_, err := repo.Enqueue(context.Background(), &model.OutboxMessage{
		Id:          id,
		Topic:       "topic",
		MessageKey:  "pending",
		Payload:     []byte(`{"kind":"pending"}`),
		Status:      model.OutboxStatusPending,
		NextAttempt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("failed to seed outbox message: %v", err)
	}

	if err := repo.MarkAsPublished(context.Background(), id, now.Add(time.Minute)); err != nil {
		t.Fatalf("expected mark as published attempt without error, got %v", err)
	}

	var status string
	var publishedAt *time.Time
	if err := pool.QueryRow(context.Background(), "SELECT STATUS, PUBLISHED_AT FROM OUTBOX WHERE ID = $1", id).Scan(&status, &publishedAt); err != nil {
		t.Fatalf("failed to query outbox row: %v", err)
	}
	if status != string(model.OutboxStatusPending) {
		t.Fatalf("expected status pending after no-op, got %s", status)
	}
	if publishedAt != nil {
		t.Fatalf("expected published_at to remain nil, got %v", *publishedAt)
	}

	claimed, err := repo.ListPending(context.Background(), 1, 5, now)
	if err != nil {
		t.Fatalf("expected list pending without error, got %v", err)
	}
	if len(claimed) != 1 || claimed[0].Id != id {
		t.Fatalf("expected message %s to remain claimable, got %+v", id, claimed)
	}
}
