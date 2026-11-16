package ai

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/aiusage"
	"github.com/FacileStudio/Bulle/core/jev"
	"gorm.io/gorm"
)

const (
	maxQuestions = 8
	maxOptions   = 64
	maxState     = 32 << 10
	maxBody      = 128 << 10
)

// Decider is the decision model a decide call is answered by.
type Decider interface {
	Evaluate(ctx context.Context, state any, questions map[string]jev.Question) (jev.Response, error)
}

// Ledger records what each call cost. A nil Ledger records nothing.
type Ledger interface {
	Record(ctx context.Context, userID int64, source, caller string, call aiusage.Call)
}

// Budget refuses a call for a user whose monthly AI budget on the server's
// keys is used up. A nil Budget refuses nobody. Completions and summaries are
// checked by their Completer, which knows whose key pays for them.
type Budget interface {
	Check(ctx context.Context, userID int64) error
}

// Config is what the AI routes need. A nil Decider, Summarizer or Completer
// answers every authenticated call of its route with 503, a nil Clef makes a
// two-judge decide answer "not configured" for clef, and RequireAuth is
// the session and bearer middleware every route runs behind. A nil Cap counts
// completions in memory.
type Config struct {
	Decider     Decider
	Clef        Decider
	Summarizer  Summarizer
	Completer   Completer
	Cap         CompletionCap
	RequireAuth func(http.Handler) http.Handler
	Access      Access
	Ledger      Ledger
	Budget      Budget
	Logger      *slog.Logger
}

// Access answers whether a user's plan covers the AI features: nil when it
// does, the 402 payment_required refusal when it does not. spaceID, when not
// empty, is the Registre shared space the request is for: its plan counts
// too when the user is a member of it.
type Access interface {
	CheckAIIn(ctx context.Context, userID int64, spaceID string) error
}

// Judge names. A decide call asks Jev alone unless Judges says otherwise.
const (
	JudgeJev  = "jev"
	JudgeClef = "clef"
)

// DecideRequest is Jev's own contract: one state and the typed questions asked
// about it. Judges is optional: ["jev"], the default, asks Jev alone, and
// ["jev","clef"] asks both in parallel.
type DecideRequest struct {
	State     json.RawMessage         `json:"state"`
	Questions map[string]jev.Question `json:"questions"`
	Judges    []string                `json:"judges,omitempty"`
}

// Answer is one choice question's result. With two judges, Judges holds each
// judge's own answer and Agree says whether they picked the same option; then
// Choice is set only when they agree, Confidence is the lower of the two (0
// when a judge gave no answer) and Probabilities is empty. A single-judge
// answer carries neither Agree nor Judges.
type Answer struct {
	Choice        string                 `json:"choice"`
	Confidence    float64                `json:"confidence"`
	Probabilities map[string]float64     `json:"probabilities"`
	Agree         *bool                  `json:"agree,omitempty"`
	Judges        map[string]JudgeAnswer `json:"judges,omitempty"`
}

// JudgeAnswer is what one judge answered to one question, or why it did not.
type JudgeAnswer struct {
	Choice        string             `json:"choice"`
	Confidence    float64            `json:"confidence"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Error         string             `json:"error,omitempty"`
}

// JudgeCall is what one judge's call served and billed, or why it failed:
// "not configured" for a judge the server has no key for, "did not answer"
// for one that failed.
type JudgeCall struct {
	Model string    `json:"model,omitempty"`
	Usage jev.Usage `json:"usage"`
	Error string    `json:"error,omitempty"`
}

// DecideResponse carries every answer keyed by its question's name, the model
// that served them and what the call billed. With two judges, Model is Jev's,
// Usage is the sum of both calls and Judges holds each call on its own.
type DecideResponse struct {
	Answers map[string]Answer    `json:"answers"`
	Model   string               `json:"model"`
	Usage   jev.Usage            `json:"usage"`
	Judges  map[string]JudgeCall `json:"judges,omitempty"`
}

// NewDecider builds the jev client the route runs on, or nil when no key is
// configured, so the route answers 503 instead of failing at the vendor.
func NewDecider(cfg jev.Config) Decider {
	client := jev.New(cfg)
	if !client.Configured() {
		return nil
	}
	return client
}

// NewClef builds the clef judge a two-judge decide call asks beside Jev, or
// nil when the server holds no OpenRouter key for it, so the call answers
// "not configured" for clef instead of failing at the vendor.
func NewClef(cfg jev.Config) Decider { return NewDecider(cfg) }

// NewCompletionCap counts each user's completions per UTC day in Postgres.
func NewCompletionCap(db *gorm.DB) CompletionCap {
	return &pgCap{db: db, limit: CompletionsPerDay, now: time.Now}
}
