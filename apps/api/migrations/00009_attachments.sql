-- +goose Up
CREATE TABLE attachments (
    id         TEXT   PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name       TEXT   NOT NULL,
    media_type TEXT   NOT NULL,
    size       BIGINT NOT NULL,
    data       BYTEA  NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX attachments_user_idx ON attachments (user_id);

-- +goose Down
DROP TABLE IF EXISTS attachments;
