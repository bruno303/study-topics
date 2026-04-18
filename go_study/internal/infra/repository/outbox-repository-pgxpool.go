package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxPgxRepository struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func NewOutboxPgxRepository(pool *pgxpool.Pool, tx pgx.Tx) applicationRepository.OutboxRepository {
	return &OutboxPgxRepository{pool: pool, tx: tx}
}

func (r *OutboxPgxRepository) Enqueue(ctx context.Context, msg applicationRepository.OutboxMessage) error {
	msg = applyOutboxDefaults(msg)

	sqlStmt := `
		INSERT INTO OUTBOX_MESSAGES (
			ID, TOPIC, PAYLOAD, STATUS, ATTEMPT_COUNT, AVAILABLE_AT, LAST_ERROR, PUBLISHED_AT, CREATED_AT, UPDATED_AT
		) VALUES ($1, $2, $3::jsonb, $4, $5, $6, $7, $8, $9, $10)
	`
	params := []any{
		msg.ID,
		msg.Topic,
		msg.Payload,
		msg.Status,
		msg.Attempts,
		msg.AvailableAt,
		msg.LastError,
		nullTime(msg.PublishedAt),
		msg.CreatedAt,
		msg.UpdatedAt,
	}

	pgTx := r.getTransactionOrNil()
	if pgTx == nil {
		_, err := r.pool.Exec(ctx, sqlStmt, params...)
		return err
	}
	_, err := pgTx.Exec(ctx, sqlStmt, params...)
	return err
}

func (r *OutboxPgxRepository) ClaimNext(ctx context.Context, maxAttempts int) (*applicationRepository.OutboxMessage, error) {
	const sqlStmt = `
		SELECT ID, TOPIC, PAYLOAD, STATUS, ATTEMPT_COUNT, AVAILABLE_AT, LAST_ERROR, PUBLISHED_AT, CREATED_AT, UPDATED_AT
		FROM OUTBOX_MESSAGES
		WHERE STATUS = $1
			AND AVAILABLE_AT <= NOW()
			AND ATTEMPT_COUNT < $2
		ORDER BY CREATED_AT ASC, ID ASC
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`

	pgTx := r.getTransactionOrNil()
	if pgTx == nil {
		return nil, errors.New("outbox claim requires active transaction")
	}

	row := pgTx.QueryRow(ctx, sqlStmt, applicationRepository.OutboxStatusPending, maxAttempts)
	msg, err := scanOutboxMessage(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("claim next outbox message: %w", err)
	}

	return msg, nil
}

func (r *OutboxPgxRepository) Update(ctx context.Context, msg applicationRepository.OutboxMessage) error {
	msg = applyOutboxDefaults(msg)

	const sqlStmt = `
		UPDATE OUTBOX_MESSAGES
		SET TOPIC = $2,
			PAYLOAD = $3::jsonb,
			STATUS = $4,
			ATTEMPT_COUNT = $5,
			AVAILABLE_AT = $6,
			LAST_ERROR = $7,
			PUBLISHED_AT = $8,
			UPDATED_AT = $9
		WHERE ID = $1
	`

	params := []any{
		msg.ID,
		msg.Topic,
		msg.Payload,
		msg.Status,
		msg.Attempts,
		msg.AvailableAt,
		msg.LastError,
		nullTime(msg.PublishedAt),
		msg.UpdatedAt,
	}

	pgTx := r.getTransactionOrNil()
	if pgTx == nil {
		_, err := r.pool.Exec(ctx, sqlStmt, params...)
		return err
	}
	_, err := pgTx.Exec(ctx, sqlStmt, params...)
	return err
}

func (r *OutboxPgxRepository) getTransactionOrNil() pgx.Tx {
	return r.tx
}

func applyOutboxDefaults(msg applicationRepository.OutboxMessage) applicationRepository.OutboxMessage {
	now := time.Now()
	if msg.Status == "" {
		msg.Status = applicationRepository.OutboxStatusPending
	}
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	if msg.UpdatedAt.IsZero() {
		msg.UpdatedAt = now
	}
	if msg.AvailableAt.IsZero() {
		msg.AvailableAt = now
	}
	return msg
}

func nullTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func scanOutboxMessage(row pgx.Row) (*applicationRepository.OutboxMessage, error) {
	var msg applicationRepository.OutboxMessage
	var publishedAt sql.NullTime
	var lastError sql.NullString

	err := row.Scan(
		&msg.ID,
		&msg.Topic,
		&msg.Payload,
		&msg.Status,
		&msg.Attempts,
		&msg.AvailableAt,
		&lastError,
		&publishedAt,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if publishedAt.Valid {
		msg.PublishedAt = publishedAt.Time
	}
	if lastError.Valid {
		msg.LastError = lastError.String
	}
	return &msg, nil
}
