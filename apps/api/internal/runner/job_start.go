package runner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// JobStarter is what StartJob needs from the runner; *Runner satisfies it.
type JobStarter interface {
	CheckSpaceAccess(ctx context.Context, userID int64, spaceID string) error
	Defaults() (backend, model string)
	Start(ctx context.Context, userID int64, sessionID, text string) (string, error)
	RunStopper
}

// JobStart is a job run that started: its conversation and its job_runs row.
type JobStart struct {
	Session store.Session
	JobRun  store.JobRun
}

// StartJob runs the job once in a new conversation, for the scheduler and for
// "Run now" alike. A user who may not use AI, on their own plan or the plan of
// the shared space the job's space is linked to, is refused before any
// conversation exists, and a run that fails to start takes its conversation
// with it, so a job never leaves an empty unread conversation behind. The
// refusal or start error is returned unchanged for the caller to map.
func StartJob(ctx context.Context, st store.Store, rn JobStarter, job store.Job, at time.Time) (JobStart, error) {
	if err := rn.CheckSpaceAccess(ctx, job.UserID, job.SpaceID); err != nil {
		return JobStart{}, err
	}
	backend, model := rn.Defaults()
	session, err := st.CreateSession(ctx, store.Session{
		ID:      store.NewID(),
		UserID:  job.UserID,
		SpaceID: job.SpaceID,
		Channel: store.ChannelCron,
		Title:   job.Name,
		Backend: backend,
		Model:   model,
	})
	if err != nil {
		return JobStart{}, fmt.Errorf("create run session: %w", err)
	}
	runID, err := rn.Start(ctx, job.UserID, session.ID, job.Prompt)
	if err != nil {
		if delErr := st.DeleteSession(context.WithoutCancel(ctx), job.UserID, session.ID); delErr != nil {
			return JobStart{}, errors.Join(err, fmt.Errorf("delete run session: %w", delErr))
		}
		return JobStart{}, err
	}
	jobRun := store.JobRun{
		ID:        store.NewID(),
		JobID:     job.ID,
		RunID:     runID,
		Status:    store.RunRunning,
		StartedAt: at,
	}
	if err := st.CreateJobRun(ctx, jobRun); err != nil {
		stopErr := rn.Stop(context.WithoutCancel(ctx), job.UserID, runID)
		return JobStart{}, errors.Join(fmt.Errorf("record job run: %w", err), stopErr)
	}
	return JobStart{Session: session, JobRun: jobRun}, nil
}
