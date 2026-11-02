package jobs

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
	"github.com/robfig/cron/v3"
)

type Handler struct {
	store      store.Store
	runner     *runner.Runner
	logger     *slog.Logger
	cronParser cron.Parser
}

func NewHandler(st store.Store, rn *runner.Runner, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		store:      st,
		runner:     rn,
		logger:     logger,
		cronParser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor),
	}
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	items, err := h.store.ListJobs(r.Context(), identity.UserID)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to list jobs", err))
		return
	}
	res := make([]JobResponse, 0, len(items))
	for _, item := range items {
		res = append(res, toJobResponse(item))
	}
	httpjson.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) createJob(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	var body CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpjson.WriteError(w, troncerrors.Invalid("invalid json payload"))
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		httpjson.WriteError(w, troncerrors.Invalid("job name is required"))
		return
	}
	schedule := strings.TrimSpace(body.Schedule)
	if schedule == "" {
		httpjson.WriteError(w, troncerrors.Invalid("job schedule is required"))
		return
	}
	if _, err := h.cronParser.Parse(schedule); err != nil {
		httpjson.WriteError(w, troncerrors.Invalid("invalid cron schedule expression"))
		return
	}
	prompt := strings.TrimSpace(body.Prompt)
	if prompt == "" {
		httpjson.WriteError(w, troncerrors.Invalid("job prompt is required"))
		return
	}
	delivery := strings.TrimSpace(body.Delivery)
	if delivery != "" && delivery != DeliveryPush {
		httpjson.WriteError(w, troncerrors.Invalid("delivery must be empty or \"push\""))
		return
	}
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	job, err := h.store.CreateJob(r.Context(), store.Job{
		UserID:   identity.UserID,
		Name:     name,
		Schedule: schedule,
		Prompt:   prompt,
		Profile:  strings.TrimSpace(body.Profile),
		Delivery: delivery,
		Enabled:  enabled,
	})
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to create job", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusCreated, toJobResponse(job))
}

func (h *Handler) getJob(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	job, err := h.store.GetJob(r.Context(), identity.UserID, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpjson.WriteError(w, troncerrors.NotFound("job not found"))
			return
		}
		httpjson.WriteError(w, troncerrors.Internal("failed to get job", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, toJobResponse(job))
}

func (h *Handler) updateJob(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	existing, err := h.store.GetJob(r.Context(), identity.UserID, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpjson.WriteError(w, troncerrors.NotFound("job not found"))
			return
		}
		httpjson.WriteError(w, troncerrors.Internal("failed to get job", err))
		return
	}

	var body UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpjson.WriteError(w, troncerrors.Invalid("invalid json payload"))
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		httpjson.WriteError(w, troncerrors.Invalid("job name is required"))
		return
	}
	schedule := strings.TrimSpace(body.Schedule)
	if schedule == "" {
		httpjson.WriteError(w, troncerrors.Invalid("job schedule is required"))
		return
	}
	if _, err := h.cronParser.Parse(schedule); err != nil {
		httpjson.WriteError(w, troncerrors.Invalid("invalid cron schedule expression"))
		return
	}
	prompt := strings.TrimSpace(body.Prompt)
	if prompt == "" {
		httpjson.WriteError(w, troncerrors.Invalid("job prompt is required"))
		return
	}
	delivery := strings.TrimSpace(body.Delivery)
	if delivery != "" && delivery != DeliveryPush && delivery != existing.Delivery {
		httpjson.WriteError(w, troncerrors.Invalid("delivery must be empty or \"push\""))
		return
	}
	enabled := existing.Enabled
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	updated := store.Job{
		ID:        existing.ID,
		UserID:    identity.UserID,
		Name:      name,
		Schedule:  schedule,
		Prompt:    prompt,
		Profile:   strings.TrimSpace(body.Profile),
		Delivery:  delivery,
		Enabled:   enabled,
		CreatedAt: existing.CreatedAt,
		UpdatedAt: time.Now().UTC(),
	}
	if err := h.store.UpdateJob(r.Context(), updated); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpjson.WriteError(w, troncerrors.NotFound("job not found"))
			return
		}
		httpjson.WriteError(w, troncerrors.Internal("failed to update job", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, toJobResponse(updated))
}

func (h *Handler) deleteJob(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.store.DeleteJob(r.Context(), identity.UserID, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpjson.WriteError(w, troncerrors.NotFound("job not found"))
			return
		}
		httpjson.WriteError(w, troncerrors.Internal("failed to delete job", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) runJob(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	job, err := h.store.GetJob(r.Context(), identity.UserID, id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpjson.WriteError(w, troncerrors.NotFound("job not found"))
			return
		}
		httpjson.WriteError(w, troncerrors.Internal("failed to get job", err))
		return
	}

	backend, model := "", ""
	if h.runner != nil {
		backend, model = h.runner.Defaults()
	}

	session, err := h.store.CreateSession(r.Context(), store.Session{
		ID:      store.NewID(),
		UserID:  identity.UserID,
		Channel: store.ChannelCron,
		Title:   job.Name,
		Backend: backend,
		Model:   model,
	})
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to create run session", err))
		return
	}

	runID, err := h.runner.Start(r.Context(), identity.UserID, session.ID, job.Prompt)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to start run", err))
		return
	}

	now := time.Now().UTC()
	jobRun := store.JobRun{
		ID:        store.NewID(),
		JobID:     job.ID,
		RunID:     runID,
		Status:    store.RunRunning,
		StartedAt: now,
	}
	if err := h.store.CreateJobRun(r.Context(), jobRun); err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to record job run", err))
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, RunJobResponse{
		RunID:  runID,
		JobRun: toJobRunResponse(jobRun),
	})
}

func (h *Handler) listJobRuns(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	id := chi.URLParam(r, "id")
	if _, err := h.store.GetJob(r.Context(), identity.UserID, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpjson.WriteError(w, troncerrors.NotFound("job not found"))
			return
		}
		httpjson.WriteError(w, troncerrors.Internal("failed to get job", err))
		return
	}

	runs, err := h.store.ListJobRuns(r.Context(), id)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to list job runs", err))
		return
	}
	res := make([]JobRunResponse, 0, len(runs))
	for _, run := range runs {
		res = append(res, toJobRunResponse(run))
	}
	httpjson.WriteJSON(w, http.StatusOK, res)
}

func toJobResponse(job store.Job) JobResponse {
	return JobResponse{
		ID:        job.ID,
		UserID:    job.UserID,
		Name:      job.Name,
		Schedule:  job.Schedule,
		Prompt:    job.Prompt,
		Profile:   job.Profile,
		Delivery:  job.Delivery,
		Enabled:   job.Enabled,
		CreatedAt: job.CreatedAt.UTC(),
		UpdatedAt: job.UpdatedAt.UTC(),
	}
}

func toJobRunResponse(run store.JobRun) JobRunResponse {
	return JobRunResponse{
		ID:         run.ID,
		JobID:      run.JobID,
		RunID:      run.RunID,
		Status:     run.Status,
		StartedAt:  run.StartedAt.UTC(),
		FinishedAt: run.FinishedAt,
	}
}
