package store

import (
	"context"
	"errors"
	"slices"
	"time"
)

// CreateAttachment stores the bytes and answers the row without them.
func (m *Memory) CreateAttachment(_ context.Context, a Attachment) (Attachment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.attachments == nil {
		m.attachments = map[string]Attachment{}
	}
	used := a.Size
	for _, kept := range m.attachments {
		if kept.UserID == a.UserID {
			used += kept.Size
		}
	}
	if used > AttachmentQuota {
		return Attachment{}, ErrAttachmentQuota
	}
	a.CreatedAt = time.Now().UTC()
	a.Data = slices.Clone(a.Data)
	m.attachments[a.ID] = a
	a.Data = nil
	return a, nil
}

// GetAttachment answers one of the user's attachments with its bytes.
func (m *Memory) GetAttachment(_ context.Context, userID int64, id string) (Attachment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	found, ok := m.attachments[id]
	if !ok || found.UserID != userID {
		return Attachment{}, ErrNotFound
	}
	found.Data = slices.Clone(found.Data)
	return found, nil
}

// AttachmentInfo answers the metadata of the ids the user owns, in the order
// asked.
func (m *Memory) AttachmentInfo(_ context.Context, userID int64, ids []string) ([]Attachment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	byID := map[string]Attachment{}
	for _, id := range ids {
		if found, ok := m.attachments[id]; ok && found.UserID == userID {
			found.Data = nil
			byID[id] = found
		}
	}
	return inOrder(ids, byID), nil
}

// MyceliumSpace answers the stored space id, or ErrNotFound.
func (m *Memory) MyceliumSpace(_ context.Context, userID int64) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id, ok := m.mycelium[userID]
	if !ok {
		return "", ErrNotFound
	}
	return id, nil
}

// SaveMyceliumSpace records the user's space, replacing any earlier one.
func (m *Memory) SaveMyceliumSpace(_ context.Context, userID int64, spaceID string) error {
	if spaceID == "" {
		return errors.New("store: empty mycelium space id")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mycelium == nil {
		m.mycelium = map[int64]string{}
	}
	m.mycelium[userID] = spaceID
	return nil
}
