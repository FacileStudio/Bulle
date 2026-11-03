package matrixbot

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// ErrUnknownSpace is a link to a space the user does not own.
var ErrUnknownSpace = errors.New("unknown space")

// Bot is one user's Matrix bot as stored. Password is sealed.
type Bot struct {
	UserID     int64
	Enabled    bool
	Homeserver string
	User       string
	Password   []byte
	Nonce      []byte
	Allow      []string
	UpdatedAt  time.Time
}

type repo struct{ db *sql.DB }

const botColumns = `user_id, enabled, homeserver, bot_user, password, nonce, allow, updated_at`

func scanBot(row interface{ Scan(...any) error }) (Bot, error) {
	var b Bot
	var allow []byte
	err := row.Scan(&b.UserID, &b.Enabled, &b.Homeserver, &b.User, &b.Password, &b.Nonce, &allow, &b.UpdatedAt)
	if err != nil {
		return Bot{}, err
	}
	return b, json.Unmarshal(allow, &b.Allow)
}

func (r repo) get(ctx context.Context, userID int64) (Bot, bool, error) {
	b, err := scanBot(r.db.QueryRowContext(ctx, `SELECT `+botColumns+` FROM matrix_bots WHERE user_id = $1`, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Bot{UserID: userID}, false, nil
	}
	return b, err == nil, err
}

func (r repo) enabled(ctx context.Context) ([]Bot, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+botColumns+` FROM matrix_bots WHERE enabled`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []Bot
	for rows.Next() {
		b, err := scanBot(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r repo) put(ctx context.Context, b Bot) error {
	allow, err := json.Marshal(nonNil(b.Allow))
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO matrix_bots (user_id, enabled, homeserver, bot_user, password, nonce, allow, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now())
		ON CONFLICT (user_id) DO UPDATE SET
			enabled = EXCLUDED.enabled, homeserver = EXCLUDED.homeserver, bot_user = EXCLUDED.bot_user,
			password = COALESCE(EXCLUDED.password, matrix_bots.password),
			nonce = COALESCE(EXCLUDED.nonce, matrix_bots.nonce),
			allow = EXCLUDED.allow, updated_at = now()`,
		b.UserID, b.Enabled, b.Homeserver, b.User, b.Password, b.Nonce, allow)
	return err
}

func (r repo) threadSession(ctx context.Context, userID int64, room, thread string) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `SELECT session_id FROM matrix_threads WHERE user_id = $1 AND room_id = $2 AND thread_id = $3`,
		userID, room, thread).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (r repo) setThreadSession(ctx context.Context, userID int64, room, thread, sessionID string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO matrix_threads (user_id, room_id, thread_id, session_id) VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, room_id, thread_id) DO UPDATE SET session_id = EXCLUDED.session_id`, userID, room, thread, sessionID)
	return err
}

func (r repo) roomSpace(ctx context.Context, userID int64, room string) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `SELECT space_id FROM matrix_space_rooms WHERE user_id = $1 AND room_id = $2`, userID, room).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (r repo) spaceRooms(ctx context.Context, userID int64) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT space_id, room_id FROM matrix_space_rooms WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[string]string{}
	for rows.Next() {
		var space, room string
		if err := rows.Scan(&space, &room); err != nil {
			return nil, err
		}
		out[space] = room
	}
	return out, rows.Err()
}

// setSpaceRooms replaces the user's links in one transaction: an empty room
// unlinks its space, and the space ids are checked against the user's own.
func (r repo) setSpaceRooms(ctx context.Context, userID int64, links map[string]string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM matrix_space_rooms WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for space, room := range links {
		if room == "" {
			continue
		}
		res, err := tx.ExecContext(ctx, `
			INSERT INTO matrix_space_rooms (space_id, user_id, room_id)
			SELECT id, user_id, $3 FROM spaces WHERE id = $1 AND user_id = $2`, space, userID, room)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrUnknownSpace
		}
	}
	return tx.Commit()
}

func nonNil(list []string) []string {
	if list == nil {
		return []string{}
	}
	return list
}
