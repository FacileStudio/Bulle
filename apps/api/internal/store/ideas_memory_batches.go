package store

import (
	"cmp"
	"context"
	"slices"
	"time"
)

func (m *Memory) CreateIdeaBatch(_ context.Context, batch IdeaBatch) (IdeaBatch, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if space, ok := m.spaces[batch.SpaceID]; !ok || space.UserID != batch.UserID {
		return IdeaBatch{}, ErrNotFound
	}
	for _, other := range m.ideaBatches {
		if other.SpaceID == batch.SpaceID && other.Status == BatchRunning {
			return IdeaBatch{}, ErrIdeaBatchRunning
		}
	}
	if batch.ID == "" {
		batch.ID = NewID()
	}
	batch.Status = BatchRunning
	batch.CreatedAt = time.Now().UTC()
	m.ideaBatches[batch.ID] = batch
	return batch, nil
}

func (m *Memory) UpdateIdeaBatch(_ context.Context, batch IdeaBatch) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.ideaBatches[batch.ID]
	if !ok || existing.UserID != batch.UserID {
		return ErrNotFound
	}
	existing.RunID, existing.Status, existing.FinishedAt = batch.RunID, batch.Status, batch.FinishedAt
	m.ideaBatches[batch.ID] = existing
	return nil
}

func (m *Memory) RunningIdeaBatch(_ context.Context, userID int64, spaceID string) (IdeaBatch, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, batch := range m.ideaBatches {
		if batch.UserID == userID && batch.SpaceID == spaceID && batch.Status == BatchRunning {
			return batch, true, nil
		}
	}
	return IdeaBatch{}, false, nil
}

func (m *Memory) InterruptIdeaBatches(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	for id, batch := range m.ideaBatches {
		if batch.Status == BatchRunning {
			batch.Status, batch.FinishedAt = BatchError, &now
			m.ideaBatches[id] = batch
		}
	}
	return nil
}

func (m *Memory) LastIdeaBatch(_ context.Context, userID int64, spaceID string) (IdeaBatch, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var last IdeaBatch
	var found bool
	for _, batch := range m.ideaBatches {
		if batch.UserID != userID || batch.SpaceID != spaceID || batch.Status != BatchFinished {
			continue
		}
		if !found || batch.CreatedAt.After(last.CreatedAt) {
			last, found = batch, true
		}
	}
	return last, found, nil
}

func (m *Memory) CountIdeaBatches(_ context.Context, userID int64, spaceID, trigger string, since time.Time) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, batch := range m.ideaBatches {
		if batch.UserID == userID && batch.SpaceID == spaceID && batch.Trigger == trigger && !batch.CreatedAt.Before(since) {
			count++
		}
	}
	return count, nil
}

func (m *Memory) GetIdeaSettings(_ context.Context, userID int64, spaceID string) (IdeaSettings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	settings, ok := m.ideaSettings[spaceID]
	if !ok || settings.UserID != userID {
		return IdeaSettings{}, ErrNotFound
	}
	return settings, nil
}

func (m *Memory) SaveIdeaSettings(_ context.Context, settings IdeaSettings) (IdeaSettings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if space, ok := m.spaces[settings.SpaceID]; !ok || space.UserID != settings.UserID {
		return IdeaSettings{}, ErrNotFound
	}
	if existing, ok := m.ideaSettings[settings.SpaceID]; ok {
		settings.LastBatchDate = existing.LastBatchDate
	}
	settings.UpdatedAt = time.Now().UTC()
	m.ideaSettings[settings.SpaceID] = settings
	return settings, nil
}

func (m *Memory) ListEnabledIdeaSettings(_ context.Context) ([]IdeaSettings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []IdeaSettings
	for _, settings := range m.ideaSettings {
		if settings.Enabled {
			out = append(out, settings)
		}
	}
	slices.SortFunc(out, func(a, b IdeaSettings) int { return cmp.Compare(a.SpaceID, b.SpaceID) })
	return out, nil
}

func (m *Memory) MarkIdeaBatchDate(_ context.Context, userID int64, spaceID, date string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	settings, ok := m.ideaSettings[spaceID]
	if !ok || settings.UserID != userID || settings.LastBatchDate == date {
		return false, nil
	}
	settings.LastBatchDate = date
	m.ideaSettings[spaceID] = settings
	return true, nil
}
