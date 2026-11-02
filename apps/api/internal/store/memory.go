package store

import (
	"cmp"
	"context"
	"encoding/json"
	"slices"
	"sync"
	"time"
)

type Memory struct {
	mu        sync.Mutex
	sessions  map[string]Session
	messages  map[string][]json.RawMessage
	runs      map[string]Run
	events    map[string][]Event
	approvals map[string]string
	profiles  map[int64]AssistantProfile
}

func NewMemory() *Memory {
	return &Memory{
		sessions:  map[string]Session{},
		messages:  map[string][]json.RawMessage{},
		runs:      map[string]Run{},
		events:    map[string][]Event{},
		approvals: map[string]string{},
		profiles:  map[int64]AssistantProfile{},
	}
}

func (m *Memory) CreateSession(_ context.Context, session Session) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	session.CreatedAt, session.UpdatedAt = now, now
	m.sessions[session.ID] = session
	return session, nil
}

func (m *Memory) ListSessions(_ context.Context, userID int64) ([]Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Session
	for _, session := range m.sessions {
		if session.UserID == userID {
			out = append(out, session)
		}
	}
	slices.SortFunc(out, func(a, b Session) int {
		return cmp.Or(b.UpdatedAt.Compare(a.UpdatedAt), cmp.Compare(a.ID, b.ID))
	})
	return out, nil
}

func (m *Memory) GetSession(_ context.Context, userID int64, id string) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[id]
	if !ok || session.UserID != userID {
		return Session{}, ErrNotFound
	}
	return session, nil
}

func (m *Memory) RenameSession(_ context.Context, id, title string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if session, ok := m.sessions[id]; ok {
		session.Title, session.UpdatedAt = title, time.Now().UTC()
		m.sessions[id] = session
	}
	return nil
}

func (m *Memory) DeleteSession(_ context.Context, userID int64, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[id]
	if !ok || session.UserID != userID {
		return ErrNotFound
	}
	delete(m.sessions, id)
	delete(m.messages, id)
	for runID, run := range m.runs {
		if run.SessionID == id {
			delete(m.runs, runID)
			delete(m.events, runID)
		}
	}
	return nil
}

func (m *Memory) Messages(_ context.Context, sessionID string) ([]Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	payloads := m.messages[sessionID]
	out := make([]Message, len(payloads))
	for i, payload := range payloads {
		out[i] = Message{Seq: i, Payload: payload}
	}
	return out, nil
}

func (m *Memory) ReplaceMessages(_ context.Context, sessionID string, payloads []json.RawMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages[sessionID] = slices.Clone(payloads)
	if session, ok := m.sessions[sessionID]; ok {
		session.UpdatedAt = time.Now().UTC()
		m.sessions[sessionID] = session
	}
	return nil
}

func (m *Memory) CreateRun(_ context.Context, run Run) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, other := range m.runs {
		if other.SessionID == run.SessionID && other.Status == RunRunning {
			return ErrSessionBusy
		}
	}
	run.Status = RunRunning
	run.StartedAt = time.Now().UTC()
	m.runs[run.ID] = run
	return nil
}

func (m *Memory) FinishRun(_ context.Context, id, status, stop string, usage json.RawMessage, costUSD float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	run, ok := m.runs[id]
	if !ok {
		return ErrNotFound
	}
	now := time.Now().UTC()
	run.Status, run.StopReason, run.Usage, run.CostUSD, run.FinishedAt = status, stop, usage, costUSD, &now
	m.runs[id] = run
	return nil
}

func (m *Memory) GetRun(_ context.Context, userID int64, id string) (Run, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	run, ok := m.runs[id]
	if !ok || m.sessions[run.SessionID].UserID != userID {
		return Run{}, ErrNotFound
	}
	return run, nil
}

func (m *Memory) ActiveRun(_ context.Context, sessionID string) (Run, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, run := range m.runs {
		if run.SessionID == sessionID && run.Status == RunRunning {
			return run, true, nil
		}
	}
	return Run{}, false, nil
}

func (m *Memory) LatestRun(_ context.Context, sessionID string) (Run, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var latest Run
	var found bool
	for _, run := range m.runs {
		if run.SessionID != sessionID {
			continue
		}
		if !found || run.StartedAt.After(latest.StartedAt) {
			latest = run
			found = true
		}
	}
	return latest, found, nil
}

func (m *Memory) InterruptRunning(_ context.Context) ([]Run, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Run
	now := time.Now().UTC()
	for id, run := range m.runs {
		if run.Status == RunRunning {
			run.Status, run.StopReason, run.FinishedAt = RunInterrupted, RunInterrupted, &now
			m.runs[id] = run
			out = append(out, run)
		}
	}
	return out, nil
}

func (m *Memory) AppendEvent(_ context.Context, event Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[event.RunID] = append(m.events[event.RunID], event)
	return nil
}

func (m *Memory) Events(_ context.Context, runID string, after int) ([]Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Event
	for _, event := range m.events[runID] {
		if event.Seq > after {
			out = append(out, event)
		}
	}
	return out, nil
}

func (m *Memory) LastSeq(_ context.Context, runID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	events := m.events[runID]
	if len(events) == 0 {
		return 0, nil
	}
	return events[len(events)-1].Seq, nil
}

func (m *Memory) CreateApproval(_ context.Context, runID, callID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := runID + "/" + callID
	if _, ok := m.approvals[key]; !ok {
		m.approvals[key] = ApprovalPending
	}
	return nil
}

func (m *Memory) DecideApproval(_ context.Context, runID, callID, status, _ string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := runID + "/" + callID
	if m.approvals[key] != ApprovalPending {
		return false, nil
	}
	m.approvals[key] = status
	return true, nil
}

func (m *Memory) GetAssistantProfile(_ context.Context, userID int64) (AssistantProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	profile, ok := m.profiles[userID]
	if !ok {
		return AssistantProfile{}, ErrNotFound
	}
	if profile.Traits == nil {
		profile.Traits = []string{}
	} else {
		traitsCopy := make([]string, len(profile.Traits))
		copy(traitsCopy, profile.Traits)
		profile.Traits = traitsCopy
	}
	return profile, nil
}

func (m *Memory) SaveAssistantProfile(_ context.Context, profile AssistantProfile) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().UTC()
	existing, ok := m.profiles[profile.UserID]
	if !ok {
		profile.CreatedAt = now
	} else {
		profile.CreatedAt = existing.CreatedAt
	}
	profile.UpdatedAt = now
	if profile.Traits == nil {
		profile.Traits = []string{}
	} else {
		traitsCopy := make([]string, len(profile.Traits))
		copy(traitsCopy, profile.Traits)
		profile.Traits = traitsCopy
	}
	m.profiles[profile.UserID] = profile
	return nil
}
