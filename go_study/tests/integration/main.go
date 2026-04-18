package integration

import (
	"context"
	"os"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/infra/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(t *testing.T) (context.Context, *config.Config) {
	t.Helper()

	os.Setenv("CONFIG_FILE", "test.yaml")
	return context.Background(), config.LoadConfig()
}

func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx, cfg := Setup(t)

	pool := database.Connect(cfg)

	createTableSQL := `
		CREATE TABLE IF NOT EXISTS HELLO_DATA (
			ID VARCHAR(100) PRIMARY KEY,
			NAME VARCHAR(300) NOT NULL,
			AGE INT
		);
		CREATE TABLE IF NOT EXISTS OUTBOX_MESSAGES (
			ID UUID PRIMARY KEY,
			TOPIC VARCHAR(255) NOT NULL,
			PAYLOAD JSONB NOT NULL,
			STATUS VARCHAR(32) NOT NULL,
			ATTEMPT_COUNT INT NOT NULL DEFAULT 0,
			AVAILABLE_AT TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			LAST_ERROR TEXT,
			PUBLISHED_AT TIMESTAMPTZ,
			CREATED_AT TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UPDATED_AT TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`
	if _, err := pool.Exec(ctx, createTableSQL); err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	return pool
}

func CleanupTestDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	if _, err := pool.Exec(ctx, "DELETE FROM HELLO_DATA"); err != nil {
		t.Logf("Warning: Failed to clean up test data: %v", err)
	}
	if _, err := pool.Exec(ctx, "DELETE FROM OUTBOX_MESSAGES"); err != nil {
		t.Logf("Warning: Failed to clean up outbox test data: %v", err)
	}

	pool.Close()
}
