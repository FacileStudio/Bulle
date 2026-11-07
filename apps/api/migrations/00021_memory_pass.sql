-- +goose Up
ALTER TABLE sessions ADD COLUMN memory_seen_seq INT NOT NULL DEFAULT 0;
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS sessions_channel_check;
ALTER TABLE sessions ADD CONSTRAINT sessions_channel_check CHECK (channel IN ('web', 'matrix', 'cron', 'guest', 'ideas', 'memory'));

-- +goose Down
DELETE FROM sessions WHERE channel = 'memory';
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS sessions_channel_check;
ALTER TABLE sessions ADD CONSTRAINT sessions_channel_check CHECK (channel IN ('web', 'matrix', 'cron', 'guest', 'ideas'));
ALTER TABLE sessions DROP COLUMN IF EXISTS memory_seen_seq;
