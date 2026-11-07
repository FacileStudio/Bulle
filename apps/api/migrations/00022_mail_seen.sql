-- +goose Up
CREATE TABLE mail_seen (
    user_id  BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    email_id BIGINT      NOT NULL,
    seen_at  TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, email_id)
);
CREATE INDEX mail_seen_user_seen_at ON mail_seen (user_id, seen_at);

-- +goose Down
DROP TABLE IF EXISTS mail_seen;
