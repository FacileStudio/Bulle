-- +goose Up
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS sessions_channel_check;
ALTER TABLE sessions ADD CONSTRAINT sessions_channel_check CHECK (channel IN ('web', 'matrix', 'cron', 'guest', 'ideas'));

CREATE TABLE ideas (
    id          TEXT        PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    space_id    TEXT        NOT NULL REFERENCES spaces (id) ON DELETE CASCADE,
    batch_id    TEXT        NOT NULL,
    kind        TEXT        NOT NULL,
    title       TEXT        NOT NULL,
    why         TEXT        NOT NULL,
    prompt      TEXT        NOT NULL,
    draft       TEXT        NOT NULL DEFAULT '',
    sources     JSONB       NOT NULL DEFAULT '[]',
    fingerprint TEXT        NOT NULL,
    status      TEXT        NOT NULL DEFAULT 'new',
    feedback    TEXT        NOT NULL DEFAULT '',
    reason      TEXT        NOT NULL DEFAULT '',
    session_id  TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX ideas_space_status ON ideas (user_id, space_id, status);
CREATE INDEX ideas_fingerprint ON ideas (space_id, fingerprint);

CREATE TABLE idea_settings (
    space_id        TEXT        PRIMARY KEY REFERENCES spaces (id) ON DELETE CASCADE,
    user_id         BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    enabled         BOOLEAN     NOT NULL DEFAULT TRUE,
    hour            INT         NOT NULL DEFAULT 8,
    timezone        TEXT        NOT NULL DEFAULT 'UTC',
    push            BOOLEAN     NOT NULL DEFAULT FALSE,
    steering        TEXT        NOT NULL DEFAULT '',
    last_batch_date TEXT        NOT NULL DEFAULT '',
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE idea_batches (
    id          TEXT        PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    space_id    TEXT        NOT NULL REFERENCES spaces (id) ON DELETE CASCADE,
    run_id      TEXT        NOT NULL DEFAULT '',
    trigger     TEXT        NOT NULL,
    status      TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX idea_batches_space_created ON idea_batches (user_id, space_id, created_at DESC);
CREATE UNIQUE INDEX idea_batches_one_running ON idea_batches (space_id) WHERE status = 'running';

CREATE TABLE goals (
    id         TEXT        PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    space_id   TEXT        NOT NULL REFERENCES spaces (id) ON DELETE CASCADE,
    title      TEXT        NOT NULL,
    category   TEXT        NOT NULL,
    done       BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX goals_space ON goals (user_id, space_id);

-- +goose Down
DROP TABLE IF EXISTS goals;
DROP TABLE IF EXISTS idea_batches;
DROP TABLE IF EXISTS idea_settings;
DROP TABLE IF EXISTS ideas;

DELETE FROM sessions WHERE channel = 'ideas';
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS sessions_channel_check;
ALTER TABLE sessions ADD CONSTRAINT sessions_channel_check CHECK (channel IN ('web', 'matrix', 'cron', 'guest'));
