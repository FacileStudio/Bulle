-- +goose Up
ALTER TABLE user_preferences DROP CONSTRAINT IF EXISTS user_preferences_computer_network_check;
ALTER TABLE user_preferences
    ADD CONSTRAINT user_preferences_computer_network_check
    CHECK (computer_network IN ('', 'open', 'offline', 'allowlist'));
ALTER TABLE user_preferences ALTER COLUMN computer_network SET DEFAULT '';
ALTER TABLE user_preferences
    ADD COLUMN computer_net_allow JSONB NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE user_preferences DROP COLUMN IF EXISTS computer_net_allow;
UPDATE user_preferences SET computer_network = 'open' WHERE computer_network IN ('', 'allowlist');
ALTER TABLE user_preferences ALTER COLUMN computer_network SET DEFAULT 'open';
ALTER TABLE user_preferences DROP CONSTRAINT IF EXISTS user_preferences_computer_network_check;
ALTER TABLE user_preferences
    ADD CONSTRAINT user_preferences_computer_network_check
    CHECK (computer_network IN ('open', 'offline'));
