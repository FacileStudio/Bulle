package runner

import (
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

var trackedStart = time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)

func trackedFor(wall, waited time.Duration, channel string) trackedRun {
	return trackedRun{
		runID: "r1", session: store.Session{ID: "s1", Channel: channel, Title: "plan the week"},
		started: trackedStart, stopped: trackedStart.Add(wall), waited: waited,
		usage: nacelle.Usage{InputTokens: 10, OutputTokens: 3},
	}
}

func TestAgentSessionCountsWallTimeMinusApprovalWait(t *testing.T) {
	profile := store.AssistantProfile{Name: "Jarvis", AvatarURL: "https://blobatar.dev/avatar/Jarvis?gen=2"}
	got, ok := agentSession(trackedFor(10*time.Minute, 4*time.Minute, store.ChannelWeb), profile)
	if !ok {
		t.Fatal("a ten minute run must be tracked")
	}
	if got.ActiveSeconds != 360 || got.FacileID != "bulle_run_r1" || got.AgentKey != "bulle" || !got.Cowork {
		t.Fatalf("unexpected session: %+v", got)
	}
	if got.Name != "Jarvis" || got.AvatarURL != profile.AvatarURL || got.Description != "plan the week" || got.TokensIn != 10 || got.TokensOut != 3 {
		t.Fatalf("unexpected identity or details: %+v", got)
	}
	if !got.StartedAt.Equal(trackedStart) || !got.StoppedAt.Equal(trackedStart.Add(10*time.Minute)) {
		t.Fatalf("unexpected span: %s to %s", got.StartedAt, got.StoppedAt)
	}
}

func TestAgentSessionFallsBackToBulle(t *testing.T) {
	got, _ := agentSession(trackedFor(time.Minute, 0, store.ChannelCron), store.AssistantProfile{})
	if got.Name != "Bulle" || got.AvatarURL != "https://blobatar.dev/avatar/Bulle?gen=2" || got.Cowork {
		t.Fatalf("want the Bulle fallback and no cowork for cron, got %+v", got)
	}
	named, _ := agentSession(trackedFor(time.Minute, 0, store.ChannelMatrix), store.AssistantProfile{Name: "Ada Lovelace"})
	if named.AvatarURL != "https://blobatar.dev/avatar/Ada%20Lovelace?gen=2" || !named.Cowork {
		t.Fatalf("want a blobatar for the profile name and cowork for matrix, got %+v", named)
	}
}

func TestAgentSessionSkipsShortRunsAndNeverGoesNegative(t *testing.T) {
	if _, ok := agentSession(trackedFor(4*time.Second, 0, store.ChannelWeb), store.AssistantProfile{}); ok {
		t.Fatal("a run under five seconds must be skipped")
	}
	got, ok := agentSession(trackedFor(time.Minute, 2*time.Minute, store.ChannelWeb), store.AssistantProfile{})
	if !ok || got.ActiveSeconds != 0 {
		t.Fatalf("want zero active seconds, got %+v", got)
	}
}

func TestHeartbeatIsDueOncePerSessionEveryTwoMinutes(t *testing.T) {
	now := trackedStart
	r := New(Config{Now: func() time.Time { return now }})
	if !r.beatDue("s1") || r.beatDue("s1") || !r.beatDue("s2") {
		t.Fatal("want one beat per session")
	}
	now = now.Add(119 * time.Second)
	if r.beatDue("s1") {
		t.Fatal("a beat within two minutes must be skipped")
	}
	now = now.Add(time.Second)
	if !r.beatDue("s1") || len(r.beats) != 1 {
		t.Fatalf("want a new beat after two minutes and stale sessions pruned, got %d entries", len(r.beats))
	}
}
