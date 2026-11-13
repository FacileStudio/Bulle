-- +goose Up
ALTER TABLE sessions ADD COLUMN seen_at TIMESTAMPTZ;
UPDATE sessions SET seen_at = updated_at;

-- +goose Down
ALTER TABLE sessions DROP COLUMN IF EXISTS seen_at;
