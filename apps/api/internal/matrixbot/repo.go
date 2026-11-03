package matrixbot

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Bot is one user's agent account on the chat homeserver. Token is sealed.
// Owner is the Matrix id of the user the agent works for, empty until Prisme
// has named it.
type Bot struct {
	UserID     int64
	Enabled    bool
	Homeserver string
	User       string
	Owner      string
	Token      []byte
	Nonce      []byte
	DeviceID   string
	UpdatedAt  time.Time
}

// Room is a chat room the agent sits in and the Bulle space it works in.
type Room struct {
	RoomID  string `json:"room_id"`
	Name    string `json:"name"`
	SpaceID string `json:"space_id"`
}

type user struct {
	ID    int64
	Email string
}

type repo struct{ db *sql.DB }

const botColumns = `user_id, enabled, homeserver, bot_user, owner_user, access_token, token_nonce, device_id, updated_at`

func scanBot(row interface{ Scan(...any) error }) (Bot, error) {
	var b Bot
	err := row.Scan(&b.UserID, &b.Enabled, &b.Homeserver, &b.User, &b.Owner, &b.Token, &b.Nonce, &b.DeviceID, &b.UpdatedAt)
	return b, err
}

func (r repo) get(ctx context.Context, userID int64) (Bot, bool, error) {
	b, err := scanBot(r.db.QueryRowContext(ctx, `SELECT `+botColumns+` FROM matrix_bots WHERE user_id = $1`, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Bot{UserID: userID}, false, nil
	}
	return b, err == nil, err
}

func (r repo) put(ctx context.Context, b Bot) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO matrix_bots (user_id, enabled, homeserver, bot_user, owner_user, access_token, token_nonce, device_id, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
		ON CONFLICT (user_id) DO UPDATE SET
			enabled = EXCLUDED.enabled, homeserver = EXCLUDED.homeserver, bot_user = EXCLUDED.bot_user, owner_user = EXCLUDED.owner_user,
			access_token = EXCLUDED.access_token, token_nonce = EXCLUDED.token_nonce,
			device_id = EXCLUDED.device_id, updated_at = now()`,
		b.UserID, b.Enabled, b.Homeserver, b.User, b.Owner, b.Token, b.Nonce, b.DeviceID)
	return err
}

func (r repo) users(ctx context.Context) ([]user, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, email FROM users WHERE email <> '' ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []user
	for rows.Next() {
		var u user
		if err := rows.Scan(&u.ID, &u.Email); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r repo) threadSession(ctx context.Context, userID int64, room, thread string, guest bool) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `SELECT session_id FROM matrix_threads WHERE user_id = $1 AND room_id = $2 AND thread_id = $3 AND guest = $4`,
		userID, room, thread, guest).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (r repo) setThreadSession(ctx context.Context, userID int64, room, thread string, guest bool, sessionID string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO matrix_threads (user_id, room_id, thread_id, guest, session_id) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, room_id, thread_id, guest) DO UPDATE SET session_id = EXCLUDED.session_id`, userID, room, thread, guest, sessionID)
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

func (r repo) rooms(ctx context.Context, userID int64) ([]Room, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT room_id, name, space_id FROM matrix_space_rooms WHERE user_id = $1 ORDER BY name`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []Room{}
	for rows.Next() {
		var room Room
		if err := rows.Scan(&room.RoomID, &room.Name, &room.SpaceID); err != nil {
			return nil, err
		}
		out = append(out, room)
	}
	return out, rows.Err()
}

// setRooms replaces the user's rooms in one transaction.
func (r repo) setRooms(ctx context.Context, userID int64, rooms []Room) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM matrix_space_rooms WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, room := range rooms {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO matrix_space_rooms (user_id, room_id, name, space_id) VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id, room_id) DO NOTHING`, userID, room.RoomID, room.Name, room.SpaceID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
