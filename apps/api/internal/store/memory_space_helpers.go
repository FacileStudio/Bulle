package store

import (
	"cmp"
	"slices"
	"time"
)

func (m *Memory) defaultSpace(userID int64) Space {
	if spaces := m.userSpaces(userID); len(spaces) > 0 {
		return spaces[0]
	}
	return m.insertSpace(Space{UserID: userID, Name: DefaultSpaceName})
}

func (m *Memory) insertSpace(space Space) Space {
	if space.ID == "" {
		space.ID = NewID()
	}
	now := time.Now().UTC()
	if !now.After(m.lastSpaceAt) {
		now = m.lastSpaceAt.Add(time.Nanosecond)
	}
	m.lastSpaceAt = now
	space.CreatedAt, space.UpdatedAt = now, now
	m.spaces[space.ID] = space
	return space
}

func (m *Memory) userSpaces(userID int64) []Space {
	var out []Space
	for _, space := range m.spaces {
		if space.UserID == userID {
			out = append(out, space)
		}
	}
	slices.SortFunc(out, func(a, b Space) int {
		return cmp.Or(a.CreatedAt.Compare(b.CreatedAt), cmp.Compare(a.ID, b.ID))
	})
	return out
}

func (m *Memory) ownedSpace(userID int64, spaceID string) (string, error) {
	if spaceID == "" {
		return m.defaultSpace(userID).ID, nil
	}
	if space, ok := m.spaces[spaceID]; !ok || space.UserID != userID {
		return "", ErrNotFound
	}
	return spaceID, nil
}

func (m *Memory) dropSession(id string) {
	delete(m.sessions, id)
	delete(m.messages, id)
	for runID, run := range m.runs {
		if run.SessionID == id {
			delete(m.runs, runID)
			delete(m.events, runID)
		}
	}
}

func (m *Memory) dropJob(id string) {
	delete(m.jobs, id)
	for runID, run := range m.jobRuns {
		if run.JobID == id {
			delete(m.jobRuns, runID)
		}
	}
}
