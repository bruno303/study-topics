-- +goose Up
CREATE TABLE IF NOT EXISTS outbox (
    id              VARCHAR(36)  PRIMARY KEY,
    payload         JSONB        NOT NULL,
    topic           TEXT         NOT NULL,
    headers         JSONB        DEFAULT '{}'::jsonb,
    status          VARCHAR(10)  NOT NULL DEFAULT 'pending',
    attempt_counter INT          NOT NULL DEFAULT 0,
    last_error      TEXT,
    created_at      TIMESTAMPTZ  NOT NULL,
    updated_at      TIMESTAMPTZ  NOT NULL,
    sent_at         TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_outbox_status_created_at ON outbox (status, created_at);

-- +goose Down
DROP TABLE IF EXISTS outbox;
