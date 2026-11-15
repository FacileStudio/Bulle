-- +goose Up
CREATE TABLE ai_daily_usage (
    user_id     BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    day         DATE   NOT NULL,
    completions INT    NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, day)
);

-- +goose Down
DROP TABLE IF EXISTS ai_daily_usage;
