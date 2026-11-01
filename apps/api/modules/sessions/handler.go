package sessions

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store  store.Store
	runner *runner.Runner
	logger *slog.Logger
}

func NewHandler(st store.Store, rn *runner.Runner, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		store:  st,
		runner: rn,
		logger: logger,
	}
}

func (h *Handler) listSessions(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	items, err := h.store.ListSessions(r.Context(), identity.UserID)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to list sessions", err))
		return
	}
	summaries := make([]SessionSummary, 0, len(items))
	for _, item := range items {
		summaries = append(summaries, SessionSummary{
			ID:        item.ID,
			Title:     item.Title,
			UpdatedAt: item.UpdatedAt.UTC(),
		})
	}
	httpjson.WriteJSON(w, http.StatusOK, summaries)
}

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	var body CreateSessionRequest
	_ = httpjson.DecodeJSON(w, r, &body)

	defBackend, defModel := h.runner.Defaults()
	s := store.Session{
		ID:        store.NewID(),
		UserID:    identity.UserID,
		Channel:   store.ChannelWeb,
		Title:     strings.TrimSpace(body.Title),
		Backend:   defBackend,
		Model:     defModel,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	created, err := h.store.CreateSession(r.Context(), s)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to create session", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusCreated, CreateSessionResponse{
		ID:    created.ID,
		Title: created.Title,
	})
}

func (h *Handler) getSession(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	sessionID := chi.URLParam(r, "id")
	session, err := h.store.GetSession(r.Context(), identity.UserID, sessionID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpjson.WriteError(w, troncerrors.NotFound("session not found"))
			return
		}
		httpjson.WriteError(w, troncerrors.Internal("failed to load session", err))
		return
	}
	stored, err := h.store.Messages(r.Context(), sessionID)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to load messages", err))
		return
	}
	activeRun, hasActive, err := h.store.ActiveRun(r.Context(), sessionID)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to check active run", err))
		return
	}
	if !hasActive {
		latest, hasLatest, err := h.store.LatestRun(r.Context(), sessionID)
		if err != nil {
			httpjson.WriteError(w, troncerrors.Internal("failed to check latest run", err))
			return
		}
		if hasLatest && latest.Status == store.RunError {
			activeRun = latest
			hasActive = true
		}
	}

	messages := make([]StoredMessage, 0, len(stored))
	for i, raw := range stored {
		role, content, err := runner.MessageText(raw.Payload)
		if err != nil {
			continue
		}
		if hasActive && i == len(stored)-1 && role == "assistant" {
			continue
		}
		messages = append(messages, StoredMessage{
			Seq:     raw.Seq,
			Role:    role,
			Content: content,
		})
	}

	res := SessionDetailResponse{
		Session: SessionSummary{
			ID:        session.ID,
			Title:     session.Title,
			UpdatedAt: session.UpdatedAt.UTC(),
		},
		Messages: messages,
	}
	if hasActive {
		res.ActiveRunID = activeRun.ID
	}
	httpjson.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) deleteSession(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	sessionID := chi.URLParam(r, "id")
	if err := h.store.DeleteSession(r.Context(), identity.UserID, sessionID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpjson.WriteError(w, troncerrors.NotFound("session not found"))
			return
		}
		httpjson.WriteError(w, troncerrors.Internal("failed to delete session", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, EmptyResponse{})
}

func (h *Handler) sendMessage(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	sessionID := chi.URLParam(r, "id")
	var body SendMessageRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	runID, err := h.runner.Start(r.Context(), identity.UserID, sessionID, body.Text)
	if err != nil {
		switch {
		case errors.Is(err, runner.ErrEmptyMessage):
			httpjson.WriteError(w, troncerrors.Invalid("message cannot be empty"))
		case errors.Is(err, runner.ErrTooManyRuns):
			httpjson.WriteError(w, troncerrors.RateLimited("too many active runs"))
		case errors.Is(err, runner.ErrShuttingDown):
			httpjson.WriteError(w, troncerrors.Unavailable("server is shutting down"))
		case errors.Is(err, store.ErrNotFound):
			httpjson.WriteError(w, troncerrors.NotFound("session not found"))
		case errors.Is(err, store.ErrSessionBusy):
			httpjson.WriteError(w, troncerrors.Conflict("session already has an active run"))
		default:
			httpjson.WriteError(w, troncerrors.Internal("failed to start run", err))
		}
		return
	}
	httpjson.WriteJSON(w, http.StatusAccepted, SendMessageResponse{RunID: runID})
}

func (h *Handler) stopRun(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	runID := chi.URLParam(r, "id")
	if err := h.runner.Stop(r.Context(), identity.UserID, runID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpjson.WriteError(w, troncerrors.NotFound("run not found"))
			return
		}
		httpjson.WriteError(w, troncerrors.Internal("failed to stop run", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, EmptyResponse{})
}

func (h *Handler) decideApproval(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	runID := chi.URLParam(r, "id")
	callID := chi.URLParam(r, "call")
	var body DecideRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if err := h.runner.Decide(r.Context(), identity.UserID, runID, callID, body.Approve); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			httpjson.WriteError(w, troncerrors.NotFound("run not found"))
		case errors.Is(err, runner.ErrRunNotActive):
			httpjson.WriteError(w, troncerrors.Conflict("run is not active"))
		case errors.Is(err, runner.ErrNoPendingApproval):
			httpjson.WriteError(w, troncerrors.NotFound("no pending approval for this call"))
		default:
			httpjson.WriteError(w, troncerrors.Internal("failed to record approval decision", err))
		}
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, EmptyResponse{})
}

