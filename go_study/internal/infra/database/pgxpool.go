package database

import (
	"context"
	"fmt"
	"main/internal/config"
	"os"
	"os/signal"
	"syscall"

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

	go func() {
		exitChan := make(chan os.Signal, 1)
		signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)
		<-exitChan
		fmt.Println("Closing database pool")
		// TODO: this close is waiting forever
		// pool.Close()
	}()

	return pool
}
