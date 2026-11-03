-- +goose Up
CREATE TABLE matrix_bots (
    user_id    BIGINT  PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    enabled    BOOLEAN NOT NULL DEFAULT false,
    homeserver TEXT    NOT NULL DEFAULT '',
    bot_user   TEXT    NOT NULL DEFAULT '',
    password   BYTEA,
    nonce      BYTEA,
    allow      JSONB   NOT NULL DEFAULT '[]',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE matrix_space_rooms (
    space_id TEXT   PRIMARY KEY REFERENCES spaces (id) ON DELETE CASCADE,
    user_id  BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    room_id  TEXT   NOT NULL,
    UNIQUE (user_id, room_id)
);

CREATE TABLE matrix_threads (
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    room_id    TEXT   NOT NULL,
    thread_id  TEXT   NOT NULL,
    session_id TEXT   NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, room_id, thread_id)
);

-- +goose Down
DROP TABLE IF EXISTS matrix_threads;
DROP TABLE IF EXISTS matrix_space_rooms;
DROP TABLE IF EXISTS matrix_bots;
