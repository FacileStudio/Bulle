package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

const (
	// MinJobInterval is the shortest gap a job's schedule may leave between
	// two runs.
	MinJobInterval = 15 * time.Minute
	// MaxJobsPerUser caps the jobs one user may keep, across every space.
	MaxJobsPerUser = 20
	jobRunsChecked = 200
)

// jobLimits is the store the job tools write through. It refuses a schedule
// that fires more often than MinJobInterval, and a job past MaxJobsPerUser,
// so neither the model nor an injected prompt can set up a runaway cron.
type jobLimits struct {
	store.Store
}

func (l jobLimits) CreateJob(ctx context.Context, job store.Job) (store.Job, error) {
	if err := checkInterval(job.Schedule, time.Now()); err != nil {
		return store.Job{}, err
	}
	jobs, err := l.ListJobs(ctx, job.UserID, "")
	if err != nil {
		return store.Job{}, err
	}
	if len(jobs) >= MaxJobsPerUser {
		return store.Job{}, fmt.Errorf("you already have %d jobs, the most one user can keep; delete one first", len(jobs))
	}
	return l.Store.CreateJob(ctx, job)
}

func (l jobLimits) UpdateJob(ctx context.Context, job store.Job) error {
	if err := checkInterval(job.Schedule, time.Now()); err != nil {
		return err
	}
	return l.Store.UpdateJob(ctx, job)
}

// checkInterval walks the schedule's next runs from now and refuses it when
// two of them are closer than MinJobInterval.
func checkInterval(schedule string, now time.Time) error {
	parsed, err := cronParser.Parse(schedule)
	if err != nil {
		return fmt.Errorf("invalid cron schedule: %w", err)
	}
	previous := parsed.Next(now)
	for range jobRunsChecked {
		next := parsed.Next(previous)
		if next.IsZero() {
			return nil
		}
		if gap := next.Sub(previous); gap < MinJobInterval {
			return fmt.Errorf("schedule %q has two runs only %s apart; jobs must be at least %s apart", schedule, gap, MinJobInterval)
		}
		previous = next
	}
	return nil
}
