package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
	"github.com/robfig/cron/v3"
)

var cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

type createJobInput struct {
	Name     string `json:"name" jsonschema:"required,description=A short name describing the job"`
	Schedule string `json:"schedule" jsonschema:"required,description=Cron schedule expression (e.g. 0 9 * * * for 9am daily)"`
	Prompt   string `json:"prompt" jsonschema:"required,description=The instruction or prompt to execute on schedule"`
	Delivery string `json:"delivery,omitempty" jsonschema:"description=Optional notification target (default antenne)"`
}

type listJobsInput struct{}

type updateJobInput struct {
	ID       string  `json:"id" jsonschema:"required,description=The job ID to update"`
	Enabled  *bool   `json:"enabled,omitempty" jsonschema:"description=Whether the job is enabled"`
	Schedule *string `json:"schedule,omitempty" jsonschema:"description=New cron schedule expression"`
	Prompt   *string `json:"prompt,omitempty" jsonschema:"description=New prompt for the job"`
	Delivery *string `json:"delivery,omitempty" jsonschema:"description=New delivery target"`
}

type deleteJobInput struct {
	ID string `json:"id" jsonschema:"required,description=The job ID to delete"`
}

func buildJobTools(st store.Store, userID int64, channel string) []nacelle.Tool {
	if channel == store.ChannelCron || st == nil {
		return nil
	}
	t1, e1 := nacelle.NewTool("create_background_job", "Create a scheduled background job.", func(ctx context.Context, in createJobInput) (string, error) {
		return runCreateJob(ctx, st, userID, in)
	})
	t2, e2 := nacelle.NewToolWithOptions("list_background_jobs", "List all background jobs.", func(ctx context.Context, _ listJobsInput) (string, error) {
		return runListJobs(ctx, st, userID)
	}, nacelle.ToolOptions{ReadOnly: true})
	t3, e3 := nacelle.NewTool("update_background_job", "Update an existing background job.", func(ctx context.Context, in updateJobInput) (string, error) {
		return runUpdateJob(ctx, st, userID, in)
	})
	t4, e4 := nacelle.NewTool("delete_background_job", "Delete a background job by ID.", func(ctx context.Context, in deleteJobInput) (string, error) {
		return runDeleteJob(ctx, st, userID, in)
	})
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		return nil
	}
	return []nacelle.Tool{t1, t2, t3, t4}
}

func runCreateJob(ctx context.Context, st store.Store, userID int64, in createJobInput) (string, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return "", errors.New("job name is required")
	}
	schedule := strings.TrimSpace(in.Schedule)
	if schedule == "" {
		return "", errors.New("job schedule is required")
	}
	if _, err := cronParser.Parse(schedule); err != nil {
		return "", fmt.Errorf("invalid cron schedule: %w", err)
	}
	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" {
		return "", errors.New("job prompt is required")
	}
	delivery := strings.TrimSpace(in.Delivery)
	if delivery == "" {
		delivery = "antenne"
	}
	created, err := st.CreateJob(ctx, store.Job{
		UserID:   userID,
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

func runListJobs(ctx context.Context, st store.Store, userID int64) (string, error) {
	jobs, err := st.ListJobs(ctx, userID)
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
			return fmt.Errorf("invalid cron schedule: %w", err)
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
	if in.Delivery != nil {
		job.Delivery = strings.TrimSpace(*in.Delivery)
	}
	return nil
}

func runUpdateJob(ctx context.Context, st store.Store, userID int64, in updateJobInput) (string, error) {
	id := strings.TrimSpace(in.ID)
	if id == "" {
		return "", errors.New("job id is required")
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

func runDeleteJob(ctx context.Context, st store.Store, userID int64, in deleteJobInput) (string, error) {
	id := strings.TrimSpace(in.ID)
	if id == "" {
		return "", errors.New("job id is required")
	}
	if err := st.DeleteJob(ctx, userID, id); err != nil {
		return "", err
	}
	return fmt.Sprintf("job %s deleted", id), nil
}
