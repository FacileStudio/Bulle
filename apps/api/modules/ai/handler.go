package ai

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/aiusage"
	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/core/jev"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
)

type handler struct {
	cfg     Config
	limiter *troncmiddleware.Limiter
}

type actingKey struct{}

func actingOf(r *http.Request) authcontext.Identity {
	identity, _ := r.Context().Value(actingKey{}).(authcontext.Identity)
	return identity
}

func (h *handler) decide(w http.ResponseWriter, r *http.Request) {
	identity := actingOf(r)
	body, ok := h.readDecide(w, r, identity)
	if !ok {
		return
	}
	if len(body.Judges) == 2 {
		h.decideTwice(w, r, identity, body)
		return
	}
	response, err := h.cfg.Decider.Evaluate(r.Context(), body.State, body.Questions)
	h.cfg.record(r.Context(), identity.UserID, aiusage.SourceDecide, callerOf(r), decided(response))
	h.audit(callerOf(r), identity.UserID, len(body.Questions), response, err)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Unavailable("the decision model did not answer"))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, answer(response))
}

// readDecide decodes and checks a decide request, and answers the refusal
// itself when there is one: a bad body, no judge configured for it, or the
// user's budget used up.
func (h *handler) readDecide(w http.ResponseWriter, r *http.Request, identity authcontext.Identity) (DecideRequest, bool) {
	var body DecideRequest
	err := httpjson.DecodeJSONLimit(w, r, &body, maxBody)
	if err == nil {
		err = validate(body)
	}
	if err == nil && !h.judged(body.Judges) {
		err = troncerrors.Unavailable("no decision model is configured")
	}
	if err == nil {
		err = h.cfg.overBudget(r.Context(), identity.UserID)
	}
	if err != nil {
		httpjson.WriteError(w, err)
	}
	return body, err == nil
}

// judged reports whether the server has at least one judge for the request.
func (h *handler) judged(judges []string) bool {
	return h.cfg.Decider != nil || len(judges) == 2 && h.cfg.Clef != nil
}

func (h *handler) actingFor(r *http.Request) (authcontext.Identity, error) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		return authcontext.Identity{}, troncerrors.Unauthorized("a token exchanged by an app for its user is required")
	}
	return identity, nil
}

func (h *handler) audit(caller string, userID int64, asked int, response jev.Response, err error) {
	attrs := []any{
		"caller", caller, "user", userID, "questions", asked, "model", response.Model,
		"input_tokens", response.Usage.InputTokens, "output_tokens", response.Usage.OutputTokens, "cost", response.Usage.Cost,
	}
	if err != nil {
		h.cfg.Logger.Warn("ai decide", append(attrs, "error", err)...)
		return
	}
	h.cfg.Logger.Info("ai decide", attrs...)
}

func validate(body DecideRequest) error {
	if len(body.State) == 0 || string(body.State) == "null" {
		return troncerrors.Invalid("state is required")
	}
	if len(body.State) > maxState {
		return troncerrors.TooLarge("state is larger than 32 KiB")
	}
	if n := len(body.Questions); n == 0 || n > maxQuestions {
		return troncerrors.Invalid("ask between 1 and 8 questions")
	}
	if !validJudges(body.Judges) {
		return troncerrors.Invalid(`judges must be ["jev"] or ["jev","clef"]`)
	}
	for name, question := range body.Questions {
		if name == "" || question.Type != "choice" || len(question.Options) < 2 || len(question.Options) > maxOptions {
			return troncerrors.Invalid("every question needs a name, the choice type and 2 to 64 options")
		}
	}
	return nil
}

func answer(response jev.Response) DecideResponse {
	answers := make(map[string]Answer, len(response.Answers))
	for name, got := range response.Answers {
		answers[name] = Answer{Choice: got.Choice, Confidence: got.Confidence, Probabilities: got.Probabilities}
	}
	return DecideResponse{Answers: answers, Model: response.Model, Usage: response.Usage}
}

func (h *handler) auditSummary(caller string, userID int64, sources int, model string, err error) {
	attrs := []any{"caller", caller, "user", userID, "sources", sources, "model", model}
	if err != nil {
		h.cfg.Logger.Warn("ai summarize", append(attrs, "error", err)...)
		return
	}
	h.cfg.Logger.Info("ai summarize", attrs...)
}