func (h *Handler) streamEvents(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	runID := chi.URLParam(r, "id")
	run, err := h.store.GetRun(r.Context(), identity.UserID, runID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpjson.WriteError(w, troncerrors.NotFound("run not found"))
			return
		}
		httpjson.WriteError(w, troncerrors.Internal("failed to load run", err))
		return
	}

	afterSeq := 0
	if rawID := r.URL.Query().Get("last_event_id"); rawID != "" {
		if parsed, err := strconv.Atoi(rawID); err == nil && parsed > 0 {
			afterSeq = parsed
		}
	}
	if rawHeader := r.Header.Get("Last-Event-ID"); rawHeader != "" && afterSeq == 0 {
		if parsed, err := strconv.Atoi(rawHeader); err == nil && parsed > 0 {
			afterSeq = parsed
		}
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Now().Add(60 * time.Second))
	w.WriteHeader(http.StatusOK)
	_ = rc.Flush()

	notifyCh, unsubscribe := h.runner.Subscribe(runID)
	defer unsubscribe()

	sendEvents := func() (bool, error) {
		events, err := h.store.Events(r.Context(), runID, afterSeq)
		if err != nil {
			return false, err
		}
		for _, ev := range events {
			_ = rc.SetWriteDeadline(time.Now().Add(60 * time.Second))
			if _, err := fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", ev.Seq, ev.Type, string(ev.Payload)); err != nil {
				return false, err
			}
			if err := rc.Flush(); err != nil {
				return false, err
			}
			afterSeq = ev.Seq
			if runner.IsTerminal(ev.Type) {
				return true, nil
			}
		}
		return false, nil
	}

	terminal, err := sendEvents()
	if err != nil || terminal {
		return
	}

	if !h.runner.IsLive(runID) && run.Status != store.RunRunning {
		_, _ = sendEvents()
		return
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-notifyCh:
			term, err := sendEvents()
			if err != nil || term {
				return
			}
			if !h.runner.IsLive(runID) {
				currentRun, err := h.store.GetRun(r.Context(), identity.UserID, runID)
				if err == nil && currentRun.Status != store.RunRunning {
					_, _ = sendEvents()
					return
				}
			}
		case <-ticker.C:
			_ = rc.SetWriteDeadline(time.Now().Add(60 * time.Second))
			if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
				return
			}
			if err := rc.Flush(); err != nil {
				return
			}
			if !h.runner.IsLive(runID) {
				term, err := sendEvents()
				if err != nil || term {
					return
				}
				currentRun, err := h.store.GetRun(r.Context(), identity.UserID, runID)
				if err == nil && currentRun.Status != store.RunRunning {
					return
				}
			}
		}
	}
}
