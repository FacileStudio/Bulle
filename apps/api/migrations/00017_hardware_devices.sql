-- +goose Up
CREATE TABLE hardware_devices (
    id             TEXT        PRIMARY KEY,
    user_id        BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name           TEXT        NOT NULL,
    device_class   TEXT        NOT NULL DEFAULT 'charm' CHECK (device_class IN ('charm', 'field', 'totem')),
    hardware_id    TEXT        NOT NULL CHECK (hardware_id ~ '^[0-9a-f]{8}$'),
    token_hash     TEXT        NOT NULL UNIQUE,
    capabilities   JSONB       NOT NULL DEFAULT '[]',
    last_seen_at   TIMESTAMPTZ,
    last_transport TEXT        CHECK (last_transport IN ('ble', 'wifi', 'usb')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, hardware_id)
);

-- +goose Down
DROP TABLE IF EXISTS hardware_devices;
