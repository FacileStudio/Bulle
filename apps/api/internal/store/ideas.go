package store

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"
)

// ErrIdeaBatchRunning answers a new batch for a space that already has one running.
var ErrIdeaBatchRunning = errors.New("store: an ideas batch is already running for this space")

const (
	IdeaNew       = "new"
	IdeaSaved     = "saved"
	IdeaDone      = "done"
	IdeaDismissed = "dismissed"
	IdeaExpired   = "expired"

	FeedbackUp   = "up"
	FeedbackDown = "down"

	BatchDaily  = "daily"
	BatchManual = "manual"

	BatchRunning  = "running"
	BatchFinished = "finished"
	BatchError    = "error"
)

// Server-side limits on idea, settings and goal text, counted in runes.
const (
	IdeaTitleLimit    = 120
	IdeaWhyLimit      = 280
	IdeaPromptLimit   = 2000
	IdeaDraftLimit    = 4000
	IdeaReasonLimit   = 200
	IdeaSourcesLimit  = 5
	IdeaBatchLimit    = 7
	SteeringLimit     = 500
	GoalTitleLimit    = 120
	OpenGoalsLimit    = 20
	DefaultIdeaHour   = 8
	DefaultIdeaZone   = "UTC"
	IdeaLifetime      = 72 * time.Hour
	SuppressionWindow = 14 * 24 * time.Hour
)

// IdeaAppBulle is the app of an idea that belongs to no single suite app.
const IdeaAppBulle = "bulle"

var ideaApps = []string{"courrier", "agenda", "opus", "sablier", "mycelium", "nuage", "glouton", IdeaAppBulle}

var ideaKinds = []string{"mail", "task", "event", "follow_up", "research", "other"}

var goalCategories = []string{"health", "relationships", "finance", "career", "interests", "productivity", "other"}

// IdeaKinds lists the kinds an idea may have.
func IdeaKinds() []string { return slices.Clone(ideaKinds) }

// IdeaApps lists the apps an idea may belong to: the suite apps and Bulle.
func IdeaApps() []string { return slices.Clone(ideaApps) }

// IdeaAppOf answers the app a fingerprint names before its first colon, and
// Bulle when that prefix is no suite app. Migration 00027 backfills the same way.
func IdeaAppOf(fingerprint string) string {
	prefix, _, _ := strings.Cut(fingerprint, ":")
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix == "memory" {
		return "mycelium"
	}
	if slices.Contains(ideaApps, prefix) {
		return prefix
	}
	return IdeaAppBulle
}

// GoalCategories lists the categories a goal may have.
func GoalCategories() []string { return slices.Clone(goalCategories) }

// IdeaSource is one thing an idea came from, such as an email or a task.
type IdeaSource struct {
	Label string `json:"label"`
	Ref   string `json:"ref"`
}

// Idea is one suggestion the generator left for a space: a card with a reason,
// a ready prompt and an optional read-only draft.
type Idea struct {
	ID          string
	UserID      int64
	SpaceID     string
	BatchID     string
	App         string
	Kind        string
	Title       string
	Why         string
	Prompt      string
	Draft       string
	Sources     []IdeaSource
	Fingerprint string
	Status      string
	Feedback    string
	Reason      string
	SessionID   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ExpiresAt   time.Time
}

// IdeaSettings is how and when a space's daily batch runs.
type IdeaSettings struct {
	SpaceID       string
	UserID        int64
	Enabled       bool
	Hour          int
	Timezone      string
	Push          bool
	Steering      string
	LastBatchDate string
	UpdatedAt     time.Time
}

// DefaultIdeaSettings is what a space without a settings row uses.
func DefaultIdeaSettings(userID int64, spaceID string) IdeaSettings {
	return IdeaSettings{SpaceID: spaceID, UserID: userID, Enabled: true, Hour: DefaultIdeaHour, Timezone: DefaultIdeaZone}
}

// IdeaBatch is one generator run for a space. App is empty for a batch over
// every app, and names the one app a scoped batch reads and proposes for.
type IdeaBatch struct {
	ID         string
	UserID     int64
	SpaceID    string
	App        string
	RunID      string
	Trigger    string
	Status     string
	CreatedAt  time.Time
	FinishedAt *time.Time
}

