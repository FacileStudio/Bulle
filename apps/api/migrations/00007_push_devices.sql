-- +goose Up
CREATE TABLE push_devices (
    token      TEXT   PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind       TEXT   NOT NULL,
    p256dh     TEXT   NOT NULL DEFAULT '',
    auth       TEXT   NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX push_devices_user_idx ON push_devices (user_id);

-- +goose Down
DROP TABLE IF EXISTS push_devices;
