-- +goose Up
CREATE TABLE sessions (
    id         TEXT        PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    channel    TEXT        NOT NULL DEFAULT 'web' CHECK (channel IN ('web', 'matrix', 'cron')),
    title      TEXT        NOT NULL DEFAULT '',
    backend    TEXT        NOT NULL DEFAULT '',
    model      TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX sessions_user_updated_idx ON sessions (user_id, updated_at DESC);

CREATE TABLE messages (
    session_id TEXT   NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    seq        INT    NOT NULL,
    payload    JSONB  NOT NULL,
    PRIMARY KEY (session_id, seq)
);

CREATE TABLE runs (
    id          TEXT          PRIMARY KEY,
    session_id  TEXT          NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    status      TEXT          NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'finished', 'stopped', 'error', 'interrupted')),
    stop_reason TEXT          NOT NULL DEFAULT '',
    usage       JSONB         NOT NULL DEFAULT '{}'::jsonb,
    cost_usd    NUMERIC(12,6) NOT NULL DEFAULT 0,
    started_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX runs_session_idx ON runs (session_id, started_at DESC);
CREATE UNIQUE INDEX runs_one_active_per_session_idx ON runs (session_id) WHERE status = 'running';

CREATE TABLE run_events (
    run_id  TEXT  NOT NULL REFERENCES runs (id) ON DELETE CASCADE,
    seq     INT   NOT NULL,
    type    TEXT  NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    PRIMARY KEY (run_id, seq)
);

CREATE TABLE approvals (
    run_id       TEXT        NOT NULL REFERENCES runs (id) ON DELETE CASCADE,
    tool_call_id TEXT        NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'denied')),
    decided_by   TEXT        NOT NULL DEFAULT '',
    decided_at   TIMESTAMPTZ,
    PRIMARY KEY (run_id, tool_call_id)
);

-- +goose Down
DROP TABLE IF EXISTS approvals;
DROP TABLE IF EXISTS run_events;
DROP TABLE IF EXISTS runs;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS sessions;
