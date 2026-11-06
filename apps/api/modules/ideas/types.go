package ideas

import (
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// IdeaResponse is one idea card.
type IdeaResponse struct {
	ID        string             `json:"id"`
	SpaceID   string             `json:"space_id"`
	BatchID   string             `json:"batch_id"`
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
		ID: idea.ID, SpaceID: idea.SpaceID, BatchID: idea.BatchID, Kind: idea.Kind, Title: idea.Title,
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
