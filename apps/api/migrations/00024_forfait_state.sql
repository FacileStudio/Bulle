-- +goose Up
CREATE TABLE IF NOT EXISTS forfait_state (
    id         SMALLINT    PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    document   TEXT        NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS forfait_state;
