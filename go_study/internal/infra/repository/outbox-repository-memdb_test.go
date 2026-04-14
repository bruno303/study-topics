package repository

import (
	"context"
	"testing"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/model"
	"github.com/bruno303/study-topics/go-study/internal/infra/database"
)

func TestOutboxMemDbRepository_ListPending_ClaimsByCreatedAtOrder(t *testing.T) {
	db := database.NewMemDbRepository[memDbOutboxRecord]()
	repo := NewOutboxMemDbRepository(db)
	now := time.Now().UTC()

	messages := []model.OutboxMessage{
		{Id: "m3", Topic: "topic", Status: model.OutboxStatusPending, Attempt: 0, NextAttempt: now, CreatedAt: now.Add(3 * time.Second), UpdatedAt: now},
		{Id: "m1", Topic: "topic", Status: model.OutboxStatusPending, Attempt: 1, NextAttempt: now, CreatedAt: now.Add(1 * time.Second), UpdatedAt: now},
		{Id: "m2", Topic: "topic", Status: model.OutboxStatusError, Attempt: 2, NextAttempt: now, CreatedAt: now.Add(2 * time.Second), UpdatedAt: now},
	}

	for _, message := range messages {
		msg := message
		if _, err := repo.Enqueue(context.Background(), &msg); err != nil {
			t.Fatalf("failed to enqueue message %s: %v", message.Id, err)
		}
	}

	claimed, err := repo.ListPending(context.Background(), 2, 5, now)
	if err != nil {
		t.Fatalf("expected claim without error, got %v", err)
	}
	if len(claimed) != 2 {
		t.Fatalf("expected two claimed messages, got %d", len(claimed))
	}
	if claimed[0].Id != "m1" || claimed[1].Id != "m2" {
		t.Fatalf("expected claim order [m1, m2], got [%s, %s]", claimed[0].Id, claimed[1].Id)
	}
	if claimed[0].Status != model.OutboxStatusProcessing || claimed[1].Status != model.OutboxStatusProcessing {
		t.Fatalf("expected claimed messages to be processing, got [%s, %s]", claimed[0].Status, claimed[1].Status)
	}
	if claimed[0].Attempt != 2 || claimed[1].Attempt != 3 {
		t.Fatalf("expected attempts to be incremented to [2,3], got [%d,%d]", claimed[0].Attempt, claimed[1].Attempt)
	}

	secondClaim, err := repo.ListPending(context.Background(), 10, 5, now)
	if err != nil {
		t.Fatalf("expected second claim without error, got %v", err)
	}
	if len(secondClaim) != 1 || secondClaim[0].Id != "m3" {
		t.Fatalf("expected only remaining claim m3, got %+v", secondClaim)
	}
}

