package ideas

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/porte"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// AppRunsPerHour caps the runs one app starts for one user per hour.
const AppRunsPerHour = 30

// runWindow is the span AppRunsPerHour counts over.
const runWindow = time.Hour

// runLimiter counts the runs started per user and app inside runWindow.
type runLimiter struct {
	mu     sync.Mutex
	starts map[string][]time.Time
}

func (l *runLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.starts == nil {
		l.starts = map[string][]time.Time{}
	}
	recent := l.starts[key][:0]
	for _, at := range l.starts[key] {
		if now.Sub(at) < runWindow {
			recent = append(recent, at)
		}
	}
	if len(recent) >= AppRunsPerHour {
		l.starts[key] = recent
		return false
	}
	l.starts[key] = append(recent, now)
	return true
}

// callerClient is the OAuth client id the token was exchanged for, which
// Registre writes as the act claim. It identifies the calling app but is not
// its name.
func callerClient(r *http.Request) (string, bool) {
	actor, ok := porte.ActorFrom(r.Context())
	return strings.TrimSpace(actor), ok && strings.TrimSpace(actor) != ""
}

// appRun is a checked request to start a run.
type appRun struct {
	client, app, prompt, title, spaceID string
	page                                runner.PageContext
}

// readAppRun decodes the request and checks it: the token must act for a
// client, the app must be a suite app, the prompt and title within their
// caps, and the page a page of that app. The act claim is a client id, so app
// is a label the first-party caller chooses, not an identity.
func readAppRun(w http.ResponseWriter, r *http.Request) (appRun, error) {
	var body StartAppRunRequest
	if err := httpjson.DecodeJSON(w, r, &body); err != nil {
		return appRun{}, err
	}
	client, acting := callerClient(r)
	if !acting {
		return appRun{}, troncerrors.Forbidden("a token exchanged by an app for its user is required")
	}
	app := strings.ToLower(strings.TrimSpace(body.App))
	run := appRun{client: client, app: app, prompt: strings.TrimSpace(body.Prompt), title: strings.TrimSpace(body.Title), spaceID: body.SpaceID, page: body.Page}
	return run, validAppRun(app, run.prompt, run.title, run.page)
}

// startAppRun starts a run from the prompt a suite app sends for its
// signed-in user. The text is the owner's own message, typed or a fixed
// template of the app, so the run is a normal chat run with the normal
// approval gate and is not marked as coming from an idea.
func (h *Handler) startAppRun(w http.ResponseWriter, r *http.Request) {
	userID, ok := userOf(w, r)
	if !ok {
		return
	}
	run, err := readAppRun(w, r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	space, err := h.space(r.Context(), userID, run.spaceID)
	if err != nil {
		httpjson.WriteError(w, storeError("space", err))
		return
	}
	if !h.runs.allow(run.client+"\x00"+strconv.FormatInt(userID, 10), h.now()) {
		httpjson.WriteError(w, troncerrors.RateLimited("at most 30 runs per hour for this app and user"))
		return
	}
	response, err := h.launchAppRun(r, userID, space.ID, run)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusAccepted, response)
}

func (h *Handler) launchAppRun(r *http.Request, userID int64, spaceID string, run appRun) (RunStartResponse, error) {
	title := run.title
	if title == "" {
		title = firstRunes(run.prompt, store.IdeaTitleLimit)
	}
	session, err := h.runner.CreateWebSession(r.Context(), userID, spaceID, title)
	if err != nil {
		return RunStartResponse{}, runError(err)
	}
	runID, err := h.runner.StartPrompt(r.Context(), userID, session.ID, runner.Prompt{Text: run.prompt, Page: run.page})
	if err != nil {
		if deleted := h.store.DeleteSession(context.WithoutCancel(r.Context()), userID, session.ID); deleted != nil {
			h.logger.Warn("deleting the session of a run that did not start", "error", deleted)
		}
		return RunStartResponse{}, appRunError(err)
	}
	return RunStartResponse{RunID: runID, SessionID: session.ID}, nil
}

func validAppRun(app, prompt, title string, page runner.PageContext) error {
	switch {
	case app == store.IdeaAppBulle || !slices.Contains(store.IdeaApps(), app):
		return troncerrors.Invalid("app must be one of the suite apps")
	case prompt == "":
		return troncerrors.Invalid("prompt cannot be empty")
	case utf8.RuneCountInString(prompt) > store.IdeaPromptLimit:
		return troncerrors.Invalid("prompt is longer than 2000 characters")
	case utf8.RuneCountInString(title) > store.IdeaTitleLimit:
		return troncerrors.Invalid("title is longer than 120 characters")
	case !page.IsZero() && strings.ToLower(strings.TrimSpace(page.App)) != app:
		return troncerrors.Invalid("page.app must be the calling app")
	}
	if !page.IsZero() {
		if _, err := page.Note(); err != nil {
			return appRunError(err)
		}
	}
	return nil
}

func appRunError(err error) error {
	switch {
	case errors.Is(err, runner.ErrBadPage):
		return troncerrors.Invalid("page must name an agenda, vision, docs or feuilles page with an id and an https url")
	case errors.Is(err, runner.ErrEmptyMessage):
		return troncerrors.Invalid("prompt cannot be empty")
	}
	return runError(err)
}

func firstRunes(text string, limit int) string {
	runes := []rune(strings.Join(strings.Fields(text), " "))
	if len(runes) > limit {
		runes = runes[:limit]
	}
	return string(runes)
}
