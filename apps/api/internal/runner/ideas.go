package runner

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// ErrIdeasBusy answers a batch request for a space that already has one running.
var ErrIdeasBusy = errors.New("runner: an ideas batch is already running")

var errNoIdeasBatch = errors.New("this conversation is not an ideas batch in progress")

type ideaRun struct {
	batch   store.IdeaBatch
	space   store.Space
	session string

	mu    sync.Mutex
	count int
}

// GenerateIdeas starts one ideas batch over every app of the user's space and
// returns its id. See GenerateAppIdeas.
func (r *Runner) GenerateIdeas(ctx context.Context, userID int64, spaceID, trigger string) (string, error) {
	return r.GenerateAppIdeas(ctx, userID, spaceID, "", trigger)
}

// GenerateAppIdeas starts one ideas batch for the user's space and returns
// its id. The user's own plan or, for a space linked to Registre, that
// space's plan must allow AI. With an app, the batch reads only that app's
// tools, follows that app's guidance and files every idea under it. The run reads with read-only
// tools and leaves at most seven ideas through idea_propose; when it ends, the
// batch is finished, the older new ideas of the same scope expire, and a daily
// batch with ideas sends one push when the space asked for it.
func (r *Runner) GenerateAppIdeas(ctx context.Context, userID int64, spaceID, app, trigger string) (string, error) {
	space, err := r.cfg.Store.GetSpace(ctx, userID, spaceID)
	if err != nil {
		return "", err
	}
	if err := r.checkAccessIn(ctx, userID, space.RegistreSpaceID); err != nil {
		return "", err
	}
	batch, err := r.cfg.Store.CreateIdeaBatch(ctx, store.IdeaBatch{UserID: userID, SpaceID: space.ID, App: app, Trigger: trigger})
	if errors.Is(err, store.ErrIdeaBatchRunning) {
		return "", ErrIdeasBusy
	}
	if err != nil {
		return "", err
	}
	run := &ideaRun{batch: batch, space: space}
	runID, err := r.startIdeas(ctx, run)
	if err != nil {
		r.endBatch(context.WithoutCancel(ctx), run, store.BatchError)
		return "", err
	}
	run.batch.RunID = runID
	if err := r.cfg.Store.UpdateIdeaBatch(ctx, run.batch); err != nil {
		r.cfg.Logger.Warn("recording the ideas run", "batch", batch.ID, "error", err)
	}
	go r.watchIdeas(run, runID)
	return batch.ID, nil
}

func (r *Runner) startIdeas(ctx context.Context, run *ideaRun) (string, error) {
	prompt, err := r.ideasPrompt(ctx, run)
	if err != nil {
		return "", err
	}
	backend, model := r.Defaults()
	title := "Ideas for " + run.space.Name
	if run.batch.App != "" {
		title = "Ideas for " + run.space.Name + " in " + appLabel(run.batch.App)
	}
	session, err := r.cfg.Store.CreateSession(ctx, store.Session{
		ID: store.NewID(), UserID: run.batch.UserID, SpaceID: run.space.ID, Channel: store.ChannelIdeas,
		Title: title, Backend: backend, Model: model,
	})
	if err != nil {
		return "", err
	}
	run.session = session.ID
	r.mu.Lock()
	r.ideaRuns[session.ID] = run
	r.mu.Unlock()
	runID, err := r.Start(ctx, run.batch.UserID, session.ID, prompt)
	if err != nil {
		r.forgetIdeas(run)
		_ = r.cfg.Store.DeleteSession(context.WithoutCancel(ctx), run.batch.UserID, session.ID)
	}
	return runID, err
}

func (r *Runner) ideaRunFor(sessionID string) (*ideaRun, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	run, ok := r.ideaRuns[sessionID]
	return run, ok
}

func (r *Runner) forgetIdeas(run *ideaRun) {
	r.mu.Lock()
	delete(r.ideaRuns, run.session)
	r.mu.Unlock()
}

func (r *Runner) watchIdeas(run *ideaRun, runID string) {
	notify, unsubscribe := r.Subscribe(runID)
	defer unsubscribe()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for !r.ideasEnded(run, runID) {
		select {
		case <-r.base.Done():
			return
		case <-notify:
		case <-ticker.C:
		}
	}
}

func (r *Runner) ideasEnded(run *ideaRun, runID string) bool {
	if r.IsLive(runID) {
		return false
	}
	found, err := r.cfg.Store.GetRun(r.base, run.batch.UserID, runID)
	if err == nil && found.Status == store.RunRunning {
		return false
	}
	status := store.BatchError
	if err == nil && found.Status == store.RunFinished {
		status = store.BatchFinished
	}
	r.finishIdeas(run, status)
	return true
}

func (r *Runner) finishIdeas(run *ideaRun, status string) {
	r.forgetIdeas(run)
	ctx := r.base
	defer r.endBatch(ctx, run, status)
	run.mu.Lock()
	count := run.count
	run.mu.Unlock()
	if count == 0 {
		return
	}
	if err := r.cfg.Store.ExpireIdeas(ctx, run.batch.UserID, run.space.ID, run.batch.App, run.batch.ID); err != nil {
		r.cfg.Logger.Warn("expiring older ideas", "batch", run.batch.ID, "error", err)
	}
	r.pushIdeas(ctx, run, count)
}

func (r *Runner) endBatch(ctx context.Context, run *ideaRun, status string) {
	now := r.now().UTC()
	run.batch.Status, run.batch.FinishedAt = status, &now
	if err := r.cfg.Store.UpdateIdeaBatch(ctx, run.batch); err != nil {
		r.cfg.Logger.Warn("finishing the ideas batch", "batch", run.batch.ID, "error", err)
	}
}

func (r *Runner) pushIdeas(ctx context.Context, run *ideaRun, count int) {
	if r.cfg.Push == nil || run.batch.Trigger != store.BatchDaily {
		return
	}
	settings, err := r.cfg.Store.GetIdeaSettings(ctx, run.batch.UserID, run.space.ID)
	if err != nil || !settings.Push {
		return
	}
	body := fmt.Sprintf("%d ideas for %s", count, run.space.Name)
	if count == 1 {
		body = "1 idea for " + run.space.Name
	}
	r.cfg.Push.Notify(ctx, run.batch.UserID, push.Notification{Title: "Bulle", Body: body, Path: "ideas"})
}
