package runner

import (
	"fmt"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// Job tool names, as the gate's approval policies key on them.
const (
	JobsCreateTool = "jobs_create"
	JobsListTool   = "jobs_list"
	JobsUpdateTool = "jobs_update"
	JobsDeleteTool = "jobs_delete"
)

const jobDeliveryPush = "push"

type createJobInput struct {
	Name     string `json:"name" jsonschema:"required,description=A short name the user recognises the job by"`
	Schedule string `json:"schedule" jsonschema:"required,description=Five-field cron expression such as 0 9 * * 1-5; prefix CRON_TZ=<IANA zone> to set its time zone"`
	Prompt   string `json:"prompt" jsonschema:"required,description=The instruction each run starts from, written as the user's own message"`
	Delivery string `json:"delivery,omitempty" jsonschema:"enum=none,enum=push,description=push notifies the user when a run ends; none (the default) only keeps the chat"`
}

type listJobsInput struct{}

type updateJobInput struct {
	JobID    string  `json:"job_id" jsonschema:"required,description=Job id as jobs_list answers it"`
	Enabled  *bool   `json:"enabled,omitempty" jsonschema:"description=false pauses the job and true resumes it"`
	Schedule *string `json:"schedule,omitempty" jsonschema:"description=New five-field cron expression, CRON_TZ=<IANA zone> prefix allowed"`
	Prompt   *string `json:"prompt,omitempty" jsonschema:"description=New instruction for each run"`
	Delivery *string `json:"delivery,omitempty" jsonschema:"enum=none,enum=push,description=push notifies the user when a run ends; none only keeps the chat"`
}

type deleteJobInput struct {
	JobID string `json:"job_id" jsonschema:"required,description=Job id as jobs_list answers it"`
}

// jobDescriptions are the four job tools' descriptions, in create, list,
// update, delete order. The server's clock zone is named because a schedule
// without CRON_TZ runs on it.
func jobDescriptions() [4]string {
	zone := time.Now().Format("MST, UTC-07:00")
	return [4]string{
		"Create a scheduled background job: on each tick of schedule, Bulle runs prompt in a new chat of its own, without the user present. " +
			"Use it only when the user asks for something recurring, such as a morning brief. " +
			"schedule is a five-field cron expression on the server clock (" + zone + "); prefix CRON_TZ=<IANA zone>, " +
			"as in CRON_TZ=Europe/Paris 0 9 * * 1-5, to run on the user's time. delivery push also notifies the user when a run ends.",
		"List the user's background jobs in this space with their job_id, schedule, prompt, delivery and whether they are enabled. " +
			"Use it to find a job_id before changing or deleting a job, or to check what already runs before creating another.",
		"Change a background job: pause or resume it with enabled, or change its schedule, prompt or delivery. Only the fields you pass change. " +
			"job_id comes from jobs_list. A schedule without a CRON_TZ= prefix runs on the server clock (" + zone + ").",
		"Delete a background job for good and stop any run of it still going. Use it only when the user asks to delete the job; " +
			"to pause it, set enabled to false with jobs_update. job_id comes from jobs_list.",
	}
}

// jobDelivery reads a delivery the model passed: none or empty keeps the
// chat only, push notifies the user when a run ends.
func jobDelivery(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "none":
		return "", nil
	case jobDeliveryPush:
		return jobDeliveryPush, nil
	}
	return "", fmt.Errorf("delivery %q is not valid; use none or push", value)
}

func applyDelivery(job *store.Job, value *string) error {
	if value == nil {
		return nil
	}
	delivery, err := jobDelivery(*value)
	if err != nil {
		return err
	}
	job.Delivery = delivery
	return nil
}
