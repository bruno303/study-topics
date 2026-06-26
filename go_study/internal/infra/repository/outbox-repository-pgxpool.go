package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/outbox"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxRepository struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

const outboxTraceName = "OutboxRepository"

func NewOutboxPgxRepository(pool *pgxpool.Pool) OutboxRepository {
	return newOutboxPgxRepository(pool, nil)
}

func newOutboxPgxRepository(pool *pgxpool.Pool, tx pgx.Tx) OutboxRepository {
	return OutboxRepository{
		pool: pool,
		tx:   tx,
	}
}

func (r OutboxRepository) Insert(ctx context.Context, msg *outbox.OutboxMessage) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig(outboxTraceName, "Insert"))
	defer end()

	trace.InjectAttributes(ctx, attr.New("msgID", msg.ID))

	const sql = "INSERT INTO outbox (id, payload, topic, headers, status, attempt_counter, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"

	pgTx := r.getTransactionOrNil()
	var err error
	if pgTx == nil {
		_, err = r.pool.Exec(ctx, sql, msg.ID, msg.Payload, msg.Topic, msg.Headers, msg.Status, msg.AttemptCounter, msg.CreatedAt, msg.UpdatedAt)
	} else {
		_, err = pgTx.Exec(ctx, sql, msg.ID, msg.Payload, msg.Topic, msg.Headers, msg.Status, msg.AttemptCounter, msg.CreatedAt, msg.UpdatedAt)
	}

	if err != nil {
		trace.InjectError(ctx, err)
		return fmt.Errorf("insert outbox message: %w", err)
	}
	return nil
}

func (r OutboxRepository) FetchPendingBatch(ctx context.Context, limit int) ([]*outbox.OutboxMessage, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(outboxTraceName, "FetchPendingBatch"))
	defer end()

	trace.InjectAttributes(ctx, attr.New("limit", fmt.Sprintf("%d", limit)))

	const sql = "SELECT id, payload, topic, headers, status, attempt_counter, last_error, created_at, updated_at, sent_at, deleted_at FROM outbox WHERE status = 'pending' AND deleted_at IS NULL ORDER BY created_at LIMIT $1 FOR UPDATE SKIP LOCKED"

	pgTx := r.getTransactionOrNil()
	var rows pgx.Rows
	var err error
	if pgTx == nil {
		rows, err = r.pool.Query(ctx, sql, limit)
	} else {
		rows, err = pgTx.Query(ctx, sql, limit)
	}

	if err != nil {
		trace.InjectError(ctx, err)
		return nil, fmt.Errorf("fetch pending outbox batch: %w", err)
	}
	defer rows.Close()

	result := make([]*outbox.OutboxMessage, 0)
	for rows.Next() {
		msg, err := r.mapRowsNextValue(&rows)
		if err != nil {
			trace.InjectError(ctx, err)
			return nil, fmt.Errorf("scan outbox row: %w", err)
		}
		result = append(result, msg)
	}

	if err = rows.Err(); err != nil {
		trace.InjectError(ctx, err)
		return nil, fmt.Errorf("iterate outbox rows: %w", err)
	}

	return result, nil
}

func (r OutboxRepository) MarkSent(ctx context.Context, msgID string) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig(outboxTraceName, "MarkSent"))
	defer end()

	trace.InjectAttributes(ctx, attr.New("msgID", msgID))

	const sql = "UPDATE outbox SET status = 'sent', sent_at = $2, updated_at = $3, last_error = NULL WHERE id = $1"
	now := time.Now()

	pgTx := r.getTransactionOrNil()
	var err error
	if pgTx == nil {
		_, err = r.pool.Exec(ctx, sql, msgID, now, now)
	} else {
		_, err = pgTx.Exec(ctx, sql, msgID, now, now)
	}

	if err != nil {
		trace.InjectError(ctx, err)
		return fmt.Errorf("mark outbox sent: %w", err)
	}
	return nil
}

func (r OutboxRepository) MarkFailed(ctx context.Context, msgID string, lastError string, maxAttempts int) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig(outboxTraceName, "MarkFailed"))
	defer end()

	trace.InjectAttributes(ctx, attr.New("msgID", msgID))

	const sql = "UPDATE outbox SET attempt_counter = attempt_counter + 1, last_error = $2, updated_at = $3, status = CASE WHEN attempt_counter + 1 >= $4 THEN 'error' ELSE status END WHERE id = $1"
	now := time.Now()

	pgTx := r.getTransactionOrNil()
	var err error
	if pgTx == nil {
		_, err = r.pool.Exec(ctx, sql, msgID, lastError, now, maxAttempts)
	} else {
		_, err = pgTx.Exec(ctx, sql, msgID, lastError, now, maxAttempts)
	}

	if err != nil {
		trace.InjectError(ctx, err)
		return fmt.Errorf("mark outbox failed: %w", err)
	}
	return nil
}

func (r OutboxRepository) getTransactionOrNil() pgx.Tx {
	return r.tx
}

func (r OutboxRepository) mapRowsNextValue(rows *pgx.Rows) (*outbox.OutboxMessage, error) {
	result := new(outbox.OutboxMessage)
	err := (*rows).Scan(
		&result.ID,
		&result.Payload,
		&result.Topic,
		&result.Headers,
		&result.Status,
		&result.AttemptCounter,
		&result.LastError,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.SentAt,
		&result.DeletedAt,
	)
	return result, err
}
