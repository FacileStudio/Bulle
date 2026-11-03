-- +goose Up
ALTER TABLE users ADD COLUMN picture_url TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS picture_url;
