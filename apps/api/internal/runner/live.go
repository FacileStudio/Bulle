package runner

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/live"
	"github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

// execute runs the turn to its end, then drops the run from the live set
// before it announces the outcome, so a list fetched on run.finished is idle.
func (r *Runner) execute(ctx context.Context, lr *liveRun, session store.Session, history []nacelle.Message) {
	defer r.release(lr.userID)
	status := r.complete(ctx, lr, session, history)
	r.mu.Lock()
	delete(r.live, lr.id)
	r.mu.Unlock()
	r.announceFinish(lr, status)
}

func (r *Runner) publish(userID int64, eventType string, data any) {
	if r.cfg.Push != nil {
		r.cfg.Push.Publish(userID, eventType, data)
	}
}

func (r *Runner) announceSession(ctx context.Context, lr *liveRun) {
	session, err := r.cfg.Store.GetSession(ctx, lr.userID, lr.session.ID)
	if err != nil || store.Hidden(session.Channel) {
		return
	}
	r.publish(lr.userID, live.SessionUpdated, live.SessionOf(session, r.sessionStatus(lr.userID, session.ID)))
}

func (r *Runner) runEvent(lr *liveRun, outcome string) live.Run {
	return live.Run{
		SessionID: lr.session.ID, RunID: lr.id, Channel: lr.session.Channel,
		Status: outcome, SessionStatus: r.sessionStatus(lr.userID, lr.session.ID),
	}
}

func (r *Runner) announceStart(ctx context.Context, lr *liveRun) {
	r.publish(lr.userID, live.RunStarted, r.runEvent(lr, ""))
	r.announceSession(context.WithoutCancel(ctx), lr)
}

func (r *Runner) announceFinish(lr *liveRun, outcome string) {
	r.publish(lr.userID, live.RunFinished, r.runEvent(lr, outcome))
	r.announceSession(r.base, lr)
}

func (r *Runner) approvalPending(lr *liveRun) func(callID, tool string) {
	return func(callID, tool string) {
		r.publish(lr.userID, live.ApprovalPending, live.Approval{
			SessionID: lr.session.ID, RunID: lr.id, CallID: callID, Tool: tool, SessionStatus: live.StatusWaiting,
		})
		if r.cfg.Push == nil || r.hub.watched(lr.id) {
			return
		}
		note := push.Notification{Title: "Bulle", Body: "Approval needed", Path: "sessions/" + lr.session.ID}
		go r.cfg.Push.Notify(r.base, lr.userID, note)
	}
}

func (r *Runner) approvalDecided(lr *liveRun) func(callID string, approved bool) {
	return func(callID string, approved bool) {
		r.publish(lr.userID, live.ApprovalDecided, live.Approval{
			SessionID: lr.session.ID, RunID: lr.id, CallID: callID, Approved: &approved,
			SessionStatus: r.sessionStatus(lr.userID, lr.session.ID),
		})
	}
}
