-- +goose Up
ALTER TABLE porte_login_codes
    ADD COLUMN IF NOT EXISTS provider TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS subject  TEXT NOT NULL DEFAULT '';

ALTER TABLE users ADD COLUMN IF NOT EXISTS locale TEXT NOT NULL DEFAULT 'en';

CREATE TABLE IF NOT EXISTS directory_deletion_state (
    id         SMALLINT    PRIMARY KEY CHECK (id = 1),
    state      BYTEA       NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS directory_deletion_state;
ALTER TABLE users DROP COLUMN IF EXISTS locale;
ALTER TABLE porte_login_codes DROP COLUMN IF EXISTS subject, DROP COLUMN IF EXISTS provider;
