package infra

import (
	"context"
	"fmt"
	"main/internal/config"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	Pgxpool *pgxpool.Pool
}

func NewContainer(config *config.Config) Container {
	cfg := config.Database
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DatabaseName)
	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	return Container{
		Pgxpool: pool,
	}
}
