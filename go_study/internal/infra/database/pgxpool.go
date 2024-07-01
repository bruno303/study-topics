package database

import (
	"context"
	"fmt"
	"os"

	"github.com/bruno303/study-topics/go-study/internal/config"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/log"
	"github.com/bruno303/study-topics/go-study/internal/infra/utils/shutdown"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(config *config.Config) *pgxpool.Pool {
	ctx := context.Background()
	cfg := config.Database
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DatabaseName)
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		log.Log().Error(ctx, "Unable to create connection pool", err)
		os.Exit(1)
	}
	if err = pool.Ping(ctx); err != nil {
		log.Log().Error(ctx, fmt.Sprintf("Unable to connect to %s:%d", cfg.Host, cfg.Port), err)
		os.Exit(1)
	}

	shutdown.CreateListener(func() {
		log.Log().Info(ctx, "Closing database pool")
		pool.Close()
	})

	return pool
}
