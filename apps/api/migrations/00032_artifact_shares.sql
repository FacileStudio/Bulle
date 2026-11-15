-- +goose Up
CREATE TABLE artifact_shares (
    token          TEXT        PRIMARY KEY,
    user_id        BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    artifact_id    TEXT        NOT NULL,
    mycelium_space TEXT        NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at     TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX artifact_shares_live_idx
    ON artifact_shares (user_id, artifact_id, mycelium_space)
    WHERE revoked_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS artifact_shares;
