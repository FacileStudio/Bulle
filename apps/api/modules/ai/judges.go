package ai

import (
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/aiusage"
	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/core/jev"
	"github.com/FacileStudio/tronc/httpjson"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
)

const (
	errNotConfigured = "not configured"
	errNoAnswer      = "did not answer"
)

var bothJudges = []string{JudgeJev, JudgeClef}

// validJudges accepts no list, ["jev"] and ["jev","clef"]: nothing else names
// a judge set the server can answer.
func validJudges(judges []string) bool {
	return len(judges) == 0 || slices.Equal(judges, []string{JudgeJev}) || slices.Equal(judges, bothJudges)
}

// decisions spends one rate-limit token for the call, per app and acting user.
// A two-judge call spends a second one in decideTwice.
func (h *handler) decisions(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.spend(w, r) {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *handler) spend(w http.ResponseWriter, r *http.Request) bool {
	allowed, wait := h.limiter.AllowAt(callerAndUser(r), time.Now())
	if allowed {
		return true
	}
	w.Header().Set("Retry-After", strconv.Itoa(int(wait.Seconds())))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = w.Write([]byte(troncmiddleware.RateLimitExceededBody))
	return false
}

type judgeRun struct {
	response jev.Response
	err      error
	asked    bool
}

// decideTwice asks Jev and clef the same questions at the same time and
// answers each judge's result, never one standing in for the other. A call
// counts as two against the caller's rate limit, and each judge's usage is
// recorded in the ledger on its own, so it counts as two against the budget.
func (h *handler) decideTwice(w http.ResponseWriter, r *http.Request, identity authcontext.Identity, body DecideRequest) {
	if !h.spend(w, r) {
		return
	}
	runs := map[string]*judgeRun{JudgeJev: {}, JudgeClef: {}}
	deciders := map[string]Decider{JudgeJev: h.cfg.Decider, JudgeClef: h.cfg.Clef}
	var wg sync.WaitGroup
	for name, decider := range deciders {
		if decider == nil {
			continue
		}
		run := runs[name]
		run.asked = true
		wg.Go(func() { run.response, run.err = decider.Evaluate(r.Context(), body.State, body.Questions) })
	}
	wg.Wait()
	caller := callerOf(r)
	for _, name := range bothJudges {
		run := runs[name]
		if !run.asked {
			continue
		}
		h.cfg.record(r.Context(), identity.UserID, aiusage.SourceDecide, caller, decided(run.response))
		h.audit(caller+"/"+name, identity.UserID, len(body.Questions), run.response, run.err)
	}
	httpjson.WriteJSON(w, http.StatusOK, combine(body.Questions, runs))
}

func failure(run *judgeRun) string {
	switch {
	case !run.asked:
		return errNotConfigured
	case run.err != nil:
		return errNoAnswer
	}
	return ""
}

func combine(questions map[string]jev.Question, runs map[string]*judgeRun) DecideResponse {
	calls := make(map[string]JudgeCall, len(runs))
	var usage jev.Usage
	for name, run := range runs {
		call := JudgeCall{Error: failure(run)}
		if call.Error == "" {
			call.Model, call.Usage = run.response.Model, run.response.Usage
		}
		calls[name] = call
		usage.InputTokens += run.response.Usage.InputTokens
		usage.OutputTokens += run.response.Usage.OutputTokens
		usage.Cost += run.response.Usage.Cost
	}
	answers := make(map[string]Answer, len(questions))
	for name := range questions {
		answers[name] = judged(name, runs)
	}
	return DecideResponse{Answers: answers, Model: calls[JudgeJev].Model, Usage: usage, Judges: calls}
}

func judged(question string, runs map[string]*judgeRun) Answer {
	each := make(map[string]JudgeAnswer, len(runs))
	var got []jev.Answer
	for _, name := range bothJudges {
		run := runs[name]
		if reason := failure(run); reason != "" {
			each[name] = JudgeAnswer{Error: reason}
			continue
		}
		answer, ok := run.response.Answers[question]
		if !ok {
			each[name] = JudgeAnswer{Error: errNoAnswer}
			continue
		}
		each[name] = JudgeAnswer{Choice: answer.Choice, Confidence: answer.Confidence, Probabilities: answer.Probabilities}
		got = append(got, answer)
	}
	agree := len(got) == 2 && got[0].Choice != "" && got[0].Choice == got[1].Choice
	out := Answer{Probabilities: map[string]float64{}, Agree: &agree, Judges: each}
	if len(got) == 2 {
		out.Confidence = min(got[0].Confidence, got[1].Confidence)
	}
	if agree {
		out.Choice = got[0].Choice
	}
	return out
}
