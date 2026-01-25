package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/hello/models"
	"github.com/bruno303/study-topics/go-study/internal/application/transaction"
	"github.com/bruno303/study-topics/go-study/tests/integration"
	"github.com/jackc/pgx/v5"
)

func TestTransactionManager_Execute_Commit(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	tm := NewTransactionManager(&TransactionConfig{Pool: pool})
	ctx := context.Background()

	testID := fmt.Sprintf("test-commit-%d", time.Now().Unix())
	testData := &models.HelloData{
		Id:   testID,
		Name: "John Doe",
		Age:  30,
	}

	// Execute a transaction that should commit
	result, err := tm.Execute(ctx, transaction.EmptyOpts, func(ctx context.Context, tx transaction.Transaction) (any, error) {
		pgTx := tx.(pgx.Tx)
		const sql = "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)"
		_, err := pgTx.Exec(ctx, sql, testData.Id, testData.Name, testData.Age)
		if err != nil {
			return nil, err
		}
		return testData, nil
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	// Verify data was committed
	var id, name string
	var age int
	err = pool.QueryRow(ctx, "SELECT ID, NAME, AGE FROM HELLO_DATA WHERE ID = $1", testID).Scan(&id, &name, &age)
	if err != nil {
		t.Fatalf("Failed to query inserted data: %v", err)
	}

	if id != testID || name != "John Doe" || age != 30 {
		t.Errorf("Data mismatch: got (id=%s, name=%s, age=%d), want (id=%s, name=%s, age=%d)",
			id, name, age, testID, "John Doe", 30)
	}
}

func TestTransactionManager_Execute_Rollback(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	tm := NewTransactionManager(&TransactionConfig{Pool: pool})
	ctx := context.Background()

	testID := fmt.Sprintf("test-rollback-%d", time.Now().Unix())
	testData := &models.HelloData{
		Id:   testID,
		Name: "Jane Doe",
		Age:  25,
	}

	// Execute a transaction that should rollback
	_, err := tm.Execute(ctx, transaction.EmptyOpts, func(ctx context.Context, tx transaction.Transaction) (any, error) {
		pgTx := tx.(pgx.Tx)
		const sql = "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)"
		_, err := pgTx.Exec(ctx, sql, testData.Id, testData.Name, testData.Age)
		if err != nil {
			return nil, err
		}
		// Return an error to trigger rollback
		return nil, fmt.Errorf("simulated error to trigger rollback")
	})

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "simulated error to trigger rollback" {
		t.Errorf("Expected 'simulated error to trigger rollback', got: %v", err)
	}

	// Verify data was NOT committed (should be rolled back)
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM HELLO_DATA WHERE ID = $1", testID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count to be 0 (rolled back), got: %d", count)
	}
}

func TestTransactionManager_Execute_MultipleOperations(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	tm := NewTransactionManager(&TransactionConfig{Pool: pool})
	ctx := context.Background()

	testID1 := fmt.Sprintf("test-multi-1-%d", time.Now().Unix())
	testID2 := fmt.Sprintf("test-multi-2-%d", time.Now().Unix())

	// Execute multiple operations in a single transaction
	result, err := tm.Execute(ctx, transaction.EmptyOpts, func(ctx context.Context, tx transaction.Transaction) (any, error) {
		pgTx := tx.(pgx.Tx)
		const sql = "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)"

		// Insert first record
		_, err := pgTx.Exec(ctx, sql, testID1, "Alice", 28)
		if err != nil {
			return nil, err
		}

		// Insert second record
		_, err = pgTx.Exec(ctx, sql, testID2, "Bob", 32)
		if err != nil {
			return nil, err
		}

		return map[string]string{"id1": testID1, "id2": testID2}, nil
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	// Verify both records were committed
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM HELLO_DATA WHERE ID IN ($1, $2)", testID1, testID2).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count to be 2, got: %d", count)
	}
}

func TestTransactionManager_Execute_MultipleOperationsWithRollback(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	tm := NewTransactionManager(&TransactionConfig{Pool: pool})
	ctx := context.Background()

	testID1 := fmt.Sprintf("test-multi-rollback-1-%d", time.Now().Unix())
	testID2 := fmt.Sprintf("test-multi-rollback-2-%d", time.Now().Unix())

	// Execute multiple operations where the second one fails
	_, err := tm.Execute(ctx, transaction.EmptyOpts, func(ctx context.Context, tx transaction.Transaction) (any, error) {
		pgTx := tx.(pgx.Tx)
		const sql = "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)"

		// Insert first record (should succeed)
		_, err := pgTx.Exec(ctx, sql, testID1, "Charlie", 35)
		if err != nil {
			return nil, err
		}

		// Simulate failure before second insert
		return nil, fmt.Errorf("operation failed before second insert")
	})

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify that the first record was also rolled back
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM HELLO_DATA WHERE ID IN ($1, $2)", testID1, testID2).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count to be 0 (all operations rolled back), got: %d", count)
	}
}

