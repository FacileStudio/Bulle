package suite

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const sablierRunningPath = "/api/time-entries/running"

type Sablier struct{ gw *Gateway }

func NewSablier(gw *Gateway) Sablier { return Sablier{gw: gw} }

func (s Sablier) Running(ctx context.Context, userID int64) ([]byte, error) {
	return s.gw.Get(ctx, userID, AppSablier, sablierRunningPath, nil)
}

const (
	sablierProjectsPath = "/api/projects/"
	sablierTasksPath    = "/api/projects/%d/tasks"
	sablierStartPath    = "/api/time-entries/start"
	sablierStopPath     = "/api/time-entries/stop"
)

func (s Sablier) Projects(ctx context.Context, userID, projectID int64) ([]byte, error) {
	if projectID > 0 {
		return s.gw.Get(ctx, userID, AppSablier, fmt.Sprintf(sablierTasksPath, projectID), nil)
	}
	return s.gw.Get(ctx, userID, AppSablier, sablierProjectsPath, nil)
}

func (s Sablier) Start(ctx context.Context, userID, projectID, taskID int64) ([]byte, error) {
	payload := map[string]any{"project_id": projectID, "task_id": taskID}
	return s.gw.Do(ctx, userID, Request{App: AppSablier, Method: http.MethodPost, Path: sablierStartPath, Body: payload})
}

func (s Sablier) Stop(ctx context.Context, userID int64) ([]byte, error) {
	return s.gw.Do(ctx, userID, Request{App: AppSablier, Method: http.MethodPost, Path: sablierStopPath, Body: nil})
}

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
}

// Heartbeat is one sign that the user was working at a given moment.
type Heartbeat struct {
	At      time.Time `json:"at"`
	Source  string    `json:"source"`
	Entity  string    `json:"entity"`
	Project string    `json:"project"`
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
