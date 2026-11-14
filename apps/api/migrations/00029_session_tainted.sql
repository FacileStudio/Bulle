-- +goose Up
ALTER TABLE sessions ADD COLUMN tainted BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE sessions DROP COLUMN IF EXISTS tainted;
