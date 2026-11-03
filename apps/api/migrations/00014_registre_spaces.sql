-- +goose Up
ALTER TABLE spaces ADD COLUMN IF NOT EXISTS registre_space_id TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS spaces_user_registre_idx
    ON spaces (user_id, registre_space_id) WHERE registre_space_id <> '';

-- +goose Down
DROP INDEX IF EXISTS spaces_user_registre_idx;
ALTER TABLE spaces DROP COLUMN IF EXISTS registre_space_id;
