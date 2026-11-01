-- +goose Up
CREATE TABLE computers (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT      NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    box_id         TEXT        NOT NULL DEFAULT '',
    status         TEXT        NOT NULL DEFAULT 'creating',
    ssh_host       TEXT        NOT NULL DEFAULT '',
    ssh_port       INTEGER     NOT NULL DEFAULT 0,
    cpus           INTEGER     NOT NULL DEFAULT 0,
    memory_mb      INTEGER     NOT NULL DEFAULT 0,
    disk_gb        INTEGER     NOT NULL DEFAULT 0,
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX computers_status_active_idx ON computers (status, last_active_at);

-- +goose Down
DROP TABLE IF EXISTS computers;
