// Package live names the events Bulle sends on a user's live stream, so
// every open app sees the same sessions, runs and approvals.
package live

import (
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

const (
	SessionCreated  = "session.created"
	SessionUpdated  = "session.updated"
	SessionDeleted  = "session.deleted"
	SessionRead     = "session.read"
	RunStarted      = "run.started"
	RunFinished     = "run.finished"
	ApprovalPending = "approval.pending"
	ApprovalDecided = "approval.decided"
)

const (
	StatusIdle    = "idle"
	StatusRunning = "running"
	StatusWaiting = "waiting"
)

// Publisher sends one event to one user's open apps and never blocks.
type Publisher interface {
	Publish(userID int64, eventType string, data any)
}

// Session is the payload of session.created and session.updated. Status is
// idle, running or waiting (a tool call waits for approval).
type Session struct {
	SessionID string    `json:"session_id"`
	SpaceID   string    `json:"space_id"`
	Title     string    `json:"title"`
	Channel   string    `json:"channel,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
	Status    string    `json:"status"`
	Unread    bool      `json:"unread"`
}

// Deleted is the payload of session.deleted and session.read.
type Deleted struct {
	SessionID string `json:"session_id"`
}

// Run is the payload of run.started and run.finished. Status is the run's
// outcome on run.finished; SessionStatus is the session's status after the
// event.
type Run struct {
	SessionID     string `json:"session_id"`
	RunID         string `json:"run_id"`
	Channel       string `json:"channel,omitempty"`
	Status        string `json:"status,omitempty"`
	SessionStatus string `json:"session_status"`
}

// Approval is the payload of approval.pending and approval.decided.
type Approval struct {
	SessionID     string `json:"session_id"`
	RunID         string `json:"run_id"`
	CallID        string `json:"call_id"`
	Tool          string `json:"tool,omitempty"`
	Approved      *bool  `json:"approved,omitempty"`
	SessionStatus string `json:"session_status"`
}

// SessionOf is the session.created or session.updated payload for s.
func SessionOf(s store.Session, status string) Session {
	if status == "" {
		status = StatusIdle
	}
	return Session{
		SessionID: s.ID, SpaceID: s.SpaceID, Title: s.Title, Channel: s.Channel,
		UpdatedAt: s.UpdatedAt.UTC(), Status: status, Unread: s.Unread(),
	}
}
