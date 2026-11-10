package ai

import (
	"net/http"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/core/jev"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

const actingForHeader = "X-Bulle-Acting-For"

type handler struct {
	cfg Config
}

func (h *handler) decide(w http.ResponseWriter, r *http.Request) {
	identity, err := h.actingFor(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if h.cfg.Access != nil {
		if err := h.cfg.Access.CheckAI(r.Context(), identity.UserID); err != nil {
			httpjson.WriteError(w, err)
			return
		}
	}
	if h.cfg.Decider == nil {
		httpjson.WriteError(w, troncerrors.Unavailable("no decision model is configured"))
		return
	}
	var body DecideRequest
	if err := httpjson.DecodeJSONLimit(w, r, &body, maxBody); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if err := validate(body); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	response, err := h.cfg.Decider.Evaluate(r.Context(), body.State, body.Questions)
	h.audit(callerOf(r), identity.UserID, len(body.Questions), response, err)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Unavailable("the decision model did not answer"))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, answer(response))
}

func (h *handler) actingFor(r *http.Request) (authcontext.Identity, error) {
	email := strings.TrimSpace(r.Header.Get(actingForHeader))
	if email == "" {
		return authcontext.Identity{}, troncerrors.Forbidden(actingForHeader + " must name a Bulle user")
	}
	identity, err := h.cfg.Users.IdentityForEmail(r.Context(), email)
	if troncerrors.Status(err) == http.StatusNotFound {
		return authcontext.Identity{}, troncerrors.Forbidden(actingForHeader + " names no Bulle user")
	}
	return identity, err
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
