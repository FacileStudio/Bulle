-- +goose Up
CREATE TABLE hardware_devices (
    id             TEXT        PRIMARY KEY,
    user_id        BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name           TEXT        NOT NULL,
    device_class   TEXT        NOT NULL DEFAULT 'charm',
    hardware_id    TEXT        NOT NULL,
    token_hash     TEXT        NOT NULL UNIQUE,
    capabilities   JSONB       NOT NULL DEFAULT '[]',
    last_seen_at   TIMESTAMPTZ,
    last_transport TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, hardware_id)
);

CREATE INDEX hardware_devices_user_idx ON hardware_devices (user_id);

-- +goose Down
DROP TABLE IF EXISTS hardware_devices;
