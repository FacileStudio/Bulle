package ideas

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// IdeaResponse is one idea card.
type IdeaResponse struct {
	ID        string             `json:"id"`
	SpaceID   string             `json:"space_id"`
	BatchID   string             `json:"batch_id"`
	App       string             `json:"app"`
	Kind      string             `json:"kind"`
	Title     string             `json:"title"`
	Why       string             `json:"why"`
	Prompt    string             `json:"prompt"`
	Draft     string             `json:"draft"`
	Sources   []store.IdeaSource `json:"sources"`
	Status    string             `json:"status"`
	Feedback  string             `json:"feedback"`
	SessionID string             `json:"session_id"`
	CreatedAt time.Time          `json:"created_at"`
	ExpiresAt time.Time          `json:"expires_at"`
}

// GoalResponse is one goal of a space.
type GoalResponse struct {
	ID        string    `json:"id"`
	SpaceID   string    `json:"space_id"`
	Title     string    `json:"title"`
	Category  string    `json:"category"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

// SettingsResponse is how a space's daily batch runs.
type SettingsResponse struct {
	SpaceID  string `json:"space_id"`
	Enabled  bool   `json:"enabled"`
	Hour     int    `json:"hour"`
	Timezone string `json:"timezone"`
	Push     bool   `json:"push"`
	Steering string `json:"steering"`
}

// ListResponse is the whole Ideas page of one space.
type ListResponse struct {
	Ideas       []IdeaResponse   `json:"ideas"`
	Goals       []GoalResponse   `json:"goals"`
	Settings    SettingsResponse `json:"settings"`
	Generating  bool             `json:"generating"`
	LastBatchAt *time.Time       `json:"last_batch_at"`
}

// UpdateIdeaRequest changes only the fields it carries.
type UpdateIdeaRequest struct {
	Status   *string `json:"status,omitempty"`
	Feedback *string `json:"feedback,omitempty"`
	Reason   *string `json:"reason,omitempty"`
}

// StartResponse names the chat that does the idea.
type StartResponse struct {
	SessionID string `json:"session_id"`
}

// RunStartResponse names the background run doing an idea and the chat it
// runs in, which the user can continue in Bulle at /s/{session_id}.
type RunStartResponse struct {
	RunID     string `json:"run_id"`
	SessionID string `json:"session_id"`
	IdeaID    string `json:"idea_id"`
}

// RunOutput is one call of a run that changed the user's data.
type RunOutput struct {
	CallID  string `json:"call_id"`
	Tool    string `json:"tool"`
	App     string `json:"app"`
	ID      string `json:"id"`
	URL     string `json:"url"`
	Preview string `json:"preview"`
}

// PendingApproval is a call waiting for the user. Answer it with POST
// decide_path and {"approve": true} or {"approve": false}.
type PendingApproval struct {
	CallID     string          `json:"call_id"`
	Tool       string          `json:"tool"`
	Args       json.RawMessage `json:"args"`
	Grantable  bool            `json:"grantable"`
	DecidePath string          `json:"decide_path"`
}

// RunResponse is a run as an app without the chat UI shows it.
type RunResponse struct {
	ID               string            `json:"id"`
	SessionID        string            `json:"session_id"`
	Status           string            `json:"status"`
	Summary          string            `json:"summary"`
	Outputs          []RunOutput       `json:"outputs"`
	PendingApprovals []PendingApproval `json:"pending_approvals"`
	StartedAt        time.Time         `json:"started_at"`
	FinishedAt       *time.Time        `json:"finished_at"`
	ContinuePath     string            `json:"continue_path"`
}

// GenerateRequest asks for a manual batch in one space.
type GenerateRequest struct {
	SpaceID string `json:"space_id"`
}

// GenerateResponse names the batch that started.
type GenerateResponse struct {
	BatchID string `json:"batch_id"`
}

// SettingsRequest changes only the fields it carries.
type SettingsRequest struct {
	SpaceID  string  `json:"space_id"`
	Enabled  *bool   `json:"enabled,omitempty"`
	Hour     *int    `json:"hour,omitempty"`
	Timezone *string `json:"timezone,omitempty"`
	Push     *bool   `json:"push,omitempty"`
	Steering *string `json:"steering,omitempty"`
}

// CreateGoalRequest adds a goal to a space.
type CreateGoalRequest struct {
	SpaceID  string `json:"space_id"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

// UpdateGoalRequest changes only the fields it carries.
type UpdateGoalRequest struct {
	Title    *string `json:"title,omitempty"`
	Category *string `json:"category,omitempty"`
	Done     *bool   `json:"done,omitempty"`
}

func ideaResponse(idea store.Idea) IdeaResponse {
	sources := idea.Sources
	if sources == nil {
		sources = []store.IdeaSource{}
	}
	return IdeaResponse{
		ID: idea.ID, SpaceID: idea.SpaceID, BatchID: idea.BatchID, App: idea.App, Kind: idea.Kind, Title: idea.Title,
		Why: idea.Why, Prompt: idea.Prompt, Draft: idea.Draft, Sources: sources, Status: idea.Status,
		Feedback: idea.Feedback, SessionID: idea.SessionID, CreatedAt: idea.CreatedAt, ExpiresAt: idea.ExpiresAt,
	}
}

func goalResponse(goal store.Goal) GoalResponse {
	return GoalResponse{
		ID: goal.ID, SpaceID: goal.SpaceID, Title: goal.Title, Category: goal.Category,
		Done: goal.Done, CreatedAt: goal.CreatedAt,
	}
}

func settingsResponse(s store.IdeaSettings) SettingsResponse {
	return SettingsResponse{
		SpaceID: s.SpaceID, Enabled: s.Enabled, Hour: s.Hour, Timezone: s.Timezone, Push: s.Push, Steering: s.Steering,
	}
}

func runResponse(view runner.RunView) RunResponse {
	res := RunResponse{
		ID: view.ID, SessionID: view.SessionID, Status: view.Status, Summary: view.Summary,
		Outputs: make([]RunOutput, 0, len(view.Outputs)), PendingApprovals: make([]PendingApproval, 0, len(view.Pending)),
		StartedAt: view.StartedAt, FinishedAt: view.FinishedAt, ContinuePath: "/s/" + url.PathEscape(view.SessionID),
	}
	for _, out := range view.Outputs {
		res.Outputs = append(res.Outputs, RunOutput{CallID: out.CallID, Tool: out.Tool, App: out.App, ID: out.ID, URL: out.URL, Preview: out.Preview})
	}
	for _, p := range view.Pending {
		res.PendingApprovals = append(res.PendingApprovals, PendingApproval{
			CallID: p.CallID, Tool: p.Tool, Args: p.Args, Grantable: p.Grantable,
			DecidePath: "/api/runs/" + url.PathEscape(view.ID) + "/approvals/" + url.PathEscape(p.CallID),
		})
	}
	return res
}
