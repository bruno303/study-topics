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
		)
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

	pool.Close()
}
