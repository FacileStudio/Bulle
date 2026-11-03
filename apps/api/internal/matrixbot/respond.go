package matrixbot

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/FacileStudio/nacelle"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/core/chat"
)

const runWait = 15 * time.Minute

// responder answers a mention in a linked room. Anything else, a message
// that does not mention the bot or a room linked to no space, gets silence.
func (m *Manager) responder(userID int64) chat.Responder {
	return func(ctx context.Context, msg chat.Message) (string, error) {
		if !msg.Mentioned || !msg.Is(chat.KindText) {
			return "", nil
		}
		space, err := m.repo.roomSpace(ctx, userID, msg.Room)
		if err != nil || space == "" {
			return "", err
		}
		sessionID, err := m.session(ctx, userID, space, msg)
		if err != nil {
			return "", err
		}
		runID, err := m.cfg.Runner.Start(ctx, userID, sessionID, "From "+msg.Sender+" in the space's Matrix room:\n\n"+msg.Text)
		if err != nil {
			return "", err
		}
		if err := m.wait(ctx, userID, runID); err != nil {
			return "", err
		}
		return m.answer(ctx, sessionID)
	}
}

// session is the thread's conversation, created on the first mention.
func (m *Manager) session(ctx context.Context, userID int64, space string, msg chat.Message) (string, error) {
	id, err := m.repo.threadSession(ctx, userID, msg.Room, msg.Thread)
	if err != nil {
		return "", err
	}
	if id != "" {
		if _, err := m.cfg.Store.GetSession(ctx, userID, id); err == nil {
			return id, nil
		}
	}
	session, err := m.cfg.Store.CreateSession(ctx, store.Session{
		ID: store.NewID(), UserID: userID, SpaceID: space, Channel: store.ChannelMatrix, Title: "Matrix thread",
	})
	if err != nil {
		return "", err
	}
	return session.ID, m.repo.setThreadSession(ctx, userID, msg.Room, msg.Thread, session.ID)
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