func TestOutboxMemDbRepository_MarkAsError_WhenNextAttemptProvided_SetsRetryableStatus(t *testing.T) {
	db := database.NewMemDbRepository[memDbOutboxRecord]()
	repo := NewOutboxMemDbRepository(db)
	now := time.Now().UTC()

	_, err := repo.Enqueue(context.Background(), &model.OutboxMessage{
		Id:          "retry-id",
		Topic:       "topic",
		Status:      model.OutboxStatusProcessing,
		Attempt:     1,
		NextAttempt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("failed to seed outbox message: %v", err)
	}

	nextAttempt := now.Add(1 * time.Minute)
	if err := repo.MarkAsError(context.Background(), "retry-id", "temporary error", nextAttempt); err != nil {
		t.Fatalf("expected mark as error without error, got %v", err)
	}

	claimed, err := repo.ListPending(context.Background(), 10, 5, nextAttempt)
	if err != nil {
		t.Fatalf("expected list pending without error, got %v", err)
	}
	if len(claimed) != 1 {
		t.Fatalf("expected one claimable message, got %d", len(claimed))
	}
	if claimed[0].Id != "retry-id" {
		t.Fatalf("expected claimed id retry-id, got %s", claimed[0].Id)
	}
}

func TestOutboxMemDbRepository_MarkAsError_WhenNextAttemptIsZero_MarksAsFailed(t *testing.T) {
	db := database.NewMemDbRepository[memDbOutboxRecord]()
	repo := NewOutboxMemDbRepository(db)
	now := time.Now().UTC()

	_, err := repo.Enqueue(context.Background(), &model.OutboxMessage{
		Id:        "failed-id",
		Topic:     "topic",
		Status:    model.OutboxStatusProcessing,
		Attempt:   3,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("failed to seed outbox message: %v", err)
	}

	if err := repo.MarkAsError(context.Background(), "failed-id", "terminal error", time.Time{}); err != nil {
		t.Fatalf("expected mark as failed without error, got %v", err)
	}

	claimed, err := repo.ListPending(context.Background(), 10, 5, now.Add(time.Hour))
	if err != nil {
		t.Fatalf("expected list pending without error, got %v", err)
	}
	if len(claimed) != 0 {
		t.Fatalf("expected failed message not to be claimable, got %d messages", len(claimed))
	}
}

func TestOutboxMemDbRepository_MarkAsError_WhenAlreadyPublished_DoesNotRevertToError(t *testing.T) {
	db := database.NewMemDbRepository[memDbOutboxRecord]()
	repo := NewOutboxMemDbRepository(db)
	now := time.Now().UTC()
	publishedAt := now.Add(30 * time.Second)

	_, err := repo.Enqueue(context.Background(), &model.OutboxMessage{
		Id:          "published-id",
		Topic:       "topic",
		Status:      model.OutboxStatusProcessing,
		Attempt:     2,
		NextAttempt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("failed to seed outbox message: %v", err)
	}

	if err := repo.MarkAsPublished(context.Background(), "published-id", publishedAt); err != nil {
		t.Fatalf("expected mark as published without error, got %v", err)
	}

	if err := repo.MarkAsError(context.Background(), "published-id", "late error", now.Add(time.Minute)); err != nil {
		t.Fatalf("expected late mark as error to be no-op, got %v", err)
	}

	stored, err := db.FindById(context.Background(), "published-id")
	if err != nil {
		t.Fatalf("expected stored row to exist, got %v", err)
	}
	if stored.Message.Status != model.OutboxStatusPublished {
		t.Fatalf("expected status to remain published, got %s", stored.Message.Status)
	}
	if stored.Message.PublishedAt == nil || !stored.Message.PublishedAt.Equal(publishedAt) {
		t.Fatalf("expected published_at to remain %v, got %v", publishedAt, stored.Message.PublishedAt)
	}
	if stored.Message.LastError != nil {
		t.Fatalf("expected last_error to remain nil, got %v", *stored.Message.LastError)
	}
}

func TestOutboxMemDbRepository_MarkAsPublished_WhenStatusIsNotProcessing_IsNoOp(t *testing.T) {
	db := database.NewMemDbRepository[memDbOutboxRecord]()
	repo := NewOutboxMemDbRepository(db)
	now := time.Now().UTC()

	_, err := repo.Enqueue(context.Background(), &model.OutboxMessage{
		Id:          "pending-id",
		Topic:       "topic",
		Status:      model.OutboxStatusPending,
		NextAttempt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("failed to seed pending outbox message: %v", err)
	}

	if err := repo.MarkAsPublished(context.Background(), "pending-id", now.Add(time.Minute)); err != nil {
		t.Fatalf("expected mark as published attempt without error, got %v", err)
	}

	claimed, err := repo.ListPending(context.Background(), 1, 5, now)
	if err != nil {
		t.Fatalf("expected list pending without error, got %v", err)
	}
	if len(claimed) != 1 || claimed[0].Id != "pending-id" {
		t.Fatalf("expected pending-id to remain claimable, got %+v", claimed)
	}
}

func TestOutboxMemDbRepository_MarkAsError_WhenStatusIsNotProcessing_IsNoOp(t *testing.T) {
	db := database.NewMemDbRepository[memDbOutboxRecord]()
	repo := NewOutboxMemDbRepository(db)
	now := time.Now().UTC()

	_, err := repo.Enqueue(context.Background(), &model.OutboxMessage{
		Id:          "pending-error-id",
		Topic:       "topic",
		Status:      model.OutboxStatusPending,
		NextAttempt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("failed to seed pending outbox message: %v", err)
	}

	if err := repo.MarkAsError(context.Background(), "pending-error-id", "should be ignored", now.Add(time.Minute)); err != nil {
		t.Fatalf("expected mark as error attempt without error, got %v", err)
	}

	claimed, err := repo.ListPending(context.Background(), 1, 5, now)
	if err != nil {
		t.Fatalf("expected list pending without error, got %v", err)
	}
	if len(claimed) != 1 || claimed[0].Id != "pending-error-id" {
		t.Fatalf("expected pending-error-id to remain claimable, got %+v", claimed)
	}
}

func TestOutboxMemDbRepository_Enqueue_WhenMinimalFieldsProvided_DefaultsAndClaimsAsPending(t *testing.T) {
	db := database.NewMemDbRepository[memDbOutboxRecord]()
	repo := NewOutboxMemDbRepository(db)

	message := model.OutboxMessage{
		Id:    "minimal-memdb-id",
		Topic: "topic",
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
	if claimed[0].Id != "minimal-memdb-id" {
		t.Fatalf("expected claimed id minimal-memdb-id, got %s", claimed[0].Id)
	}
	if claimed[0].Status != model.OutboxStatusProcessing {
		t.Fatalf("expected claimed status processing, got %s", claimed[0].Status)
	}
	if claimed[0].Attempt != 1 {
		t.Fatalf("expected claimed attempt incremented to 1, got %d", claimed[0].Attempt)
	}
}

func TestOutboxMemDbRepository_ListPending_WhenProcessingIsStale_ReclaimsOnlyStaleMessages(t *testing.T) {
	db := database.NewMemDbRepository[memDbOutboxRecord]()
	repo := NewOutboxMemDbRepository(db)
	now := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)
	maxAttempts := 3

	stale := model.OutboxMessage{
		Id:          "stale-processing",
		Topic:       "topic",
		Status:      model.OutboxStatusProcessing,
		Attempt:     2,
		NextAttempt: now.Add(10 * time.Minute),
		CreatedAt:   now.Add(-2 * time.Minute),
		UpdatedAt:   now.Add(-outboxProcessingRecoveryWindow - time.Second),
	}
	exhausted := model.OutboxMessage{
		Id:          "exhausted-processing",
		Topic:       "topic",
		Status:      model.OutboxStatusProcessing,
		Attempt:     maxAttempts,
		NextAttempt: now.Add(10 * time.Minute),
		CreatedAt:   now.Add(-90 * time.Second),
		UpdatedAt:   now.Add(-outboxProcessingRecoveryWindow - time.Second),
	}
	fresh := model.OutboxMessage{
		Id:          "fresh-processing",
		Topic:       "topic",
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
		t.Fatalf("expected only one stale message to be reclaimed, got %d", len(claimed))
	}
	if claimed[0].Id != "stale-processing" {
		t.Fatalf("expected stale-processing to be reclaimed, got %s", claimed[0].Id)
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

func TestOutboxMemDbRepository_ListPending_WhenMaxAttemptsIsZeroOrNegative_ReturnsNoMessages(t *testing.T) {
	testCases := []struct {
		name        string
		maxAttempts int
	}{
		{name: "zero", maxAttempts: 0},
		{name: "negative", maxAttempts: -1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			db := database.NewMemDbRepository[memDbOutboxRecord]()
			repo := NewOutboxMemDbRepository(db)
			now := time.Now().UTC()

			_, err := repo.Enqueue(context.Background(), &model.OutboxMessage{
				Id:          "max-attempts-check-" + testCase.name,
				Topic:       "topic",
				Status:      model.OutboxStatusPending,
				Attempt:     0,
				NextAttempt: now,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
			if err != nil {
				t.Fatalf("failed to seed outbox message: %v", err)
			}

			claimed, listErr := repo.ListPending(context.Background(), 10, testCase.maxAttempts, now)
			if listErr != nil {
				t.Fatalf("expected list pending without error, got %v", listErr)
			}
			if len(claimed) != 0 {
				t.Fatalf("expected no claimed messages for max-attempts %d, got %d", testCase.maxAttempts, len(claimed))
			}
		})
	}
}
