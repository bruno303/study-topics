package database

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/pressly/goose/v3"
)

func TestRunMigrationsAcquiresLockBeforeRunningGoose(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	lockQuery := regexp.QuoteMeta("SELECT pg_advisory_lock($1)")
	unlockQuery := regexp.QuoteMeta("SELECT pg_advisory_unlock($1)")
	mock.ExpectExec(lockQuery).WithArgs(migrationLockID).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(unlockQuery).WithArgs(migrationLockID).WillReturnResult(sqlmock.NewResult(0, 0))

	originalGooseUpContext := gooseUpContext
	gooseUpContext = func(ctx context.Context, db *sql.DB, dir string, opts ...goose.OptionsFunc) error {
		if dir != "migrations" {
			t.Fatalf("unexpected migration directory: %s", dir)
		}
		return nil
	}
	t.Cleanup(func() {
		gooseUpContext = originalGooseUpContext
	})

	if err := runMigrations(context.Background(), db); err != nil {
		t.Fatalf("runMigrations() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRunMigrationsReleasesLockWhenGooseFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	lockQuery := regexp.QuoteMeta("SELECT pg_advisory_lock($1)")
	unlockQuery := regexp.QuoteMeta("SELECT pg_advisory_unlock($1)")
	mock.ExpectExec(lockQuery).WithArgs(migrationLockID).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(unlockQuery).WithArgs(migrationLockID).WillReturnResult(sqlmock.NewResult(0, 0))

	originalGooseUpContext := gooseUpContext
	gooseUpContext = func(ctx context.Context, db *sql.DB, dir string, opts ...goose.OptionsFunc) error {
		return errors.New("boom")
	}
	t.Cleanup(func() {
		gooseUpContext = originalGooseUpContext
	})

	err = runMigrations(context.Background(), db)
	if err == nil {
		t.Fatal("runMigrations() error = nil, want failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
