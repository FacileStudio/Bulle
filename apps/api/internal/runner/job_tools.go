package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
	"github.com/robfig/cron/v3"
)

var cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

// RunStopper cancels a live run; *Runner satisfies it.
type RunStopper interface {
	Stop(ctx context.Context, userID int64, runID string) error
}

func buildJobTools(st store.Store, stopper RunStopper, userID int64, spaceID, channel string) []nacelle.Tool {
	if channel == store.ChannelCron || st == nil {
		return nil
	}
	st = jobLimits{Store: st}
	usage := jobDescriptions()
	t1, e1 := nacelle.NewTool(JobsCreateTool, usage[0], func(ctx context.Context, in createJobInput) (string, error) {
		return runCreateJob(ctx, st, store.Job{UserID: userID, SpaceID: spaceID}, in)
	})
	t2, e2 := nacelle.NewToolWithOptions(JobsListTool, usage[1], func(ctx context.Context, _ listJobsInput) (string, error) {
		return runListJobs(ctx, st, userID, spaceID)
	}, nacelle.ToolOptions{ReadOnly: true})
	t3, e3 := nacelle.NewTool(JobsUpdateTool, usage[2], func(ctx context.Context, in updateJobInput) (string, error) {
		return runUpdateJob(ctx, st, userID, in)
	})
	t4, e4 := nacelle.NewTool(JobsDeleteTool, usage[3], func(ctx context.Context, in deleteJobInput) (string, error) {
		return runDeleteJob(ctx, st, stopper, userID, in)
	})
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		return nil
	}
	return []nacelle.Tool{t1, t2, suite.Aliased(t3), suite.Aliased(t4)}
}

func runCreateJob(ctx context.Context, st store.Store, owner store.Job, in createJobInput) (string, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return "", errors.New("job name is required")
	}
	schedule := strings.TrimSpace(in.Schedule)
	if schedule == "" {
		return "", errors.New("job schedule is required")
	}
	if _, err := cronParser.Parse(schedule); err != nil {
		return "", fmt.Errorf("schedule %q is not a five-field cron expression such as 0 9 * * 1-5: %w", schedule, err)
	}
	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" {
		return "", errors.New("job prompt is required")
	}
	delivery, err := jobDelivery(in.Delivery)
	if err != nil {
		return "", err
	}
	created, err := st.CreateJob(ctx, store.Job{
		UserID: owner.UserID, SpaceID: owner.SpaceID,
		Name:     name,
		Schedule: schedule,
		Prompt:   prompt,
		Delivery: delivery,
		Enabled:  true,
	})
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(created)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func runListJobs(ctx context.Context, st store.Store, userID int64, spaceID string) (string, error) {
	jobs, err := st.ListJobs(ctx, userID, spaceID)
	if err != nil {
		return "", err
	}
	if jobs == nil {
		jobs = []store.Job{}
	}
	data, err := json.Marshal(jobs)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func applyJobUpdates(job *store.Job, in updateJobInput) error {
	if in.Enabled != nil {
		job.Enabled = *in.Enabled
	}
	if in.Schedule != nil {
		sched := strings.TrimSpace(*in.Schedule)
		if sched == "" {
			return errors.New("job schedule cannot be empty")
		}
		if _, err := cronParser.Parse(sched); err != nil {
			return fmt.Errorf("schedule %q is not a five-field cron expression such as 0 9 * * 1-5: %w", sched, err)
		}
		job.Schedule = sched
	}
	if in.Prompt != nil {
		prompt := strings.TrimSpace(*in.Prompt)
		if prompt == "" {
			return errors.New("job prompt cannot be empty")
		}
		job.Prompt = prompt
	}
	return applyDelivery(job, in.Delivery)
}

func runUpdateJob(ctx context.Context, st store.Store, userID int64, in updateJobInput) (string, error) {
	id := strings.TrimSpace(in.JobID)
	if id == "" {
		return "", errors.New("job_id is required; jobs_list answers each job's id")
	}
	job, err := st.GetJob(ctx, userID, id)
	if err != nil {
		return "", err
	}
	if err := applyJobUpdates(&job, in); err != nil {
		return "", err
	}
	if err := st.UpdateJob(ctx, job); err != nil {
		return "", err
	}
	data, err := json.Marshal(job)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func runDeleteJob(ctx context.Context, st store.Store, stopper RunStopper, userID int64, in deleteJobInput) (string, error) {
	id := strings.TrimSpace(in.JobID)
	if id == "" {
		return "", errors.New("job_id is required; jobs_list answers each job's id")
	}
	if err := DeleteJob(ctx, st, stopper, userID, id); err != nil {
		return "", err
	}
	return fmt.Sprintf("job %s deleted", id), nil
}

// DeleteJob stops every run of the job that is still going, then deletes the
// job. It returns store.ErrNotFound when the user owns no such job.
func DeleteJob(ctx context.Context, st store.Store, stopper RunStopper, userID int64, jobID string) error {
	if _, err := st.GetJob(ctx, userID, jobID); err != nil {
		return fmt.Errorf("get job %s: %w", jobID, err)
	}
	runs, err := st.ListJobRuns(ctx, jobID)
	if err != nil {
		return fmt.Errorf("list runs of job %s: %w", jobID, err)
	}
	for _, run := range runs {
		if run.Status != store.RunRunning {
			continue
		}
		if err := stopper.Stop(ctx, userID, run.RunID); err != nil && !errors.Is(err, store.ErrNotFound) {
			return fmt.Errorf("stop run %s: %w", run.RunID, err)
		}
	}
	if err := st.DeleteJob(ctx, userID, jobID); err != nil {
		return fmt.Errorf("delete job %s: %w", jobID, err)
	}
	return nil
}
