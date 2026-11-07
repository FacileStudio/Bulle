// Package store persists sessions, messages, jobs, spaces, attachments, ideas and
// preferences, in Postgres or in memory for tests.
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
	ErrLastSpace   = errors.New("store: the last space cannot be deleted")
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

	ChannelWeb    = "web"
	ChannelCron   = "cron"
	ChannelMatrix = "matrix"
	ChannelGuest  = "guest"
	ChannelIdeas  = "ideas"
	ChannelMemory = "memory"
)

// Hidden reports whether sessions of this channel are background work that
// never shows in the session list or the live sidebar: an ideas batch or a
// memory pass.
func Hidden(channel string) bool {
	return channel == ChannelIdeas || channel == ChannelMemory
}

type Session struct {
	ID        string
	UserID    int64
	SpaceID   string
	Channel   string
	Title     string
	Backend   string
	Model     string
	CreatedAt time.Time
	UpdatedAt time.Time

	MemorySeenSeq int
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

type AssistantProfile struct {
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	Archetype string    `json:"archetype"`
	Traits    []string  `json:"traits"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Preferences is what a user chose for themselves where the instance leaves a
// choice: the model a new session starts on and whether their computer may
// reach the network. Empty Model means the instance default; empty
// ComputerNetwork means open.
type Preferences struct {
	UserID          int64     `json:"user_id"`
	Model           string    `json:"model"`
	ComputerNetwork string    `json:"computer_network"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Job struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	SpaceID   string    `json:"space_id"`
	Name      string    `json:"name"`
	Schedule  string    `json:"schedule"`
	Prompt    string    `json:"prompt"`
	Profile   string    `json:"profile"`
	Delivery  string    `json:"delivery"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type JobRun struct {
	ID         string     `json:"id"`
	JobID      string     `json:"job_id"`
	RunID      string     `json:"run_id"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

type Store interface {
	Spaces

	CreateSession(ctx context.Context, session Session) (Session, error)
	ListSessions(ctx context.Context, userID int64, spaceID string) ([]Session, error)
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

	GetAssistantProfile(ctx context.Context, userID int64) (AssistantProfile, error)
	SaveAssistantProfile(ctx context.Context, profile AssistantProfile) error

	GetPreferences(ctx context.Context, userID int64) (Preferences, error)
	SavePreferences(ctx context.Context, prefs Preferences) error

	CreateJob(ctx context.Context, job Job) (Job, error)
	GetJob(ctx context.Context, userID int64, id string) (Job, error)
	ListJobs(ctx context.Context, userID int64, spaceID string) ([]Job, error)
	UpdateJob(ctx context.Context, job Job) error
	DeleteJob(ctx context.Context, userID int64, id string) error
	ListDueJobs(ctx context.Context) ([]Job, error)

	CreateJobRun(ctx context.Context, run JobRun) error
	UpdateJobRunStatus(ctx context.Context, id, status string, finishedAt *time.Time) error
	ListJobRuns(ctx context.Context, jobID string) ([]JobRun, error)
	JobActivity(ctx context.Context, userID int64) (map[string]JobActivity, error)

	UsageRuns(ctx context.Context, userID int64, since time.Time) ([]UsageRun, error)

	Attachments
	MyceliumSpaces
	SearchQuota
	MailSeen
	ComputerNetworks
	Ideas
	Chats
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
