package runner

import (
	"cmp"
	"context"
	"encoding/json"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

// jobWriteTools create, change or delete a background job. Both the jobs_*
// names and the older *_background_job ones are listed, so stored approvals
// and a half-renamed tool set stay gated.
var jobWriteTools = []string{
	"jobs_create", "jobs_update", "jobs_delete",
	"create_background_job", "update_background_job", "delete_background_job",
}

// OutwardTools is every host tool that always waits for the owner, beside the
// suite writes: browser_confirm and the job writes, since a job runs a prompt
// unattended on a schedule.
func OutwardTools() []string {
	return append([]string{browserConfirmTool}, jobWriteTools...)
}

type jobInput struct {
	ID       string  `json:"id"`
	JobID    string  `json:"job_id"`
	Name     string  `json:"name"`
	Schedule *string `json:"schedule"`
	Prompt   *string `json:"prompt"`
	Delivery *string `json:"delivery"`
	Enabled  *bool   `json:"enabled"`
}

// jobPolicy shows the owner the job a call would leave behind: its schedule
// and the prompt it will run, and for a change or a deletion the job as it
// stands now. It is asked every time.
type jobPolicy struct {
	store  store.Store
	userID int64
	action string
}

func (p jobPolicy) describe(ctx context.Context, input json.RawMessage) (map[string]any, string, error) {
	var in jobInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, "", err
	}
	record := map[string]any{"kind": "job", "action": p.action}
	put(record, "name", in.Name)
	putPtr(record, "schedule", in.Schedule)
	putPtr(record, "prompt", in.Prompt)
	putPtr(record, "delivery", in.Delivery)
	if in.Enabled != nil {
		record["enabled"] = *in.Enabled
	}
	id := strings.TrimSpace(cmp.Or(in.JobID, in.ID))
	if id == "" {
		return record, "", nil
	}
	record["job_id"] = id
	if job, err := p.store.GetJob(ctx, p.userID, id); err == nil {
		record["current"] = map[string]any{"name": job.Name, "schedule": job.Schedule, "prompt": job.Prompt, "enabled": job.Enabled}
	}
	return record, "", nil
}

func (p jobPolicy) denied(context.Context, json.RawMessage) {}

func put(record map[string]any, key, value string) {
	if value = strings.TrimSpace(value); value != "" {
		record[key] = value
	}
}

func putPtr(record map[string]any, key string, value *string) {
	if value != nil {
		put(record, key, *value)
	}
}

// prime readies a new run's gate: what the conversation already holds, and
// the job cards.
func (g *gate) prime(st store.Store, userID int64, history []nacelle.Message) {
	g.seed(history)
	if st == nil {
		return
	}
	for _, name := range jobWriteTools {
		action := "update"
		switch {
		case strings.Contains(name, "create"):
			action = "create"
		case strings.Contains(name, "delete"):
			action = "delete"
		}
		g.addPolicy(name, jobPolicy{store: st, userID: userID, action: action})
	}
}

// markTainted records on the session, once per run, that a run read a third
// party's text, so the memory pass keeps only the owner's words from then on
// whichever tool brought it. When the store cannot record it, unmarked keeps
// it in memory for the memory pass instead.
func (g *gate) markTainted(ctx context.Context) {
	g.mu.Lock()
	mark := g.trust.level != clean && !g.trust.marked && g.store != nil && g.session != ""
	g.trust.marked = g.trust.marked || mark
	g.mu.Unlock()
	if !mark {
		return
	}
	if err := g.store.MarkTainted(context.WithoutCancel(ctx), g.session); err != nil {
		g.emitter.logger.Error("marking a session tainted", "run", g.runID, "session", g.session, "error", err)
		if g.unmarked != nil {
			g.unmarked(g.session)
		}
	}
}
