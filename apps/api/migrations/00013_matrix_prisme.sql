-- +goose Up
ALTER TABLE matrix_bots
    DROP COLUMN password,
    DROP COLUMN nonce,
    DROP COLUMN allow,
    ADD COLUMN access_token BYTEA,
    ADD COLUMN token_nonce  BYTEA,
    ADD COLUMN device_id    TEXT NOT NULL DEFAULT '';

DROP TABLE matrix_space_rooms;

CREATE TABLE matrix_space_rooms (
    user_id  BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    room_id  TEXT   NOT NULL,
    name     TEXT   NOT NULL DEFAULT '',
    space_id TEXT   NOT NULL REFERENCES spaces (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, room_id)
);

-- +goose Down
DROP TABLE IF EXISTS matrix_space_rooms;

CREATE TABLE matrix_space_rooms (
    space_id TEXT   PRIMARY KEY REFERENCES spaces (id) ON DELETE CASCADE,
    user_id  BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    room_id  TEXT   NOT NULL,
    UNIQUE (user_id, room_id)
);

ALTER TABLE matrix_bots
    DROP COLUMN access_token,
    DROP COLUMN token_nonce,
    DROP COLUMN device_id,
    ADD COLUMN password BYTEA,
    ADD COLUMN nonce    BYTEA,
    ADD COLUMN allow    JSONB NOT NULL DEFAULT '[]';
