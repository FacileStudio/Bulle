package matrixbot

import (
	"cmp"
	"context"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/prisme"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

const provisionEvery = 10 * time.Minute

// provisionLoop keeps every user's agent account and rooms in step with
// Prisme: new users get an agent, renamed assistants are renamed, and new
// spaces and channels are joined, with nobody configuring anything.
func (m *Manager) provisionLoop(ctx context.Context) {
	defer m.wg.Done()
	for {
		m.provisionAll(ctx)
		select {
		case <-ctx.Done():
			return
		case <-time.After(provisionEvery):
		}
	}
}

func (m *Manager) provisionAll(ctx context.Context) {
	users, err := m.repo.users(ctx)
	if err != nil {
		m.cfg.Logger.Error("matrix: listing users", slog.Any("error", err))
		return
	}
	for _, u := range users {
		if err := m.provision(ctx, u); err != nil {
			m.setStatus(u.ID, "error: "+err.Error())
			m.cfg.Logger.Warn("matrix: provisioning agent", slog.Int64("user_id", u.ID), slog.Any("error", err))
		}
	}
}

func (m *Manager) provision(ctx context.Context, u user) error {
	name, avatar := "Bulle", ""
	if profile, err := m.cfg.Store.GetAssistantProfile(ctx, u.ID); err == nil {
		name = cmp.Or(strings.TrimSpace(profile.Name), name)
		avatar = agentAvatar(profile.AvatarURL)
	}
	account, err := m.cfg.Prisme.Agent(ctx, u.Email, name, avatar)
	if err != nil {
		return err
	}
	rooms, err := m.mapRooms(ctx, u)
	if err != nil {
		return err
	}
	ids := make([]string, len(rooms))
	for i, room := range rooms {
		ids[i] = room.RoomID
	}
	if err := m.cfg.Prisme.SetRooms(ctx, u.Email, account.UserID, ids); err != nil {
		return err
	}
	if err := m.repo.setRooms(ctx, u.ID, rooms); err != nil {
		return err
	}
	return m.adopt(ctx, u.ID, account)
}

// mapRooms pairs each Prisme channel with the Bulle space of the same name,
// or the user's default space when none matches.
func (m *Manager) mapRooms(ctx context.Context, u user) ([]Room, error) {
	spaces, err := m.cfg.Prisme.Spaces(ctx, u.Email)
	if err != nil {
		return nil, err
	}
	local, err := m.cfg.Store.ListSpaces(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	fallback, err := m.cfg.Store.EnsureDefaultSpace(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	var rooms []Room
	for _, space := range spaces {
		target := matchSpace(local, space.Name, fallback.ID)
		for _, ch := range space.Channels {
			rooms = append(rooms, Room{RoomID: ch.RoomID, Name: space.Name + " / " + ch.Name, SpaceID: target})
		}
	}
	return rooms, nil
}

func matchSpace(local []store.Space, name, fallback string) string {
	for _, s := range local {
		if strings.EqualFold(strings.TrimSpace(s.Name), strings.TrimSpace(name)) {
			return s.ID
		}
	}
	return fallback
}

// adopt stores the agent's login and restarts its bot when the login changed
// or the bot is not running.
func (m *Manager) adopt(ctx context.Context, userID int64, account prisme.Account) error {
	stored, _, err := m.repo.get(ctx, userID)
	if err != nil {
		return err
	}
	current := ""
	if len(stored.Token) > 0 {
		current, _ = m.seal.open(userID, stored.Token, stored.Nonce)
	}
	same := current == account.AccessToken && stored.User == account.UserID && stored.Homeserver == account.Homeserver && stored.Owner == account.OwnerUserID
	if same && m.running(userID) {
		return nil
	}
	bot := stored
	if !same {
		sealed, nonce, err := m.seal.seal(userID, account.AccessToken)
		if err != nil {
			return err
		}
		bot = Bot{UserID: userID, Enabled: true, Homeserver: account.Homeserver, User: account.UserID, Owner: account.OwnerUserID, Token: sealed, Nonce: nonce, DeviceID: account.DeviceID}
		if err := m.repo.put(ctx, bot); err != nil {
			return err
		}
	}
	m.halt(userID)
	m.launch(bot)
	return nil
}

// agentAvatar keeps the profile's avatar only when it is an http(s) URL;
// Prisme draws the Facile blob for an empty one.
func agentAvatar(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return ""
	}
	return u.String()
}
