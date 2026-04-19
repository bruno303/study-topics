package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/bruno303/study-topics/go-study/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const migrationLockID int64 = 748_597_284_063_218_391

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

var gooseUpContext = goose.UpContext

func RunMigrations(ctx context.Context, cfg *config.Config) error {
	db, err := sql.Open("pgx", connectionString(cfg))
	if err != nil {
		return fmt.Errorf("open migration database: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err = db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("ping migration database: %w", err)
	}

	defer db.Close()

	if err := runMigrations(ctx, db); err != nil {
		return err
	}

	return nil
}

func runMigrations(ctx context.Context, db *sql.DB) (err error) {
	if _, err = db.ExecContext(ctx, "SELECT pg_advisory_lock($1)", migrationLockID); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}

	defer func() {
		if _, unlockErr := db.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", migrationLockID); unlockErr != nil && err == nil {
			err = fmt.Errorf("release migration lock: %w", unlockErr)
		}
	}()

	goose.SetBaseFS(embeddedMigrations)
	if err = goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err = gooseUpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("run goose migrations: %w", err)
	}

	return nil
}
