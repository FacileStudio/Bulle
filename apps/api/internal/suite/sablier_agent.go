package suite

import (
	"context"
	"encoding/json"
	"time"
)

// bookkeepingTools are the MCP tools only Bulle itself calls, to book the
// agent's own time to the user in Sablier. They never reach the model's tool
// list or search_tools, so the model cannot book time on its own.
var bookkeepingTools = map[string]bool{
	"sablier_record_agent_session": true,
	"sablier_record_heartbeat":     true,
}

// AgentSession is one span of agent work booked to the user's agent in Sablier.
type AgentSession struct {
	AgentKey      string    `json:"agent_key"`
	Name          string    `json:"name"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	FacileID      string    `json:"facile_id"`
	StartedAt     time.Time `json:"started_at"`
	StoppedAt     time.Time `json:"stopped_at"`
	ActiveSeconds int64     `json:"active_seconds"`
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
	Project string    `json:"project,omitempty"`

	RegistreSpaceID string `json:"registre_space_id,omitempty"`
	OpusTaskID      string `json:"opus_task_id,omitempty"`
}

// RecordAgentSession upserts the span in Sablier, keyed by its FacileID,
// through Sablier's MCP server.
func (m *MCP) RecordAgentSession(ctx context.Context, userID int64, session AgentSession) error {
	return m.Call(ctx, userID, AppSablier, "sablier_record_agent_session", session, nil)
}

// RecordHeartbeat records one heartbeat for the user through Sablier's MCP
// server.
func (m *MCP) RecordHeartbeat(ctx context.Context, userID int64, beat Heartbeat) error {
	return m.Call(ctx, userID, AppSablier, "sablier_record_heartbeat", beat, nil)
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