// Goal is something the user is working towards in a space; the generator
// reads the open ones.
type Goal struct {
	ID        string
	UserID    int64
	SpaceID   string
	Title     string
	Category  string
	Done      bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Ideas stores ideas, their batches, the per-space settings and goals. Every
// lookup by id is scoped to the user and answers ErrNotFound for anyone else.
type Ideas interface {
	IdeaCards
	IdeaHistory
	IdeaBatches
	IdeaBatchHistory
	IdeaSettingsStore
	Goals
}

// IdeaCards reads and changes the cards themselves. ListActiveIdeas answers
// every app's ideas when app is empty.
type IdeaCards interface {
	ListActiveIdeas(ctx context.Context, userID int64, spaceID, app string, now time.Time) ([]Idea, error)
	GetIdea(ctx context.Context, userID int64, id string) (Idea, error)
	CreateIdea(ctx context.Context, idea Idea) (Idea, error)
	UpdateIdea(ctx context.Context, idea Idea) (Idea, error)
}

// IdeaHistory expires old cards and reads what the user thought of past ones.
// ExpireIdeas with an app expires that app's new ideas; with none, the new
// ideas of every batch that was not scoped to one app. IdeaFeedback reads
// every app when app is empty.
type IdeaHistory interface {
	ExpireIdeas(ctx context.Context, userID int64, spaceID, app, keepBatchID string) error
	ExpireStaleIdeas(ctx context.Context, now time.Time) error
	IdeaFeedback(ctx context.Context, userID int64, spaceID, app string, limit int) ([]Idea, error)
	SuppressedFingerprints(ctx context.Context, userID int64, spaceID string, now time.Time) ([]string, error)
}

// IdeaBatches tracks generator runs. CreateIdeaBatch answers
// ErrIdeaBatchRunning while the space has a running batch for the same app,
// RunningIdeaBatch matches the app exactly, and InterruptIdeaBatches marks
// every running batch as failed at boot.
type IdeaBatches interface {
	CreateIdeaBatch(ctx context.Context, batch IdeaBatch) (IdeaBatch, error)
	UpdateIdeaBatch(ctx context.Context, batch IdeaBatch) error
	RunningIdeaBatch(ctx context.Context, userID int64, spaceID, app string) (IdeaBatch, bool, error)
	InterruptIdeaBatches(ctx context.Context) error
}

// IdeaBatchHistory reads past batches of one app scope: the latest finished
// one and how many of one trigger started since a time. CountIdeaBatches
// counts across all the user's spaces when spaceID is empty.
type IdeaBatchHistory interface {
	LastIdeaBatch(ctx context.Context, userID int64, spaceID, app string) (IdeaBatch, bool, error)
	CountIdeaBatches(ctx context.Context, userID int64, spaceID, app, trigger string, since time.Time) (int, error)
}

// IdeaSettingsStore reads and writes per-space settings. GetIdeaSettings
// answers ErrNotFound when the space has no row.
type IdeaSettingsStore interface {
	GetIdeaSettings(ctx context.Context, userID int64, spaceID string) (IdeaSettings, error)
	SaveIdeaSettings(ctx context.Context, settings IdeaSettings) (IdeaSettings, error)
	ListEnabledIdeaSettings(ctx context.Context) ([]IdeaSettings, error)
	MarkIdeaBatchDate(ctx context.Context, userID int64, spaceID, date string) (bool, error)
	ReleaseIdeaBatchDate(ctx context.Context, userID int64, spaceID, date string) error
}

// Goals stores a space's goals, open ones first.
type Goals interface {
	ListGoals(ctx context.Context, userID int64, spaceID string) ([]Goal, error)
	GetGoal(ctx context.Context, userID int64, id string) (Goal, error)
	CreateGoal(ctx context.Context, goal Goal) (Goal, error)
	UpdateGoal(ctx context.Context, goal Goal) (Goal, error)
	DeleteGoal(ctx context.Context, userID int64, id string) error
}

func activeIdea(idea Idea, now time.Time) bool {
	return idea.Status == IdeaSaved || (idea.Status == IdeaNew && idea.ExpiresAt.After(now))
}

func suppressed(idea Idea, now time.Time) bool {
	if activeIdea(idea, now) {
		return true
	}
	recent := !idea.UpdatedAt.Before(now.Add(-SuppressionWindow))
	return recent && (idea.Status == IdeaDismissed || idea.Status == IdeaDone)
}
