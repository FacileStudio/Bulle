package runner

import (
	"cmp"
	"context"
	"errors"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
)

var (
	// ErrMemoryOff answers a memory pass when the instance has no Mycelium.
	ErrMemoryOff = errors.New("runner: memory is not configured")
	// ErrMemoryBusy answers a memory pass while the user already has one
	// running, or while the chat itself still runs.
	ErrMemoryBusy = errors.New("runner: a memory pass or the chat is still running")
	// ErrNotAChat answers a memory pass on a session that is not a web or
	// matrix chat.
	ErrNotAChat = errors.New("runner: only web and matrix chats get a memory pass")
)

type memoryRun struct {
	userID  int64
	source  string
	space   store.Space
	session string
	upto    int
}

// MemoryAvailable reports whether the user reaches Mycelium, so the scheduler
// skips the memory pass of a user without it.
func (r *Runner) MemoryAvailable(ctx context.Context, userID int64) bool {
	return r.checkAccess(ctx, userID) == nil && r.myceliumEnabled() && suite.NewMycelium(r.cfg.Suite).Available(ctx, userID)
}

// MemoryPass learns from the part of a chat no pass has read yet. It builds a
// transcript of the user and assistant text only, never tool calls, tool
// results or attachments, and starts a run on a hidden memory session in the
// same space whose agent has only the memory tools. When that run ends the
// chat is marked seen up to the last message the transcript held. A user has
// at most one pass at a time.
func (r *Runner) MemoryPass(ctx context.Context, userID int64, sessionID string) error {
	if !r.myceliumEnabled() {
		return ErrMemoryOff
	}
	source, err := r.cfg.Store.GetSession(ctx, userID, sessionID)
	if err != nil {
		return err
	}
	if !store.Conversational(source.Channel) {
		return ErrNotAChat
	}
	if !r.claimMemory(userID) {
		return ErrMemoryBusy
	}
	run, err := r.startMemory(ctx, source)
	if err != nil || run == nil {
		r.releaseMemory(userID)
		return err
	}
	return nil
}

func (r *Runner) startMemory(ctx context.Context, source store.Session) (*memoryRun, error) {
	if _, active, err := r.cfg.Store.ActiveRun(ctx, source.ID); err != nil || active {
		return nil, cmp.Or(err, ErrMemoryBusy)
	}
	messages, err := r.cfg.Store.Messages(ctx, source.ID)
	if err != nil {
		return nil, err
	}
	transcript, upto := passTranscript(messages, source.MemorySeenSeq)
	if transcript == "" {
		return nil, r.cfg.Store.MarkMemorySeen(ctx, source.ID, upto)
	}
	space, err := r.cfg.Store.GetSpace(ctx, source.UserID, source.SpaceID)
	if err != nil {
		return nil, err
	}
	run := &memoryRun{userID: source.UserID, source: source.ID, space: space, upto: upto}
	runID, err := r.launchMemory(ctx, run, source.Title, transcript)
	if err != nil {
		return nil, err
	}
	go r.watchMemory(run, runID)
	return run, nil
}

func (r *Runner) launchMemory(ctx context.Context, run *memoryRun, title, transcript string) (string, error) {
	backend, model := r.Defaults()
	session, err := r.cfg.Store.CreateSession(ctx, store.Session{
		ID: store.NewID(), UserID: run.userID, SpaceID: run.space.ID, Channel: store.ChannelMemory,
		Title: titleFrom("Memory from " + cmp.Or(strings.TrimSpace(title), "a chat")), Backend: backend, Model: model,
	})
	if err != nil {
		return "", err
	}
	run.session = session.ID
	r.mu.Lock()
	r.memoryRuns[session.ID] = run
	r.mu.Unlock()
	prompt := "Learn from this conversation with the user.\n\n<conversation>\n" + transcript + "\n</conversation>"
	runID, err := r.Start(ctx, run.userID, session.ID, prompt)
	if err != nil {
		r.forgetMemory(run)
		_ = r.cfg.Store.DeleteSession(context.WithoutCancel(ctx), run.userID, session.ID)
	}
	return runID, err
}

func (r *Runner) claimMemory(userID int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.memoryUsers[userID] {
		return false
	}
	r.memoryUsers[userID] = true
	return true
}

func (r *Runner) releaseMemory(userID int64) {
	r.mu.Lock()
	delete(r.memoryUsers, userID)
	r.mu.Unlock()
}

func (r *Runner) memoryRunFor(sessionID string) (*memoryRun, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	run, ok := r.memoryRuns[sessionID]
	return run, ok
}

func (r *Runner) forgetMemory(run *memoryRun) {
	r.mu.Lock()
	delete(r.memoryRuns, run.session)
	r.mu.Unlock()
}
