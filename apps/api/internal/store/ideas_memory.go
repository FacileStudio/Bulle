package store

import (
	"cmp"
	"context"
	"slices"
	"time"
)

func (m *Memory) ListActiveIdeas(_ context.Context, userID int64, spaceID string, now time.Time) ([]Idea, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := m.spaceIdeas(userID, spaceID, func(idea Idea) bool { return activeIdea(idea, now) })
	slices.SortFunc(out, func(a, b Idea) int {
		return cmp.Or(b.CreatedAt.Compare(a.CreatedAt), cmp.Compare(a.ID, b.ID))
	})
	return out, nil
}

func (m *Memory) GetIdea(_ context.Context, userID int64, id string) (Idea, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	idea, ok := m.ideas[id]
	if !ok || idea.UserID != userID {
		return Idea{}, ErrNotFound
	}
	return cloneIdea(idea), nil
}

func (m *Memory) CreateIdea(_ context.Context, idea Idea) (Idea, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if space, ok := m.spaces[idea.SpaceID]; !ok || space.UserID != idea.UserID {
		return Idea{}, ErrNotFound
	}
	if idea.ID == "" {
		idea.ID = NewID()
	}
	if idea.Status == "" {
		idea.Status = IdeaNew
	}
	now := time.Now().UTC()
	idea.CreatedAt, idea.UpdatedAt = now, now
	idea = cloneIdea(idea)
	m.ideas[idea.ID] = idea
	return cloneIdea(idea), nil
}

func (m *Memory) UpdateIdea(_ context.Context, idea Idea) (Idea, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.ideas[idea.ID]
	if !ok || existing.UserID != idea.UserID {
		return Idea{}, ErrNotFound
	}
	existing.Status, existing.Feedback, existing.Reason, existing.SessionID = idea.Status, idea.Feedback, idea.Reason, idea.SessionID
	existing.UpdatedAt = time.Now().UTC()
	m.ideas[idea.ID] = existing
	return cloneIdea(existing), nil
}

func (m *Memory) ExpireIdeas(_ context.Context, userID int64, spaceID, keepBatchID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, idea := range m.ideas {
		if idea.UserID == userID && idea.SpaceID == spaceID && idea.Status == IdeaNew && idea.BatchID != keepBatchID {
			idea.Status, idea.UpdatedAt = IdeaExpired, time.Now().UTC()
			m.ideas[id] = idea
		}
	}
	return nil
}

func (m *Memory) ExpireStaleIdeas(_ context.Context, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, idea := range m.ideas {
		if idea.Status == IdeaNew && !idea.ExpiresAt.After(now) {
			idea.Status, idea.UpdatedAt = IdeaExpired, time.Now().UTC()
			m.ideas[id] = idea
		}
	}
	return nil
}

func (m *Memory) IdeaFeedback(_ context.Context, userID int64, spaceID string, limit int) ([]Idea, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := m.spaceIdeas(userID, spaceID, func(idea Idea) bool {
		return idea.Feedback != "" || idea.Status == IdeaDismissed
	})
	slices.SortFunc(out, func(a, b Idea) int {
		return cmp.Or(b.UpdatedAt.Compare(a.UpdatedAt), cmp.Compare(a.ID, b.ID))
	})
	return out[:min(len(out), limit)], nil
}

func (m *Memory) SuppressedFingerprints(_ context.Context, userID int64, spaceID string, now time.Time) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []string
	for _, idea := range m.spaceIdeas(userID, spaceID, func(idea Idea) bool { return suppressed(idea, now) }) {
		if !slices.Contains(out, idea.Fingerprint) {
			out = append(out, idea.Fingerprint)
		}
	}
	slices.Sort(out)
	return out, nil
}

func (m *Memory) spaceIdeas(userID int64, spaceID string, keep func(Idea) bool) []Idea {
	var out []Idea
	for _, idea := range m.ideas {
		if idea.UserID == userID && idea.SpaceID == spaceID && keep(idea) {
			out = append(out, cloneIdea(idea))
		}
	}
	return out
}

func (m *Memory) dropIdeas(spaceID string) {
	for id, idea := range m.ideas {
		if idea.SpaceID == spaceID {
			delete(m.ideas, id)
		}
	}
	for id, batch := range m.ideaBatches {
		if batch.SpaceID == spaceID {
			delete(m.ideaBatches, id)
		}
	}
	for id, goal := range m.goals {
		if goal.SpaceID == spaceID {
			delete(m.goals, id)
		}
	}
	delete(m.ideaSettings, spaceID)
}

func cloneIdea(idea Idea) Idea {
	idea.Sources = slices.Clone(idea.Sources)
	if idea.Sources == nil {
		idea.Sources = []IdeaSource{}
	}
	return idea
}
