-- +goose Up
CREATE TABLE provider_credentials (
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    backend    TEXT   NOT NULL,
    ciphertext BYTEA  NOT NULL,
    nonce      BYTEA  NOT NULL,
    last4      TEXT   NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, backend)
);

-- +goose Down
DROP TABLE IF EXISTS provider_credentials;
