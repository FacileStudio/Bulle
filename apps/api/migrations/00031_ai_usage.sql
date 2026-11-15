-- +goose Up
CREATE TABLE ai_usage (
    id                    BIGSERIAL     PRIMARY KEY,
    user_id               BIGINT        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    source                TEXT          NOT NULL CHECK (source IN ('complete', 'summarize', 'decide')),
    caller                TEXT          NOT NULL DEFAULT '',
    model                 TEXT          NOT NULL DEFAULT '',
    input_tokens          BIGINT        NOT NULL DEFAULT 0,
    output_tokens         BIGINT        NOT NULL DEFAULT 0,
    cache_read_tokens     BIGINT        NOT NULL DEFAULT 0,
    cache_creation_tokens BIGINT        NOT NULL DEFAULT 0,
    cost_usd              NUMERIC(12,6) NOT NULL DEFAULT 0,
    own_key               BOOLEAN       NOT NULL DEFAULT false,
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX ai_usage_user_created_idx ON ai_usage (user_id, created_at);

ALTER TABLE runs ADD COLUMN own_key BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE runs DROP COLUMN IF EXISTS own_key;
DROP TABLE IF EXISTS ai_usage;
