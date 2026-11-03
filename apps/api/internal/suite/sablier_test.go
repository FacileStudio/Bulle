package suite

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRecordAgentSessionPostsWithTheWriteScope(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	started := time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)
	session := AgentSession{AgentKey: "bulle", Name: "Jarvis", FacileID: "bulle_run_r1", StartedAt: started, StoppedAt: started.Add(time.Minute), ActiveSeconds: 42, Cowork: true}
	if err := NewSablier(NewGateway(w.tokens(), w.cfg, nil)).RecordAgentSession(context.Background(), 1, session); err != nil {
		t.Fatal(err)
	}
	if len(w.writes) != 1 || w.writes[0].method != "POST" || w.writes[0].path != "/api/agents/sessions" {
		t.Fatalf("want one POST to /api/agents/sessions, got %+v", w.writes)
	}
	var sent map[string]any
	if err := json.Unmarshal([]byte(w.writes[0].body), &sent); err != nil {
		t.Fatal(err)
	}
	if sent["agent_key"] != "bulle" || sent["facile_id"] != "bulle_run_r1" || sent["active_seconds"] != float64(42) || sent["cowork"] != true || sent["started_at"] != "2026-11-03T09:00:00Z" {
		t.Fatalf("unexpected payload: %v", sent)
	}
}

func TestHeartbeatWithoutWriteScopeLeavesReadsWorking(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.readOnly = true
	sablier := NewSablier(NewGateway(w.tokens(), w.cfg, nil))
	beat := Heartbeat{At: time.Now(), Source: "bulle", Entity: "bulle:session/s1"}
	if err := sablier.RecordHeartbeat(context.Background(), 1, beat); !errors.Is(err, ErrWriteNotGranted) || len(w.writes) != 0 {
		t.Fatalf("want a write-not-granted error and no request, got %v, %d writes", err, len(w.writes))
	}
	if body, err := sablier.Running(context.Background(), 1); err != nil || !strings.Contains(string(body), "timer of 1") {
		t.Fatalf("reads must keep working when writes are not granted: %s, %v", body, err)
	}
}
