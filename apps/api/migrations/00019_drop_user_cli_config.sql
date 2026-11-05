-- +goose Up
ALTER TABLE user_preferences DROP COLUMN IF EXISTS cli_config;

-- +goose Down
ALTER TABLE user_preferences ADD COLUMN IF NOT EXISTS cli_config TEXT NOT NULL DEFAULT '';
