-- +goose Up
CREATE TABLE search_usage (
    user_id BIGINT  NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    day     DATE    NOT NULL,
    count   INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, day)
);

-- +goose Down
DROP TABLE IF EXISTS search_usage;
