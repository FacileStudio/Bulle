package store

import (
	"context"
	"time"
)

func (m *Memory) ListSpaces(_ context.Context, userID int64) ([]Space, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultSpace(userID)
	return m.userSpaces(userID), nil
}

func (m *Memory) GetSpace(_ context.Context, userID int64, id string) (Space, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	space, ok := m.spaces[id]
	if !ok || space.UserID != userID {
		return Space{}, ErrNotFound
	}
	return space, nil
}

func (m *Memory) CreateSpace(_ context.Context, space Space) (Space, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.insertSpace(space), nil
}

func (m *Memory) UpdateSpace(_ context.Context, space Space) (Space, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.spaces[space.ID]
	if !ok || existing.UserID != space.UserID {
		return Space{}, ErrNotFound
	}
	space.CreatedAt, space.UpdatedAt = existing.CreatedAt, time.Now().UTC()
	m.spaces[space.ID] = space
	return space, nil
}

func (m *Memory) DeleteSpace(_ context.Context, userID int64, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	space, ok := m.spaces[id]
	if !ok || space.UserID != userID {
		return ErrNotFound
	}
	if len(m.userSpaces(userID)) <= 1 {
		return ErrLastSpace
	}
	delete(m.spaces, id)
	for sessionID, session := range m.sessions {
		if session.SpaceID == id {
			m.dropSession(sessionID)
		}
	}
	for jobID, job := range m.jobs {
		if job.SpaceID == id {
			m.dropJob(jobID)
		}
	}
	m.dropIdeas(id)
	return nil
}

func (m *Memory) EnsureDefaultSpace(_ context.Context, userID int64) (Space, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.defaultSpace(userID), nil
}
