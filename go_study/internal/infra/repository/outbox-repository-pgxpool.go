package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/bruno303/study-topics/go-study/internal/application/model"
	applicationRepository "github.com/bruno303/study-topics/go-study/internal/application/repository"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace"
	"github.com/bruno303/study-topics/go-study/internal/crosscutting/observability/trace/attr"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxPgxRepository struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

const outboxTraceName = "OutboxPgxRepository"

var _ applicationRepository.OutboxRepository = (*OutboxPgxRepository)(nil)

func NewOutboxPgxRepository(pool *pgxpool.Pool) OutboxPgxRepository {
	return newOutboxPgxRepository(pool, nil)
}

func newOutboxPgxRepository(pool *pgxpool.Pool, tx pgx.Tx) OutboxPgxRepository {
	return OutboxPgxRepository{
		pool: pool,
		tx:   tx,
	}
}

func (r OutboxPgxRepository) Enqueue(ctx context.Context, message *model.OutboxMessage) (*model.OutboxMessage, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(outboxTraceName, "Enqueue"))
	defer end()

	if message == nil {
		err := fmt.Errorf("outbox message is nil")
		trace.InjectError(ctx, err)
		return nil, err
	}

	now := time.Now().UTC()
	if message.Status == "" {
		message.Status = model.OutboxStatusPending
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = now
	}
	if message.UpdatedAt.IsZero() {
		message.UpdatedAt = now
	}
	if message.NextAttempt.IsZero() {
		message.NextAttempt = now
	}

	headers, err := marshalHeaders(message.Headers)
	if err != nil {
		trace.InjectError(ctx, err)
		return nil, fmt.Errorf("marshal outbox headers: %w", err)
	}

	trace.InjectAttributes(ctx, attr.New("id", message.Id), attr.New("topic", message.Topic))

	const sql = `
		INSERT INTO OUTBOX (ID, TOPIC, MESSAGE_KEY, PAYLOAD, HEADERS, STATUS, ATTEMPT, NEXT_ATTEMPT, PUBLISHED_AT, LAST_ERROR, CREATED_AT, UPDATED_AT, DELETED_AT)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	params := []any{
		message.Id,
		message.Topic,
		message.MessageKey,
		message.Payload,
		headers,
		string(message.Status),
		message.Attempt,
		message.NextAttempt,
		message.PublishedAt,
		message.LastError,
		message.CreatedAt,
		message.UpdatedAt,
		message.DeletedAt,
	}

	execErr := r.exec(ctx, sql, params...)
	if execErr != nil {
		trace.InjectError(ctx, execErr)
		return nil, fmt.Errorf("enqueue outbox message: %w", execErr)
	}

	return message, nil
}

func (r OutboxPgxRepository) ListPending(ctx context.Context, limit int, maxAttempts int, now time.Time) ([]model.OutboxMessage, error) {
	ctx, end := trace.Trace(ctx, trace.NameConfig(outboxTraceName, "ListPending"))
	defer end()

	trace.InjectAttributes(ctx, attr.New("limit", strconv.Itoa(limit)), attr.New("max_attempts", strconv.Itoa(maxAttempts)))
	if limit <= 0 {
		return []model.OutboxMessage{}, nil
	}

	processingReclaimBefore := now.Add(-outboxProcessingRecoveryWindow)

	const sql = `
		WITH selected AS (
			SELECT ID
			FROM OUTBOX
			WHERE (
					(STATUS IN ($1, $2) AND NEXT_ATTEMPT <= $3)
					OR (STATUS = $6 AND UPDATED_AT <= $7)
			  )
			  AND ($4 <= 0 OR ATTEMPT < $4)
			  AND PUBLISHED_AT IS NULL
			  AND DELETED_AT IS NULL
			ORDER BY CREATED_AT ASC
			LIMIT $5
			FOR UPDATE SKIP LOCKED
		),
		claimed AS (
			UPDATE OUTBOX o
			SET STATUS = $6,
			    ATTEMPT = o.ATTEMPT + 1,
			    UPDATED_AT = $3
			FROM selected s
			WHERE o.ID = s.ID
			RETURNING o.ID, o.TOPIC, o.MESSAGE_KEY, o.PAYLOAD, o.HEADERS, o.STATUS, o.ATTEMPT, o.NEXT_ATTEMPT, o.PUBLISHED_AT, o.LAST_ERROR, o.CREATED_AT, o.UPDATED_AT, o.DELETED_AT
		)
		SELECT ID, TOPIC, MESSAGE_KEY, PAYLOAD, HEADERS, STATUS, ATTEMPT, NEXT_ATTEMPT, PUBLISHED_AT, LAST_ERROR, CREATED_AT, UPDATED_AT, DELETED_AT
		FROM claimed
		ORDER BY CREATED_AT ASC
	`

	rows, err := r.query(ctx, sql,
		string(model.OutboxStatusPending),
		string(model.OutboxStatusError),
		now,
		maxAttempts,
		limit,
		string(model.OutboxStatusProcessing),
		processingReclaimBefore,
	)
	if err != nil {
		trace.InjectError(ctx, err)
		return nil, fmt.Errorf("list pending outbox messages: %w", err)
	}
	defer rows.Close()

	result := make([]model.OutboxMessage, 0)
	for rows.Next() {
		message, mapErr := r.mapRowsNextValue(&rows)
		if mapErr != nil {
			trace.InjectError(ctx, mapErr)
			return nil, fmt.Errorf("scan claimed outbox row: %w", mapErr)
		}
		result = append(result, *message)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		trace.InjectError(ctx, rowsErr)
		return nil, fmt.Errorf("iterate claimed outbox rows: %w", rowsErr)
	}

	return result, nil
}

func (r OutboxPgxRepository) MarkAsPublished(ctx context.Context, id string, publishedAt time.Time) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig(outboxTraceName, "MarkAsPublished"))
	defer end()

	trace.InjectAttributes(ctx, attr.New("id", id))

	const sql = `
		UPDATE OUTBOX
		SET STATUS = $2,
		    PUBLISHED_AT = $3,
		    LAST_ERROR = NULL,
		    UPDATED_AT = $4
		WHERE ID = $1
		  AND STATUS = $5
		  AND PUBLISHED_AT IS NULL
		  AND DELETED_AT IS NULL
	`

	err := r.exec(ctx, sql,
		id,
		string(model.OutboxStatusPublished),
		publishedAt,
		time.Now().UTC(),
		string(model.OutboxStatusProcessing),
	)
	if err != nil {
		trace.InjectError(ctx, err)
		return fmt.Errorf("mark outbox message as published: %w", err)
	}

	return nil
}

func (r OutboxPgxRepository) MarkAsError(ctx context.Context, id string, lastError string, nextAttempt time.Time) error {
	ctx, end := trace.Trace(ctx, trace.NameConfig(outboxTraceName, "MarkAsError"))
	defer end()

	trace.InjectAttributes(ctx, attr.New("id", id))

	status := model.OutboxStatusError
	if nextAttempt.IsZero() {
		status = model.OutboxStatusFailed
	}

	const sql = `
		UPDATE OUTBOX
		SET STATUS = $2,
		    LAST_ERROR = $3,
		    NEXT_ATTEMPT = $4,
		    UPDATED_AT = $5
		WHERE ID = $1
		  AND STATUS = $6
		  AND PUBLISHED_AT IS NULL
		  AND DELETED_AT IS NULL
	`

	err := r.exec(ctx, sql, id, string(status), lastError, nextAttempt, time.Now().UTC(), string(model.OutboxStatusProcessing))
	if err != nil {
		trace.InjectError(ctx, err)
		return fmt.Errorf("mark outbox message as error: %w", err)
	}

	return nil
}

func (r OutboxPgxRepository) getTransactionOrNil() pgx.Tx {
	return r.tx
}

func (r OutboxPgxRepository) exec(ctx context.Context, sql string, args ...any) error {
	pgTx := r.getTransactionOrNil()
	if pgTx == nil {
		_, err := r.pool.Exec(ctx, sql, args...)
		return err
	}

	_, err := pgTx.Exec(ctx, sql, args...)
	return err
}

func (r OutboxPgxRepository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	pgTx := r.getTransactionOrNil()
	if pgTx == nil {
		return r.pool.Query(ctx, sql, args...)
	}

	return pgTx.Query(ctx, sql, args...)
}

func (r OutboxPgxRepository) mapRowsNextValue(rows *pgx.Rows) (*model.OutboxMessage, error) {
	result := new(model.OutboxMessage)
	var headersRaw []byte
	err := (*rows).Scan(
		&result.Id,
		&result.Topic,
		&result.MessageKey,
		&result.Payload,
		&headersRaw,
		&result.Status,
		&result.Attempt,
		&result.NextAttempt,
		&result.PublishedAt,
		&result.LastError,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	headers, err := unmarshalHeaders(headersRaw)
	if err != nil {
		return nil, fmt.Errorf("unmarshal outbox headers: %w", err)
	}
	result.Headers = headers

	return result, nil
}

func marshalHeaders(headers map[string]string) ([]byte, error) {
	if headers == nil {
		return nil, nil
	}

	return json.Marshal(headers)
}

func unmarshalHeaders(raw []byte) (map[string]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	result := make(map[string]string)
	err := json.Unmarshal(raw, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
