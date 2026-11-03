-- +goose Up
CREATE TABLE IF NOT EXISTS spaces (
    id                TEXT        PRIMARY KEY,
    user_id           BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name              TEXT        NOT NULL,
    icon              TEXT        NOT NULL DEFAULT '',
    mycelium_space_id TEXT        NOT NULL DEFAULT '',
    opus_workspace    TEXT        NOT NULL DEFAULT '',
    agenda_calendar   BIGINT      NOT NULL DEFAULT 0,
    courrier_account  BIGINT      NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS spaces_user_created_idx ON spaces (user_id, created_at, id);

ALTER TABLE sessions ADD COLUMN IF NOT EXISTS space_id TEXT;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS space_id TEXT;

INSERT INTO spaces (id, user_id, name)
SELECT replace(gen_random_uuid()::text, '-', ''), u.id, 'Personal'
FROM users u
WHERE NOT EXISTS (SELECT 1 FROM spaces s WHERE s.user_id = u.id);

UPDATE sessions se
SET space_id = (SELECT s.id FROM spaces s WHERE s.user_id = se.user_id ORDER BY s.created_at, s.id LIMIT 1)
WHERE se.space_id IS NULL;

DELETE FROM jobs j
WHERE j.space_id IS NULL
  AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id::text = j.user_id);

UPDATE jobs j
SET space_id = (SELECT s.id FROM spaces s WHERE s.user_id::text = j.user_id ORDER BY s.created_at, s.id LIMIT 1)
WHERE j.space_id IS NULL;

ALTER TABLE sessions ALTER COLUMN space_id SET NOT NULL;
ALTER TABLE jobs ALTER COLUMN space_id SET NOT NULL;

ALTER TABLE sessions DROP CONSTRAINT IF EXISTS sessions_space_id_fkey;
ALTER TABLE sessions ADD CONSTRAINT sessions_space_id_fkey FOREIGN KEY (space_id) REFERENCES spaces (id) ON DELETE CASCADE;
ALTER TABLE jobs DROP CONSTRAINT IF EXISTS jobs_space_id_fkey;
ALTER TABLE jobs ADD CONSTRAINT jobs_space_id_fkey FOREIGN KEY (space_id) REFERENCES spaces (id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS sessions_space_updated_idx ON sessions (space_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS jobs_space_idx ON jobs (space_id);

-- +goose Down
ALTER TABLE jobs DROP COLUMN IF EXISTS space_id;
ALTER TABLE sessions DROP COLUMN IF EXISTS space_id;
DROP TABLE IF EXISTS spaces;
