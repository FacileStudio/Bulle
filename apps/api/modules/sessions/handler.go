package sessions

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
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
	items, err := h.store.ListSessions(r.Context(), identity.UserID, r.URL.Query().Get("space_id"))
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to list sessions", err))
		return
	}
	statuses := h.runner.Statuses(identity.UserID)
	summaries := make([]SessionSummary, 0, len(items))
	for _, item := range items {
		summaries = append(summaries, summaryOf(item, statuses))
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
	if prefs, err := h.store.GetPreferences(r.Context(), identity.UserID); err == nil {
		if prefs.Model != "" && slices.Contains(h.runner.Models(), prefs.Model) {
			defModel = prefs.Model
		}
	}
	s := store.Session{
		ID:        store.NewID(),
		UserID:    identity.UserID,
		SpaceID:   strings.TrimSpace(body.SpaceID),
		Channel:   store.ChannelWeb,
		Title:     strings.TrimSpace(body.Title),
		Backend:   defBackend,
		Model:     defModel,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	created, err := h.store.CreateSession(r.Context(), s)
	if errors.Is(err, store.ErrNotFound) {
		httpjson.WriteError(w, troncerrors.NotFound("space not found"))
		return
	}
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to create session", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusCreated, CreateSessionResponse{
		ID:      created.ID,
		SpaceID: created.SpaceID,
		Title:   created.Title,
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

	messages := decodeStoredMessages(stored, hasActive)

	res := SessionDetailResponse{
		Session:  summaryOf(session, h.runner.Statuses(identity.UserID)),
		Messages: messages,
	}
	if hasActive {
		res.ActiveRunID = activeRun.ID
	}
	httpjson.WriteJSON(w, http.StatusOK, res)
}

type pendingAssistant struct {
	texts    []string
	thinking []string
	tools    []StoredToolCall
}

func applyToolResults(assistant *pendingAssistant, parts []runner.WirePart) {
	for _, part := range parts {
		if part.Type != "tool_result" {
			continue
		}
		for idx := range assistant.tools {
			if assistant.tools[idx].CallID == part.ID {
				if part.Failed {
					assistant.tools[idx].Status = "failed"
				} else {
					assistant.tools[idx].Status = "ok"
				}
				assistant.tools[idx].OutputPreview = runner.Preview(part.Result)
				break
			}
		}
	}
}

func collectAssistantParts(assistant *pendingAssistant, parts []runner.WirePart) {
	for _, part := range parts {
		switch part.Type {
		case "text":
			if part.Text != "" {
				assistant.texts = append(assistant.texts, part.Text)
			}
		case "reasoning":
			if part.Text != "" {
				assistant.thinking = append(assistant.thinking, part.Text)
			}
		case "tool_call":
			var args any
			if len(part.Input) > 0 {
				if err := json.Unmarshal(part.Input, &args); err != nil {
					args = string(part.Input)
				}
			}
			assistant.tools = append(assistant.tools, StoredToolCall{
				CallID: part.ID,
				Name:   part.Name,
				Args:   args,
				Status: "ok",
			})
		}
	}
}

func decodeStoredMessages(stored []store.Message, hasActive bool) []StoredMessage {
	messages := make([]StoredMessage, 0, len(stored))
	seq := 0
	var assistant *pendingAssistant

	flushAssistant := func() {
		if assistant == nil {
			return
		}
		seq++
		messages = append(messages, StoredMessage{
			Seq:      seq,
			Role:     "assistant",
			Content:  strings.Join(assistant.texts, "\n\n"),
			Answer:   lastText(assistant.texts),
			Thinking: strings.Join(assistant.thinking, "\n\n"),
			Tools:    assistant.tools,
		})
		assistant = nil
	}

	for _, raw := range stored {
		wire, err := runner.DecodeWire(raw.Payload)
		if err != nil {
			continue
		}
		if wire.Role == "user" {
			if assistant != nil {
				applyToolResults(assistant, wire.Parts)
			}
			if users := userMessages(wire.Parts); len(users) > 0 {
				flushAssistant()
				for _, user := range users {
					seq++
					user.Seq = seq
					messages = append(messages, user)
				}
			}
		} else if wire.Role == "assistant" {
			if assistant == nil {
				assistant = &pendingAssistant{}
			}
			collectAssistantParts(assistant, wire.Parts)
		}
	}

	if !hasActive {
		flushAssistant()
	}

	return messages
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
	runID, err := h.runner.StartWithAttachments(r.Context(), identity.UserID, sessionID, body.Text, body.Attachments)
	if err != nil {
		switch {
		case errors.Is(err, runner.ErrEmptyMessage):
			httpjson.WriteError(w, troncerrors.Invalid("message cannot be empty"))
		case errors.Is(err, runner.ErrTooManyAttachments):
			httpjson.WriteError(w, troncerrors.Invalid(fmt.Sprintf("at most %d attachments per message", runner.MaxAttachments)))
		case errors.Is(err, runner.ErrUnknownAttachment):
			httpjson.WriteError(w, troncerrors.Invalid("unknown attachment"))
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
	if err := h.runner.DecideWith(r.Context(), identity.UserID, runID, callID, runner.Decision{Approve: body.Approve, Edit: runner.Edit{As: body.As, Args: body.Args}}); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			httpjson.WriteError(w, troncerrors.NotFound("run not found"))
		case errors.Is(err, runner.ErrRunNotActive):
			httpjson.WriteError(w, troncerrors.Conflict("run is not active"))
		case errors.Is(err, runner.ErrNoPendingApproval):
			httpjson.WriteError(w, troncerrors.NotFound("no pending approval for this call"))
		case errors.Is(err, runner.ErrNotEditable), errors.Is(err, runner.ErrBadEdit):
			httpjson.WriteError(w, troncerrors.Invalid(err.Error()))
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

	notifyCh, unsubscribe := h.runner.Watch(runID)
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
