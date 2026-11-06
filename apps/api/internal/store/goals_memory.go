package store

import (
	"cmp"
	"context"
	"slices"
	"time"
)

func (m *Memory) ListGoals(_ context.Context, userID int64, spaceID string) ([]Goal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Goal
	for _, goal := range m.goals {
		if goal.UserID == userID && goal.SpaceID == spaceID {
			out = append(out, goal)
		}
	}
	slices.SortFunc(out, compareGoals)
	return out, nil
}

func (m *Memory) GetGoal(_ context.Context, userID int64, id string) (Goal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	goal, ok := m.goals[id]
	if !ok || goal.UserID != userID {
		return Goal{}, ErrNotFound
	}
	return goal, nil
}

func (m *Memory) CreateGoal(_ context.Context, goal Goal) (Goal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if space, ok := m.spaces[goal.SpaceID]; !ok || space.UserID != goal.UserID {
		return Goal{}, ErrNotFound
	}
	if goal.ID == "" {
		goal.ID = NewID()
	}
	now := time.Now().UTC()
	if !now.After(m.lastSpaceAt) {
		now = m.lastSpaceAt.Add(time.Nanosecond)
	}
	m.lastSpaceAt = now
	goal.CreatedAt, goal.UpdatedAt = now, now
	m.goals[goal.ID] = goal
	return goal, nil
}

func (m *Memory) UpdateGoal(_ context.Context, goal Goal) (Goal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.goals[goal.ID]
	if !ok || existing.UserID != goal.UserID {
		return Goal{}, ErrNotFound
	}
	existing.Title, existing.Category, existing.Done = goal.Title, goal.Category, goal.Done
	existing.UpdatedAt = time.Now().UTC()
	m.goals[goal.ID] = existing
	return existing, nil
}

func (m *Memory) DeleteGoal(_ context.Context, userID int64, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	goal, ok := m.goals[id]
	if !ok || goal.UserID != userID {
		return ErrNotFound
	}
	delete(m.goals, id)
	return nil
}

func compareGoals(a, b Goal) int {
	if a.Done != b.Done {
		if a.Done {
			return 1
		}
		return -1
	}
	return cmp.Or(a.CreatedAt.Compare(b.CreatedAt), cmp.Compare(a.ID, b.ID))
}
