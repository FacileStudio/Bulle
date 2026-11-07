package runner

import (
	"cmp"
	"context"
	"encoding/json"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

const (
	agentKey          = "bulle"
	agentFallbackName = "Bulle"
	minTrackedRun     = 5 * time.Second
	trackingTimeout   = 10 * time.Second
	heartbeatInterval = 2 * time.Minute
	maxAvatarURL      = 2048
)

type trackedRun struct {
	runID   string
	session store.Session
	started time.Time
	stopped time.Time
	waited  time.Duration
	usage   nacelle.Usage
	space   store.Space
	opus    []string
}

func agentSession(run trackedRun, profile store.AssistantProfile) (suite.AgentSession, bool) {
	wall := run.stopped.Sub(run.started)
	if wall < minTrackedRun {
		return suite.AgentSession{}, false
	}
	name := cmp.Or(strings.TrimSpace(profile.Name), agentFallbackName)
	return suite.AgentSession{
		AgentKey:      agentKey,
		Name:          name,
		AvatarURL:     sablierAvatar(profile.AvatarURL),
		FacileID:      "bulle_run_" + run.runID,
		StartedAt:     run.started.UTC(),
		StoppedAt:     run.stopped.UTC(),
		ActiveSeconds: int64(max(wall-run.waited, 0) / time.Second),
		TokensIn:      run.usage.InputTokens,
		TokensOut:     run.usage.OutputTokens,
		Cowork:        interactive(run.session.Channel),
		Description:   run.session.Title,

		RegistreSpaceID: run.space.RegistreSpaceID,
		OpusWorkspace:   run.space.OpusWorkspace,
		OpusTaskIDs:     run.opus,
		Title:           run.session.Title,
	}, true
}

// sablierAvatar keeps the profile's avatar only when Sablier accepts it: an
// https URL of at most 2048 bytes with no credentials. Sablier refuses the
// whole session otherwise, and draws the agent's own blob for an empty one.
func sablierAvatar(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || len(raw) > maxAvatarURL || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return ""
	}
	return raw
}

func interactive(channel string) bool {
	return channel != store.ChannelCron && !store.Hidden(channel)
}

func (r *Runner) tracking() bool {
	return r.cfg.TimeTracking && r.cfg.Suite != nil && r.cfg.Suite.Config().Apps[suite.AppSablier].Enabled()
}

func (r *Runner) recordAgentTime(ctx context.Context, lr *liveRun, session store.Session, usage nacelle.Usage) {
	if !r.tracking() || !lr.engaged {
		return
	}
	run := trackedRun{runID: lr.id, session: session, started: lr.started, stopped: r.now(), waited: lr.gate.waited(), usage: usage}
	r.inBackground(ctx, lr.userID, "recording agent time in Sablier", func(ctx context.Context) error {
		profile, _ := r.cfg.Store.GetAssistantProfile(ctx, lr.userID)
		run.space, _ = r.cfg.Store.GetSpace(ctx, lr.userID, session.SpaceID)
		run.opus = r.opusTaskIDs(ctx, session.ID, r.runCalls(ctx, lr.id))
		payload, ok := agentSession(run, profile)
		if !ok {
			return nil
		}
		return suite.NewSablier(r.cfg.Suite).RecordAgentSession(ctx, lr.userID, payload)
	})
}

func (r *Runner) recordPresence(ctx context.Context, userID int64, session store.Session) {
	if !r.tracking() || !interactive(session.Channel) || !r.beatDue(session.ID) {
		return
	}
	beat := suite.Heartbeat{At: r.now().UTC(), Source: agentKey, Entity: "bulle:session/" + session.ID}
	r.inBackground(ctx, userID, "recording presence in Sablier", func(ctx context.Context) error {
		space, _ := r.cfg.Store.GetSpace(ctx, userID, session.SpaceID)
		beat.RegistreSpaceID = space.RegistreSpaceID
		if ids := r.opusTaskIDs(ctx, session.ID, nil); len(ids) > 0 {
			beat.OpusTaskID = ids[len(ids)-1]
		}
		return suite.NewSablier(r.cfg.Suite).RecordHeartbeat(ctx, userID, beat)
	})
}

// runCalls is the set of tool call ids one run made, read from its events.
func (r *Runner) runCalls(ctx context.Context, runID string) map[string]bool {
	events, _ := r.cfg.Store.Events(ctx, runID, 0)
	calls := map[string]bool{}
	for _, event := range events {
		var start struct {
			CallID string `json:"call_id"`
		}
		if event.Type == EventToolStart && json.Unmarshal(event.Payload, &start) == nil {
			calls[start.CallID] = true
		}
	}
	return calls
}

// opusTaskIDs lists the Opus facile ids the session's stored tool results name,
// most recently touched last. A nil calls set keeps every result; otherwise only
// the results of those calls count.
func (r *Runner) opusTaskIDs(ctx context.Context, sessionID string, calls map[string]bool) []string {
	messages, _ := r.cfg.Store.Messages(ctx, sessionID)
	var ids []string
	for _, message := range messages {
		wire, _ := DecodeWire(message.Payload)
		for _, part := range wire.Parts {
			id, ok := suite.OpusTaskID(part.Name, part.Result)
			if ok && !part.Failed && (calls == nil || calls[part.ID]) {
				ids = append(slices.DeleteFunc(ids, func(seen string) bool { return seen == id }), id)
			}
		}
	}
	return ids
}

func (r *Runner) beatDue(sessionID string) bool {
	now := r.now()
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, at := range r.beats {
		if now.Sub(at) >= heartbeatInterval {
			delete(r.beats, id)
		}
	}
	if _, recent := r.beats[sessionID]; recent {
		return false
	}
	r.beats[sessionID] = now
	return true
}

func (r *Runner) inBackground(ctx context.Context, userID int64, what string, call func(context.Context) error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), trackingTimeout)
	go func() {
		defer cancel()
		if err := call(ctx); err != nil {
			r.cfg.Logger.Warn(what, "user", userID, "error", err)
		}
	}()
}
