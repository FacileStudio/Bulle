-- +goose Up
ALTER TABLE user_preferences
    ADD COLUMN computer_network TEXT NOT NULL DEFAULT 'open'
    CHECK (computer_network IN ('open', 'offline'));

-- +goose Down
ALTER TABLE user_preferences DROP COLUMN IF EXISTS computer_network;
