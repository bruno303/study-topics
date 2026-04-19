package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/tests/integration"
)

func TestHelloRepository_ListAll_UsesPoolWhenNoUnitOfWorkTransaction(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	repo := NewHelloPgxRepository(pool)
	ctx := context.Background()

	testID := fmt.Sprintf("test-list-all-pool-%d", time.Now().UnixNano())
	_, err := pool.Exec(ctx, "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)", testID, "NoTx", 33)
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	result, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected one result, got %d", len(result))
	}

	if result[0].Id != testID || result[0].Name != "NoTx" || result[0].Age != 33 {
		t.Fatalf("unexpected row returned: %+v", result[0])
	}
}

func TestHelloRepository_ListAll_WhenTransactionProvided_UsesTransaction(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !strings.Contains(err.Error(), "closed") {
			t.Fatalf("failed to rollback transaction: %v", err)
		}
	}()

	txRepo := newHelloPgxRepository(pool, tx)
	poolRepo := NewHelloPgxRepository(pool)
	entity := &models.HelloData{Id: fmt.Sprintf("test-list-all-tx-%d", time.Now().UnixNano()), Name: "InTx", Age: 27}

	saved, err := txRepo.Save(ctx, entity)
	if err != nil {
		t.Fatalf("failed to save entity inside transaction: %v", err)
	}
	if saved == nil || *saved != *entity {
		t.Fatalf("expected saved entity %+v, got %+v", *entity, saved)
	}

	visibleInTx, err := txRepo.ListAll(ctx)
	if err != nil {
		t.Fatalf("failed to list entities inside transaction: %v", err)
	}
	if len(visibleInTx) != 1 {
		t.Fatalf("expected one transactional row, got %d", len(visibleInTx))
	}
	if visibleInTx[0] != *entity {
		t.Fatalf("expected transactional row %+v, got %+v", *entity, visibleInTx[0])
	}

	visibleInPool, err := poolRepo.ListAll(ctx)
	if err != nil {
		t.Fatalf("failed to list entities outside transaction: %v", err)
	}
	if len(visibleInPool) != 0 {
		t.Fatalf("expected pool-backed repository not to see uncommitted data, got %d rows", len(visibleInPool))
	}
}