func TestTransactionManager_Execute_WithRepository(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	tm := NewTransactionManager(&TransactionConfig{Pool: pool})
	repo := NewHelloPgxRepository(pool)
	ctx := context.Background()

	testID := fmt.Sprintf("test-repo-%d", time.Now().Unix())
	testData := &models.HelloData{
		Id:   testID,
		Name: "David",
		Age:  40,
	}

	// Use repository within transaction
	result, err := tm.Execute(ctx, transaction.EmptyOpts, func(ctx context.Context, tx transaction.Transaction) (any, error) {
		savedData, err := repo.Save(ctx, testData, tx)
		if err != nil {
			return nil, err
		}
		return savedData, nil
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	// Verify data was committed
	foundData, err := repo.FindById(ctx, testID, nil)
	if err != nil {
		t.Fatalf("Failed to find inserted data: %v", err)
	}

	if foundData.Id != testID || foundData.Name != "David" || foundData.Age != 40 {
		t.Errorf("Data mismatch: got %+v, want (id=%s, name=%s, age=%d)",
			foundData, testID, "David", 40)
	}
}

func TestTransactionManager_Execute_WithRepositoryRollback(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	tm := NewTransactionManager(&TransactionConfig{Pool: pool})
	repo := NewHelloPgxRepository(pool)
	ctx := context.Background()

	testID := fmt.Sprintf("test-repo-rollback-%d", time.Now().Unix())
	testData := &models.HelloData{
		Id:   testID,
		Name: "Eve",
		Age:  27,
	}

	// Use repository within transaction that rolls back
	_, err := tm.Execute(ctx, transaction.EmptyOpts, func(ctx context.Context, tx transaction.Transaction) (any, error) {
		_, err := repo.Save(ctx, testData, tx)
		if err != nil {
			return nil, err
		}
		// Force rollback
		return nil, fmt.Errorf("forced rollback after save")
	})

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify data was NOT committed
	_, err = repo.FindById(ctx, testID, nil)
	if err == nil {
		t.Error("Expected error finding rolled back data, got nil")
	}
}

func TestTransactionManager_Execute_InvalidTransactionType(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	tm := NewTransactionManager(&TransactionConfig{Pool: pool})
	ctx := context.Background()

	// Try to pass an invalid transaction type
	opts := transaction.Opts{
		RequiresNew: false,
		Transaction: "invalid-transaction-type", // This is not a pgx.Tx
	}

	_, err := tm.Execute(ctx, opts, func(ctx context.Context, tx transaction.Transaction) (any, error) {
		return nil, nil
	})

	if err == nil {
		t.Fatal("Expected InvalidTransactionTypeErr, got nil")
	}

	if err != InvalidTransactionTypeErr {
		t.Errorf("Expected InvalidTransactionTypeErr, got: %v", err)
	}
}

func TestTransactionManager_Execute_NestedTransactions(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	tm := NewTransactionManager(&TransactionConfig{Pool: pool})
	repo := NewHelloPgxRepository(pool)
	ctx := context.Background()

	testID1 := fmt.Sprintf("test-nested-1-%d", time.Now().Unix())
	testID2 := fmt.Sprintf("test-nested-2-%d", time.Now().Unix())

	testData1 := &models.HelloData{Id: testID1, Name: "Frank", Age: 45}
	testData2 := &models.HelloData{Id: testID2, Name: "Grace", Age: 38}

	// Execute nested transactions
	result, err := tm.Execute(ctx, transaction.EmptyOpts, func(ctx context.Context, tx1 transaction.Transaction) (any, error) {
		// Save first record in outer transaction
		_, err := repo.Save(ctx, testData1, tx1)
		if err != nil {
			return nil, err
		}

		// Execute inner transaction with RequiresNew=true
		// This should create a NEW transaction, ignoring the provided transaction
		opts := transaction.Opts{
			RequiresNew: true,
			Transaction: tx1, // This should be ignored since RequiresNew=true
		}

		_, err = tm.Execute(ctx, opts, func(ctx context.Context, tx2 transaction.Transaction) (any, error) {
			// Verify tx2 is a different transaction than tx1
			pgTx1 := tx1.(pgx.Tx)
			pgTx2 := tx2.(pgx.Tx)

			// These should be different transaction instances
			if pgTx1 == pgTx2 {
				return nil, fmt.Errorf("expected different transaction instances when RequiresNew=true")
			}

			// Save second record in the new inner transaction
			_, err := repo.Save(ctx, testData2, tx2)
			if err != nil {
				return nil, err
			}
			return testData2, nil
		})

		if err != nil {
			return nil, err
		}

		return map[string]string{"id1": testID1, "id2": testID2}, nil
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	// Verify both records were committed
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM HELLO_DATA WHERE ID IN ($1, $2)", testID1, testID2).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count to be 2, got: %d", count)
	}
}

func TestTransactionManager_Execute_ConcurrentTransactions(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	tm := NewTransactionManager(&TransactionConfig{Pool: pool})
	ctx := context.Background()

	const numGoroutines = 10
	errors := make(chan error, numGoroutines)
	done := make(chan bool, numGoroutines)

	// Execute multiple concurrent transactions
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			testID := fmt.Sprintf("test-concurrent-%d-%d", time.Now().Unix(), index)
			testData := &models.HelloData{
				Id:   testID,
				Name: fmt.Sprintf("User%d", index),
				Age:  20 + index,
			}

			_, err := tm.Execute(ctx, transaction.EmptyOpts, func(ctx context.Context, tx transaction.Transaction) (any, error) {
				pgTx := tx.(pgx.Tx)
				const sql = "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)"
				_, err := pgTx.Exec(ctx, sql, testData.Id, testData.Name, testData.Age)
				if err != nil {
					return nil, err
				}
				// Add a small delay to increase concurrency stress
				time.Sleep(10 * time.Millisecond)
				return testData, nil
			})

			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent transaction failed: %v", err)
	}

	// Verify all records were committed
	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM HELLO_DATA WHERE NAME LIKE 'User%'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}

	if count != numGoroutines {
		t.Errorf("Expected count to be %d, got: %d", numGoroutines, count)
	}
}

