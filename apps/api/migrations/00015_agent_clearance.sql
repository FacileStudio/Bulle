-- +goose Up
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS sessions_channel_check;
ALTER TABLE sessions ADD CONSTRAINT sessions_channel_check CHECK (channel IN ('web', 'matrix', 'cron', 'guest'));

ALTER TABLE matrix_bots ADD COLUMN IF NOT EXISTS owner_user TEXT NOT NULL DEFAULT '';

ALTER TABLE matrix_threads ADD COLUMN IF NOT EXISTS guest BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE matrix_threads DROP CONSTRAINT matrix_threads_pkey;
ALTER TABLE matrix_threads ADD PRIMARY KEY (user_id, room_id, thread_id, guest);

-- +goose Down
DELETE FROM matrix_threads WHERE guest;
ALTER TABLE matrix_threads DROP CONSTRAINT matrix_threads_pkey;
ALTER TABLE matrix_threads ADD PRIMARY KEY (user_id, room_id, thread_id);
ALTER TABLE matrix_threads DROP COLUMN IF EXISTS guest;

ALTER TABLE matrix_bots DROP COLUMN IF EXISTS owner_user;

DELETE FROM sessions WHERE channel = 'guest';
ALTER TABLE sessions DROP CONSTRAINT IF EXISTS sessions_channel_check;
ALTER TABLE sessions ADD CONSTRAINT sessions_channel_check CHECK (channel IN ('web', 'matrix', 'cron'));
