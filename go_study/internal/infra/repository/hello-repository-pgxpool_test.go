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

func TestHelloRepository_ListAll_WhenQueryFails_ReturnsWrappedError(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	repo := NewHelloPgxRepository(pool)
	ctx := context.Background()

	if _, err := pool.Exec(ctx, "DROP TABLE HELLO_DATA"); err != nil {
		t.Fatalf("failed to drop HELLO_DATA table: %v", err)
	}

	result, err := repo.ListAll(ctx)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
	if !strings.Contains(err.Error(), "list all hello data") {
		t.Fatalf("expected wrapped query error, got: %v", err)
	}
}

func TestHelloRepository_ListAll_WhenScanFails_ReturnsWrappedError(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	repo := NewHelloPgxRepository(pool)
	ctx := context.Background()

	testID := fmt.Sprintf("test-list-all-scan-error-%d", time.Now().UnixNano())
	_, err := pool.Exec(ctx, "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)", testID, "NullAge", nil)
	if err != nil {
		t.Fatalf("failed to insert row with nullable age: %v", err)
	}

	result, err := repo.ListAll(ctx)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
	if !strings.Contains(err.Error(), "scan hello data row") {
		t.Fatalf("expected wrapped scan error, got: %v", err)
	}
}

func TestHelloRepository_ListAll_WhenRowsIterationFails_ReturnsWrappedError(t *testing.T) {
	pool := integration.SetupTestDB(t)
	ctx := context.Background()

	defer func() {
		if _, err := pool.Exec(ctx, "DROP VIEW IF EXISTS HELLO_DATA"); err != nil {
			t.Errorf("failed to drop HELLO_DATA view: %v", err)
		}
		if _, err := pool.Exec(ctx, "DROP FUNCTION IF EXISTS fail_on_bad_row(TEXT, INT)"); err != nil {
			t.Errorf("failed to drop fail_on_bad_row function: %v", err)
		}
		if _, err := pool.Exec(ctx, "ALTER TABLE IF EXISTS HELLO_DATA_BASE RENAME TO HELLO_DATA"); err != nil {
			t.Errorf("failed to restore HELLO_DATA table name: %v", err)
		}
		integration.CleanupTestDB(t, pool)
	}()

	repo := NewHelloPgxRepository(pool)

	if _, err := pool.Exec(ctx, "ALTER TABLE HELLO_DATA RENAME TO HELLO_DATA_BASE"); err != nil {
		t.Fatalf("failed to rename HELLO_DATA table: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		CREATE OR REPLACE FUNCTION fail_on_bad_row(input_id TEXT, input_age INT)
		RETURNS INT
		LANGUAGE plpgsql
		AS $$
		BEGIN
			IF input_id = 'b-bad' THEN
				RAISE EXCEPTION 'forced row iteration failure';
			END IF;
			RETURN input_age;
		END;
		$$
	`); err != nil {
		t.Fatalf("failed to create fail_on_bad_row function: %v", err)
	}

	if _, err := pool.Exec(ctx, `
		CREATE VIEW HELLO_DATA AS
		SELECT ID, NAME, fail_on_bad_row(ID, AGE) AS AGE
		FROM HELLO_DATA_BASE
		ORDER BY ID
	`); err != nil {
		t.Fatalf("failed to create HELLO_DATA view: %v", err)
	}

	if _, err := pool.Exec(ctx, "INSERT INTO HELLO_DATA_BASE (ID, NAME, AGE) VALUES ($1, $2, $3)", "a-good", "Good", 10); err != nil {
		t.Fatalf("failed to insert first row: %v", err)
	}
	if _, err := pool.Exec(ctx, "INSERT INTO HELLO_DATA_BASE (ID, NAME, AGE) VALUES ($1, $2, $3)", "b-bad", "Bad", 20); err != nil {
		t.Fatalf("failed to insert second row: %v", err)
	}

	result, err := repo.ListAll(ctx)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
	if !strings.Contains(err.Error(), "iterate hello data rows") {
		t.Fatalf("expected wrapped rows iteration error, got: %v", err)
	}
}
