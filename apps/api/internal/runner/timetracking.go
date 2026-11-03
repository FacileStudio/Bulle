package runner

import (
	"cmp"
	"context"
	"net/url"
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
)

type trackedRun struct {
	runID   string
	session store.Session
	started time.Time
	stopped time.Time
	waited  time.Duration
	usage   nacelle.Usage
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
		AvatarURL:     cmp.Or(strings.TrimSpace(profile.AvatarURL), blobatar(name)),
		FacileID:      "bulle_run_" + run.runID,
		StartedAt:     run.started.UTC(),
		StoppedAt:     run.stopped.UTC(),
		ActiveSeconds: int64(max(wall-run.waited, 0) / time.Second),
		TokensIn:      run.usage.InputTokens,
		TokensOut:     run.usage.OutputTokens,
		Cowork:        interactive(run.session.Channel),
		Description:   run.session.Title,
	}, true
}

func blobatar(name string) string {
	return "https://blobatar.dev/avatar/" + url.PathEscape(name) + "?gen=2"
}

func interactive(channel string) bool {
	return channel != store.ChannelCron
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
		return suite.NewSablier(r.cfg.Suite).RecordHeartbeat(ctx, userID, beat)
	})
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
