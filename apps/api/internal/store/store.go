package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNotFound    = errors.New("store: not found")
	ErrSessionBusy = errors.New("store: session already has an active run")
)

const (
	RunRunning     = "running"
	RunFinished    = "finished"
	RunStopped     = "stopped"
	RunError       = "error"
	RunInterrupted = "interrupted"

	ApprovalPending  = "pending"
	ApprovalApproved = "approved"
	ApprovalDenied   = "denied"

	ChannelWeb = "web"
)

type Session struct {
	ID        string
	UserID    int64
	Channel   string
	Title     string
	Backend   string
	Model     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Message struct {
	Seq     int
	Payload json.RawMessage
}

type Run struct {
	ID         string
	SessionID  string
	Status     string
	StopReason string
	Usage      json.RawMessage
	CostUSD    float64
	StartedAt  time.Time
	FinishedAt *time.Time
}

type Event struct {
	RunID   string
	Seq     int
	Type    string
	Payload json.RawMessage
}

type Store interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
	ListSessions(ctx context.Context, userID int64) ([]Session, error)
	GetSession(ctx context.Context, userID int64, id string) (Session, error)
	RenameSession(ctx context.Context, id, title string) error
	DeleteSession(ctx context.Context, userID int64, id string) error

	Messages(ctx context.Context, sessionID string) ([]Message, error)
	ReplaceMessages(ctx context.Context, sessionID string, payloads []json.RawMessage) error

	CreateRun(ctx context.Context, run Run) error
	FinishRun(ctx context.Context, id, status, stop string, usage json.RawMessage, costUSD float64) error
	GetRun(ctx context.Context, userID int64, id string) (Run, error)
	ActiveRun(ctx context.Context, sessionID string) (Run, bool, error)
	LatestRun(ctx context.Context, sessionID string) (Run, bool, error)
	InterruptRunning(ctx context.Context) ([]Run, error)

	AppendEvent(ctx context.Context, event Event) error
	Events(ctx context.Context, runID string, after int) ([]Event, error)
	LastSeq(ctx context.Context, runID string) (int, error)

	CreateApproval(ctx context.Context, runID, callID string) error
	DecideApproval(ctx context.Context, runID, callID, status, decidedBy string) (bool, error)
}

func NewID() string {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		panic(err)
	}
	return hex.EncodeToString(raw)
}

func jsonOr(raw json.RawMessage, fallback string) string {
	if len(raw) == 0 {
		return fallback
	}
	return string(raw)
}
