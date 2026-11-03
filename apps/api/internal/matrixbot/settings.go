package matrixbot

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// ErrInvalid is a settings value the bot cannot use.
var ErrInvalid = errors.New("invalid matrix settings")

// Settings is what the web app shows and saves. Password is write-only: it
// is never returned, and an empty one on save keeps the stored password.
type Settings struct {
	Enabled     bool              `json:"enabled"`
	Homeserver  string            `json:"homeserver"`
	User        string            `json:"user_id"`
	Password    string            `json:"password,omitempty"`
	PasswordSet bool              `json:"password_set"`
	Allow       []string          `json:"allow"`
	Rooms       map[string]string `json:"rooms"`
	Status      string            `json:"status"`
}

// Get returns the user's settings and the bot's live status.
func (m *Manager) Get(ctx context.Context, userID int64) (Settings, error) {
	bot, _, err := m.repo.get(ctx, userID)
	if err != nil {
		return Settings{}, err
	}
	rooms, err := m.repo.spaceRooms(ctx, userID)
	if err != nil {
		return Settings{}, err
	}
	return Settings{
		Enabled: bot.Enabled, Homeserver: bot.Homeserver, User: bot.User, PasswordSet: len(bot.Password) > 0,
		Allow: nonNil(bot.Allow), Rooms: rooms, Status: m.statusOf(userID),
	}, nil
}

// Save stores the settings and restarts the bot on them.
func (m *Manager) Save(ctx context.Context, userID int64, in Settings) (Settings, error) {
	if m.seal == nil {
		return Settings{}, ErrDisabled
	}
	in = clean(in)
	if err := validate(in); err != nil {
		return Settings{}, err
	}
	if in.Enabled && in.Password == "" {
		stored, _, err := m.repo.get(ctx, userID)
		if err != nil {
			return Settings{}, err
		}
		if len(stored.Password) == 0 {
			return Settings{}, errors.Join(ErrInvalid, errors.New("an enabled bot needs its password"))
		}
	}
	bot := Bot{UserID: userID, Enabled: in.Enabled, Homeserver: in.Homeserver, User: in.User, Allow: in.Allow}
	if in.Password != "" {
		sealed, nonce, err := m.seal.seal(userID, in.Password)
		if err != nil {
			return Settings{}, err
		}
		bot.Password, bot.Nonce = sealed, nonce
	}
	if err := m.repo.put(ctx, bot); err != nil {
		return Settings{}, err
	}
	if err := m.repo.setSpaceRooms(ctx, userID, in.Rooms); err != nil {
		return Settings{}, err
	}
	if err := m.reload(ctx, userID); err != nil {
		return Settings{}, err
	}
	return m.Get(ctx, userID)
}

func clean(in Settings) Settings {
	in.Homeserver = strings.TrimRight(strings.TrimSpace(in.Homeserver), "/")
	in.User = strings.TrimSpace(in.User)
	allow := make([]string, 0, len(in.Allow))
	for _, who := range in.Allow {
		if who = strings.TrimSpace(who); who != "" {
			allow = append(allow, who)
		}
	}
	in.Allow = allow
	rooms := make(map[string]string, len(in.Rooms))
	for space, room := range in.Rooms {
		rooms[space] = strings.TrimSpace(room)
	}
	in.Rooms = rooms
	return in
}

func validate(in Settings) error {
	if in.Homeserver != "" {
		u, err := url.Parse(in.Homeserver)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return errors.Join(ErrInvalid, errors.New("homeserver must be an https URL"))
		}
	}
	if in.User != "" && !isMXID(in.User) {
		return errors.Join(ErrInvalid, errors.New("bot user id must look like @name:server"))
	}
	for _, who := range in.Allow {
		if !isMXID(who) {
			return errors.Join(ErrInvalid, errors.New(who+" is not a Matrix user id"))
		}
	}
	for _, room := range in.Rooms {
		if room != "" && !strings.HasPrefix(room, "!") {
			return errors.Join(ErrInvalid, errors.New("room ids start with ! (find it under room settings, advanced)"))
		}
	}
	if in.Enabled && (in.Homeserver == "" || in.User == "") {
		return errors.Join(ErrInvalid, errors.New("an enabled bot needs a homeserver and a user id"))
	}
	return nil
}

func isMXID(s string) bool {
	name, server, ok := strings.Cut(s, ":")
	return ok && strings.HasPrefix(name, "@") && len(name) > 1 && server != ""
}
