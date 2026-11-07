package ai

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/core/jev"
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

// Users finds the Bulle account a service acts for.
type Users interface {
	IdentityForEmail(ctx context.Context, email string) (authcontext.Identity, error)
}

// Config is what the decide route needs. A nil Decider answers every
// authenticated call with 503, and Callers maps each calling service's name to
// the SHA-256 of its key.
type Config struct {
	Decider Decider
	Callers map[string][]byte
	Users   Users
	Access  Access
	Logger  *slog.Logger
}

// Access answers whether a user's plan covers the AI features: nil when it
// does, the 402 payment_required refusal when it does not.
type Access interface {
	CheckAI(ctx context.Context, userID int64) error
}

// DecideRequest is Jev's own contract: one state and the typed questions asked
// about it.
type DecideRequest struct {
	State     json.RawMessage         `json:"state"`
	Questions map[string]jev.Question `json:"questions"`
}

// Answer is one choice question's result.
type Answer struct {
	Choice        string             `json:"choice"`
	Confidence    float64            `json:"confidence"`
	Probabilities map[string]float64 `json:"probabilities"`
}

// DecideResponse carries every answer keyed by its question's name, the model
// that served them and what the call billed.
type DecideResponse struct {
	Answers map[string]Answer `json:"answers"`
	Model   string            `json:"model"`
	Usage   jev.Usage         `json:"usage"`
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
