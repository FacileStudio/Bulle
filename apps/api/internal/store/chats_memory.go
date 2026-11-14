package store

import (
	"cmp"
	"context"
	"slices"
	"strings"
	"time"
)

func (m *Memory) ListSessionsForMemoryPass(_ context.Context, idleSince time.Time) ([]Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	oldest := idleSince.Add(-MemoryPassLookback)
	var out []Session
	for _, session := range m.sessions {
		if !Conversational(session.Channel) || !session.UpdatedAt.Before(idleSince) || !session.UpdatedAt.After(oldest) {
			continue
		}
		if len(m.messages[session.ID]) > session.MemorySeenSeq {
			out = append(out, session)
		}
	}
	slices.SortFunc(out, func(a, b Session) int {
		return cmp.Or(b.UpdatedAt.Compare(a.UpdatedAt), cmp.Compare(a.ID, b.ID))
	})
	return out[:min(len(out), memoryPassBatch)], nil
}

func (m *Memory) MarkMemorySeen(_ context.Context, sessionID string, seq int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[sessionID]
	if !ok {
		return ErrNotFound
	}
	session.MemorySeenSeq = max(session.MemorySeenSeq, seq)
	m.sessions[sessionID] = session
	return nil
}

func (m *Memory) MarkTainted(_ context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[sessionID]
	if !ok {
		return ErrNotFound
	}
	session.Tainted = true
	m.sessions[sessionID] = session
	return nil
}

func (m *Memory) SearchSessions(_ context.Context, userID int64, spaceID, query string, limit int) ([]SessionHit, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	query = strings.TrimSpace(query)
	var out []SessionHit
	for _, session := range m.sessions {
		if session.UserID != userID || session.SpaceID != spaceID || !Conversational(session.Channel) {
			continue
		}
		snippet, inText := FirstSnippet(m.messages[session.ID], query)
		_, inTitle := Snippet(session.Title, query)
		if inText || inTitle {
			out = append(out, SessionHit{SessionID: session.ID, Title: session.Title, Snippet: snippet, UpdatedAt: session.UpdatedAt})
		}
	}
	slices.SortFunc(out, func(a, b SessionHit) int {
		return cmp.Or(b.UpdatedAt.Compare(a.UpdatedAt), cmp.Compare(a.SessionID, b.SessionID))
	})
	return out[:min(len(out), max(limit, 0))], nil
}
