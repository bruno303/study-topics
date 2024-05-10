package database

import (
	"context"
	"fmt"
	"main/internal/config"
	"main/internal/infra/utils/shutdown"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(config *config.Config) *pgxpool.Pool {
	ctx := context.Background()
	cfg := config.Database
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DatabaseName)
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	if err = pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to %s:%d: %v\n", cfg.Host, cfg.Port, err)
		os.Exit(1)
	}

	shutdown.CreateListener(func() {
		fmt.Println("Closing database pool")
		pool.Close()
	})

	return pool
}
