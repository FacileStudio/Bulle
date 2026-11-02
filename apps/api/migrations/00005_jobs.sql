-- +goose Up
CREATE TABLE jobs (
    id         TEXT        PRIMARY KEY,
    user_id    TEXT        NOT NULL,
    name       TEXT        NOT NULL,
    schedule   TEXT        NOT NULL,
    prompt     TEXT        NOT NULL,
    profile    TEXT,
    delivery   TEXT,
    enabled    BOOLEAN     NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX jobs_user_idx ON jobs (user_id);

CREATE TABLE job_runs (
    id          TEXT        PRIMARY KEY,
    job_id      TEXT        NOT NULL REFERENCES jobs (id) ON DELETE CASCADE,
    run_id      TEXT        NOT NULL REFERENCES runs (id) ON DELETE CASCADE,
    status      TEXT        NOT NULL,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX job_runs_job_idx ON job_runs (job_id, started_at DESC);

-- +goose Down
DROP TABLE IF EXISTS job_runs;
DROP TABLE IF EXISTS jobs;
