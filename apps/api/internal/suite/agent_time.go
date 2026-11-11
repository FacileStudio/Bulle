package suite

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Sablier books the agent's own working time to the user in Sablier. The
// agent's Sablier tools come from Sablier's MCP server; this client only
// records agent sessions and heartbeats, which no MCP tool covers.
type Sablier struct{ gw *Gateway }

func NewSablier(gw *Gateway) Sablier { return Sablier{gw: gw} }

const (
	sablierAgentSessionsPath = "/api/agents/sessions"
	sablierHeartbeatsPath    = "/api/heartbeats"
)

// AgentSession is one span of agent work booked to the user's agent in Sablier.
type AgentSession struct {
	AgentKey      string    `json:"agent_key"`
	Name          string    `json:"name"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	FacileID      string    `json:"facile_id"`
	StartedAt     time.Time `json:"started_at"`
	StoppedAt     time.Time `json:"stopped_at"`
	ActiveSeconds int64     `json:"active_seconds"`
	TokensIn      int64     `json:"tokens_in,omitempty"`
	TokensOut     int64     `json:"tokens_out,omitempty"`
	Cowork        bool      `json:"cowork"`
	Description   string    `json:"description,omitempty"`

	RegistreSpaceID string   `json:"registre_space_id,omitempty"`
	OpusWorkspace   string   `json:"opus_workspace,omitempty"`
	OpusTaskIDs     []string `json:"opus_task_ids,omitempty"`
	Title           string   `json:"title,omitempty"`
}

// Heartbeat is one sign that the user was working at a given moment.
type Heartbeat struct {
	At      time.Time `json:"at"`
	Source  string    `json:"source"`
	Entity  string    `json:"entity"`
	Project string    `json:"project"`

	RegistreSpaceID string `json:"registre_space_id,omitempty"`
	OpusTaskID      string `json:"opus_task_id,omitempty"`
}

// RecordAgentSession upserts the span in Sablier, keyed by its FacileID, with
// the write scope.
func (s Sablier) RecordAgentSession(ctx context.Context, userID int64, session AgentSession) error {
	_, err := s.gw.Do(ctx, userID, Request{App: AppSablier, Method: http.MethodPost, Path: sablierAgentSessionsPath, Body: session})
	return err
}

// RecordHeartbeat posts one heartbeat for the user with the write scope.
func (s Sablier) RecordHeartbeat(ctx context.Context, userID int64, beat Heartbeat) error {
	payload := map[string]any{"heartbeats": []Heartbeat{beat}}
	_, err := s.gw.Do(ctx, userID, Request{App: AppSablier, Method: http.MethodPost, Path: sablierHeartbeatsPath, Body: payload})
	return err
}

// OpusTaskID reads the Opus facile id out of what opus_read_task,
// opus_create_task or opus_update_task returned. Any other tool, a result that is not a task, or
// a task without a facile id answers false.
func OpusTaskID(tool, result string) (string, bool) {
	switch tool {
	case "opus_read_task", "opus_create_task", "opus_update_task":
	default:
		return "", false
	}
	var task struct {
		FacileID string `json:"facile_id"`
	}
	if json.Unmarshal([]byte(result), &task) != nil || task.FacileID == "" {
		return "", false
	}
	return task.FacileID, true
}
