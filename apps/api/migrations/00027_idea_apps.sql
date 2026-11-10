-- +goose Up
ALTER TABLE ideas ADD COLUMN IF NOT EXISTS app TEXT NOT NULL DEFAULT 'bulle';

UPDATE ideas SET app = CASE lower(split_part(fingerprint, ':', 1))
    WHEN 'memory' THEN 'mycelium'
    ELSE lower(split_part(fingerprint, ':', 1))
END
WHERE lower(split_part(fingerprint, ':', 1)) IN ('courrier', 'agenda', 'opus', 'sablier', 'mycelium', 'memory', 'nuage', 'glouton');

CREATE INDEX IF NOT EXISTS ideas_space_app_status ON ideas (user_id, space_id, app, status);

ALTER TABLE idea_batches ADD COLUMN IF NOT EXISTS app TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS idea_batches_one_running_per_app ON idea_batches (space_id, app) WHERE status = 'running';
DROP INDEX IF EXISTS idea_batches_one_running;

CREATE INDEX IF NOT EXISTS idea_batches_user_app_created ON idea_batches (user_id, app, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idea_batches_user_app_created;
UPDATE idea_batches SET status = 'error', finished_at = now() WHERE status = 'running' AND app <> '';
CREATE UNIQUE INDEX IF NOT EXISTS idea_batches_one_running ON idea_batches (space_id) WHERE status = 'running';
DROP INDEX IF EXISTS idea_batches_one_running_per_app;
ALTER TABLE idea_batches DROP COLUMN IF EXISTS app;

DROP INDEX IF EXISTS ideas_space_app_status;
ALTER TABLE ideas DROP COLUMN IF EXISTS app;