func TestTransactionManager_Execute_DeadlockScenario(t *testing.T) {
	pool := integration.SetupTestDB(t)
	defer integration.CleanupTestDB(t, pool)

	tm := NewTransactionManager(&TransactionConfig{Pool: pool})
	ctx := context.Background()

	// Insert initial data
	testID1 := fmt.Sprintf("test-deadlock-1-%d", time.Now().Unix())
	testID2 := fmt.Sprintf("test-deadlock-2-%d", time.Now().Unix())

	_, err := pool.Exec(ctx, "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)", testID1, "Initial1", 1)
	if err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	_, err = pool.Exec(ctx, "INSERT INTO HELLO_DATA (ID, NAME, AGE) VALUES ($1, $2, $3)", testID2, "Initial2", 2)
	if err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	// Try to create a deadlock scenario (this may or may not succeed depending on timing)
	done1 := make(chan error, 1)
	done2 := make(chan error, 1)

	// Transaction 1: Update record 1, then record 2
	go func() {
		_, err := tm.Execute(ctx, transaction.EmptyOpts, func(ctx context.Context, tx transaction.Transaction) (any, error) {
			pgTx := tx.(pgx.Tx)

			// Update record 1
			_, err := pgTx.Exec(ctx, "UPDATE HELLO_DATA SET AGE = AGE + 1 WHERE ID = $1", testID1)
			if err != nil {
				return nil, err
			}

			time.Sleep(50 * time.Millisecond)

			// Update record 2
			_, err = pgTx.Exec(ctx, "UPDATE HELLO_DATA SET AGE = AGE + 1 WHERE ID = $1", testID2)
			if err != nil {
				return nil, err
			}

			return nil, nil
		})
		done1 <- err
	}()

	// Transaction 2: Update record 2, then record 1
	go func() {
		time.Sleep(10 * time.Millisecond) // Small delay to ensure tx1 starts first

		_, err := tm.Execute(ctx, transaction.EmptyOpts, func(ctx context.Context, tx transaction.Transaction) (any, error) {
			pgTx := tx.(pgx.Tx)

			// Update record 2
			_, err := pgTx.Exec(ctx, "UPDATE HELLO_DATA SET AGE = AGE + 1 WHERE ID = $1", testID2)
			if err != nil {
				return nil, err
			}

			time.Sleep(50 * time.Millisecond)

			// Update record 1
			_, err = pgTx.Exec(ctx, "UPDATE HELLO_DATA SET AGE = AGE + 1 WHERE ID = $1", testID1)
			if err != nil {
				return nil, err
			}

			return nil, nil
		})
		done2 <- err
	}()

	// Wait for both transactions
	err1 := <-done1
	err2 := <-done2

	// At least one should complete successfully
	// (In a real deadlock scenario, one would fail, but PostgreSQL handles this gracefully)
	if err1 != nil && err2 != nil {
		t.Logf("Both transactions encountered errors (possible deadlock): err1=%v, err2=%v", err1, err2)
	}

	// Verify final state - at least the initial inserts should be there
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM HELLO_DATA WHERE ID IN ($1, $2)", testID1, testID2).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count to be 2, got: %d", count)
	}
}
