package integration

import (
	"context"
	"testing"

	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/infra/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Setup(t *testing.T) (context.Context, *config.Config) {
	t.Helper()

	t.Setenv("CONFIG_FILE", "test.yaml")
	return context.Background(), config.LoadConfig()
}

func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx, cfg := Setup(t)

	if err := database.RunMigrations(ctx, cfg); err != nil {
		t.Fatalf("Failed to run test migrations: %v", err)
	}

	pool := database.Connect(cfg)

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
