package matrixbot

import (
	"context"
	"errors"
	"github.com/FacileStudio/Bulle/apps/api/internal/plans"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"strings"
	"time"

	"github.com/FacileStudio/nacelle"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/core/chat"
)

const runWait = 15 * time.Minute

// withheld replaces a reply the disclosure screen held back.
const withheld = "I can't share that here. Ask me in a direct message."

// roster lists who can read a room.
type roster interface {
	Members(ctx context.Context, room string) ([]string, error)
}

// responder answers a mention in a linked room, and every message in a direct
// message. The sender's clearance is settled first, from the homeserver's own
// facts: the owner gets their full agent, anyone else a guest session with no
// tools. A reply bound for a room that people other than the owner can read
// passes the disclosure screen before it is posted.
func (m *Manager) responder(bot Bot, rooms roster) chat.Responder {
	return func(ctx context.Context, msg chat.Message) (string, error) {
		if !msg.Is(chat.KindText) {
			return "", nil
		}
		members, err := rooms.Members(ctx, msg.Room)
		if err != nil {
			return "", err
		}
		access := assess(msg.Sender, bot.Owner, bot.User, members)
		if !msg.Mentioned && !access.Direct {
			return "", nil
		}
		space, err := m.space(ctx, bot.UserID, msg.Room, access.Direct)
		if err != nil || space == "" {
			return "", err
		}
		if !access.Owner && !m.guests.allow(bot.UserID, msg.Sender, time.Now()) {
			return slowDown, nil
		}
		reply, err := m.run(ctx, bot.UserID, space, msg, access.Owner)
		if err != nil || access.Private || m.cfg.Screen.Allows(ctx, reply) {
			return reply, err
		}
		return withheld, nil
	}
}

// space is the Bulle space a room's conversation lives in: the linked space, or
// the default one for a direct message, which is linked to none.
func (m *Manager) space(ctx context.Context, userID int64, room string, direct bool) (string, error) {
	space, err := m.repo.roomSpace(ctx, userID, room)
	if err != nil || space != "" || !direct {
		return space, err
	}
	fallback, err := m.cfg.Store.EnsureDefaultSpace(ctx, userID)
	return fallback.ID, err
}

func (m *Manager) run(ctx context.Context, userID int64, space string, msg chat.Message, owner bool) (string, error) {
	sessionID, err := m.session(ctx, userID, space, msg, !owner)
	if err != nil {
		return "", err
	}
	runID, err := m.cfg.Runner.Start(ctx, userID, sessionID, prompt(msg, owner))
	if errors.Is(err, plans.ErrAIAccess) {
		return refusal(err), nil
	}
	if err != nil {
		return "", err
	}
	if err := m.wait(ctx, userID, runID); err != nil {
		return "", err
	}
	return m.answer(ctx, sessionID)
}

func prompt(msg chat.Message, owner bool) string {
	if owner {
		return "From " + msg.Sender + " (your owner) in Matrix:\n\n" + msg.Text
	}
	return "From " + msg.Sender + ", a colleague of your owner, in Matrix:\n\n" + msg.Text
}

// session is the thread's conversation, created on the first message. The
// owner and the guests of one thread get separate sessions, so a guest's
// session never carries what the owner's tools returned.
func (m *Manager) session(ctx context.Context, userID int64, space string, msg chat.Message, guest bool) (string, error) {
	id, err := m.repo.threadSession(ctx, userID, msg.Room, msg.Thread, guest)
	if err != nil {
		return "", err
	}
	if id != "" {
		if _, err := m.cfg.Store.GetSession(ctx, userID, id); err == nil {
			return id, nil
		}
	}
	channel, title := store.ChannelMatrix, "Matrix thread"
	if guest {
		channel, title = store.ChannelGuest, "Matrix guest thread"
	}
	session, err := m.cfg.Store.CreateSession(ctx, store.Session{
		ID: store.NewID(), UserID: userID, SpaceID: space, Channel: channel, Title: title,
	})
	if err != nil {
		return "", err
	}
	return session.ID, m.repo.setThreadSession(ctx, userID, msg.Room, msg.Thread, guest, session.ID)
}

// wait blocks until the run ends, as the scheduler does for a job.
func (m *Manager) wait(ctx context.Context, userID int64, runID string) error {
	changed, unsubscribe := m.cfg.Runner.Subscribe(runID)
	defer unsubscribe()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	deadline := time.After(runWait)
	for {
		run, err := m.cfg.Store.GetRun(ctx, userID, runID)
		if err == nil && run.Status != store.RunRunning {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return errors.New("still working after 15 minutes; follow it in the Bulle web app")
		case <-changed:
		case <-ticker.C:
		}
	}
}

// answer is the text the run wrote after the question, never an earlier turn's.
func (m *Manager) answer(ctx context.Context, sessionID string) (string, error) {
	rows, err := m.cfg.Store.Messages(ctx, sessionID)
	if err != nil {
		return "", err
	}
	var parts []string
	for i := len(rows) - 1; i >= 0; i-- {
		message, err := runner.DecodeMessage(rows[i].Payload)
		if err != nil {
			continue
		}
		text := textOf(message)
		if message.Role == nacelle.RoleUser && text != "" {
			break
		}
		if message.Role == nacelle.RoleAssistant && text != "" {
			parts = append([]string{text}, parts...)
		}
	}
	if len(parts) == 0 {
		return "", errors.New("the run ended without an answer")
	}
	return parts[len(parts)-1], nil
}

func textOf(message nacelle.Message) string {
	var text strings.Builder
	for _, part := range message.Parts {
		if t, ok := part.(nacelle.Text); ok {
			text.WriteString(t.Text)
		}
	}
	return text.String()
}

func refusal(err error) string {
	var denied *troncerrors.Error
	if !errors.As(err, &denied) {
		return err.Error()
	}
	if denied.UpgradeURL == "" {
		return denied.Message
	}
	return denied.Message + " " + denied.UpgradeURL
}
