package jobs

import "time"

// DeliveryPush notifies the user through ntfy when a run ends.
const DeliveryPush = "push"

type CreateJobRequest struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Prompt   string `json:"prompt"`
	Profile  string `json:"profile,omitempty"`
	Delivery string `json:"delivery,omitempty"`
	Enabled  *bool  `json:"enabled,omitempty"`
}

type UpdateJobRequest struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Prompt   string `json:"prompt"`
	Profile  string `json:"profile,omitempty"`
	Delivery string `json:"delivery,omitempty"`
	Enabled  *bool  `json:"enabled,omitempty"`
}

type JobResponse struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id,omitempty"`
	Name      string    `json:"name"`
	Schedule  string    `json:"schedule"`
	Prompt    string    `json:"prompt"`
	Profile   string    `json:"profile,omitempty"`
	Delivery  string    `json:"delivery,omitempty"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Running       bool   `json:"running"`
	LastSessionID string `json:"last_session_id,omitempty"`
}

type JobRunResponse struct {
	ID         string     `json:"id"`
	JobID      string     `json:"job_id"`
	RunID      string     `json:"run_id"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

type RunJobResponse struct {
	RunID  string         `json:"run_id"`
	JobRun JobRunResponse `json:"job_run"`
}
