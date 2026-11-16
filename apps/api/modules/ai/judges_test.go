package ai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/aiusage"
	"github.com/FacileStudio/Bulle/apps/api/modules/ai"
	"github.com/FacileStudio/Bulle/core/jev"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

const (
	exportOption = "Acme › Billing › Export"
	headerOption = "Acme › Site › Header"
	twoJudges    = `{"state":{"title":"fix the invoice export"},"judges":["jev","clef"],"questions":{"task":{"type":"choice","instructions":"Pick the task","options":["` + exportOption + `","` + headerOption + `"]}}}`
)

type judgeStub struct {
	model      string
	choice     string
	confidence float64
	status     int
	delay      time.Duration
	calls      atomic.Int32
}

func (s *judgeStub) decider(t *testing.T, endpoint jev.Endpoint) ai.Decider {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.calls.Add(1)
		time.Sleep(s.delay)
		if s.status != 0 {
			http.Error(w, "boom", s.status)
			return
		}
		other := 1 - s.confidence
		_, _ = io.WriteString(w, fmt.Sprintf(`{"model":%q,"answers":{"task":{"type":"choice","choice":%q,"confidence":%v,"probabilities":{%q:%v,%q:%v}}},"usage":{"input_tokens":100,"output_tokens":0,"cost":0.001}}`,
			s.model, s.choice, s.confidence, s.choice, s.confidence, "other", other))
	}))
	t.Cleanup(srv.Close)
	return ai.NewDecider(jev.Config{BaseURL: srv.URL, APIKey: "key", Endpoint: endpoint, Model: s.model, Attempts: 1})
}

func withJudges(jevJudge, clefJudge ai.Decider) func(*ai.Config) {
	return func(c *ai.Config) { c.Decider, c.Clef = jevJudge, clefJudge }
}

func decideTwiceBody(t *testing.T, srv *httptest.Server) (int, ai.DecideResponse, string) {
	t.Helper()
	res := post(t, srv, bearer(toileWrite), twoJudges)
	raw, _ := io.ReadAll(res.Body)
	var got ai.DecideResponse
	_ = json.Unmarshal(raw, &got)
	return res.StatusCode, got, string(raw)
}

func TestTwoJudgesThatAgreeAnswerOneChoiceAtTheLowerConfidence(t *testing.T) {
	jevJudge := &judgeStub{model: "jev-1.13.0", choice: exportOption, confidence: 0.91}
	clefJudge := &judgeStub{model: "cloudflare/clef", choice: exportOption, confidence: 0.64}
	ledger := &fakeLedger{}
	srv := completeServer(t, nil, nil, withJudges(jevJudge.decider(t, ""), clefJudge.decider(t, jev.EndpointDecisions)), withLedger(ledger))
	status, got, raw := decideTwiceBody(t, srv)
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, raw)
	}
	task := got.Answers["task"]
	if task.Agree == nil || !*task.Agree || task.Choice != exportOption || task.Confidence != 0.64 {
		t.Fatalf("combined answer: %+v", task)
	}
	if task.Judges["jev"].Confidence != 0.91 || task.Judges["clef"].Confidence != 0.64 || task.Judges["clef"].Choice != exportOption {
		t.Fatalf("each judge's answer is missing: %+v", task.Judges)
	}
	if got.Model != "jev-1.13.0" || got.Judges["clef"].Model != "cloudflare/clef" || got.Usage.InputTokens != 200 || got.Usage.Cost != 0.002 {
		t.Fatalf("model and usage: %+v", got)
	}
	if jevJudge.calls.Load() != 1 || clefJudge.calls.Load() != 1 {
		t.Fatalf("calls: jev %d, clef %d", jevJudge.calls.Load(), clefJudge.calls.Load())
	}
	if len(ledger.rows) != 2 || ledger.rows[0].call.Model == ledger.rows[1].call.Model {
		t.Fatalf("each judge must be recorded on its own: %+v", ledger.rows)
	}
	for _, row := range ledger.rows {
		if row.source != aiusage.SourceDecide || row.userID != 7 || row.caller != "toile-client" || row.call.Usage.Cost != 0.001 {
			t.Fatalf("row %+v", row)
		}
	}
}

func TestTwoJudgesThatDisagreeLeaveTheChoiceEmpty(t *testing.T) {
	jevJudge := &judgeStub{model: "jev-1.13.0", choice: exportOption, confidence: 0.91}
	clefJudge := &judgeStub{model: "cloudflare/clef", choice: headerOption, confidence: 0.55}
	srv := completeServer(t, nil, nil, withJudges(jevJudge.decider(t, ""), clefJudge.decider(t, jev.EndpointDecisions)))
	status, got, raw := decideTwiceBody(t, srv)
	task := got.Answers["task"]
	if status != http.StatusOK || task.Agree == nil || *task.Agree || task.Choice != "" || task.Confidence != 0.55 {
		t.Fatalf("status %d, answer %+v: %s", status, task, raw)
	}
	if task.Judges["jev"].Choice != exportOption || task.Judges["clef"].Choice != headerOption {
		t.Fatalf("each judge keeps its own choice: %+v", task.Judges)
	}
}

func TestAJudgeThatFailsIsNamedAndNeverReplacedByTheOther(t *testing.T) {
	jevJudge := &judgeStub{model: "jev-1.13.0", choice: exportOption, confidence: 0.91}
	clefJudge := &judgeStub{model: "cloudflare/clef", status: http.StatusBadRequest}
	srv := completeServer(t, nil, nil, withJudges(jevJudge.decider(t, ""), clefJudge.decider(t, jev.EndpointDecisions)))
	status, got, raw := decideTwiceBody(t, srv)
	task := got.Answers["task"]
	if status != http.StatusOK || task.Agree == nil || *task.Agree || task.Choice != "" || task.Confidence != 0 {
		t.Fatalf("status %d, answer %+v: %s", status, task, raw)
	}
	if got.Judges["clef"].Error != "did not answer" || got.Judges["jev"].Error != "" || task.Judges["clef"].Error != "did not answer" || task.Judges["jev"].Choice != exportOption {
		t.Fatalf("per-judge errors: %+v / %+v", got.Judges, task.Judges)
	}
	if strings.Contains(raw, "boom") {
		t.Fatalf("a vendor error leaked: %s", raw)
	}
}

func TestClefWithoutAKeyAnswersNotConfigured(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "")
	jevJudge := &judgeStub{model: "jev-1.13.0", choice: exportOption, confidence: 0.91}
	ledger := &fakeLedger{}
	clef := ai.NewClef(jev.Config{Endpoint: jev.EndpointDecisions, Model: "cloudflare/clef"})
	srv := completeServer(t, nil, nil, withJudges(jevJudge.decider(t, ""), clef), withLedger(ledger))
	status, got, raw := decideTwiceBody(t, srv)
	task := got.Answers["task"]
	if status != http.StatusOK || got.Judges["clef"].Error != "not configured" || task.Judges["clef"].Error != "not configured" || *task.Agree || task.Choice != "" {
		t.Fatalf("status %d: %s", status, raw)
	}
	if len(ledger.rows) != 1 {
		t.Fatalf("a judge that was never asked must not be recorded: %+v", ledger.rows)
	}
}

func TestTwoJudgesWithNeitherConfiguredAreUnavailable(t *testing.T) {
	srv := completeServer(t, nil, nil, withJudges(nil, nil))
	if status, _, raw := decideTwiceBody(t, srv); status != http.StatusServiceUnavailable {
		t.Fatalf("status %d: %s", status, raw)
	}
}

func TestJudgesAreValidated(t *testing.T) {
	var calls atomic.Int32
	srv := completeServer(t, nil, nil, withDecider(stubJev(t, &calls)))
	for _, judges := range []string{`["clef"]`, `["clef","jev"]`, `["jev","jev"]`, `["jev","clef","x"]`, `["gpt"]`, `[""]`} {
		body := strings.Replace(question, `"questions"`, `"judges":`+judges+`,"questions"`, 1)
		if res := post(t, srv, bearer(toileWrite), body); res.StatusCode != http.StatusBadRequest {
			t.Errorf("judges %s: status %d, want 400", judges, res.StatusCode)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("a refused call reached jev %d times", calls.Load())
	}
}

func TestSingleJudgeResponseIsByteStable(t *testing.T) {
	const golden = `{"answers":{"task":{"choice":"Acme › Billing › Export","confidence":0.91,"probabilities":{"Acme › Billing › Export":0.91,"Acme › Site › Header":0.09}}},"model":"jev-1.13.0","usage":{"input_tokens":120,"output_tokens":0}}`
	var calls atomic.Int32
	srv := completeServer(t, nil, nil, withDecider(stubJev(t, &calls)))
	for name, body := range map[string]string{
		"no judges":    question,
		"jev named":    strings.Replace(question, `"questions"`, `"judges":["jev"],"questions"`, 1),
		"empty judges": strings.Replace(question, `"questions"`, `"judges":[],"questions"`, 1),
	} {
		res := post(t, srv, bearer(toileWrite), body)
		raw, _ := io.ReadAll(res.Body)
		if res.StatusCode != http.StatusOK || strings.TrimSpace(string(raw)) != golden {
			t.Errorf("%s: status %d, body %s", name, res.StatusCode, raw)
		}
	}
}

func TestTwoJudgesAreCountedTwiceByTheRateLimit(t *testing.T) {
	count := func(body string) int {
		jevJudge := &judgeStub{model: "jev-1.13.0", choice: exportOption, confidence: 0.9}
		clefJudge := &judgeStub{model: "cloudflare/clef", choice: exportOption, confidence: 0.9}
		srv := completeServer(t, nil, nil, withJudges(jevJudge.decider(t, ""), clefJudge.decider(t, jev.EndpointDecisions)))
		admitted := 0
		for range 40 {
			if res := post(t, srv, bearer(toileWrite), body); res.StatusCode == http.StatusOK {
				admitted++
			} else if res.StatusCode != http.StatusTooManyRequests {
				t.Fatalf("status %d", res.StatusCode)
			}
		}
		return admitted
	}
	single := count(question)
	both := count(twoJudges)
	if single < 30 || single > 36 {
		t.Fatalf("single-judge calls admitted: %d, want the burst of 30 plus a little refill", single)
	}
	if both < 15 || both > 18 {
		t.Fatalf("two-judge calls admitted: %d, want half the burst", both)
	}
}

type refusingBudget struct{ checks atomic.Int32 }

func (b *refusingBudget) Check(_ context.Context, _ int64) error {
	b.checks.Add(1)
	return troncerrors.New("rate_limited", "This month's AI allowance is used up. It resets on 1 December.", aiusage.ErrBudget)
}

func TestTwoJudgesAreRefusedByTheBudgetBeforeAnyJudgeIsAsked(t *testing.T) {
	jevJudge := &judgeStub{model: "jev-1.13.0", choice: exportOption, confidence: 0.9}
	clefJudge := &judgeStub{model: "cloudflare/clef", choice: exportOption, confidence: 0.9}
	budget := &refusingBudget{}
	ledger := &fakeLedger{}
	srv := completeServer(t, nil, nil, withJudges(jevJudge.decider(t, ""), clefJudge.decider(t, jev.EndpointDecisions)), withLedger(ledger),
		func(c *ai.Config) { c.Budget = budget })
	if status, _, raw := decideTwiceBody(t, srv); status != http.StatusTooManyRequests {
		t.Fatalf("status %d: %s", status, raw)
	}
	if jevJudge.calls.Load() != 0 || clefJudge.calls.Load() != 0 || len(ledger.rows) != 0 {
		t.Fatalf("a refused call reached a judge or the ledger")
	}
}

func TestTheJudgesAreAskedAtTheSameTime(t *testing.T) {
	jevJudge := &judgeStub{model: "jev-1.13.0", choice: exportOption, confidence: 0.9, delay: 300 * time.Millisecond}
	clefJudge := &judgeStub{model: "cloudflare/clef", choice: exportOption, confidence: 0.9, delay: 300 * time.Millisecond}
	srv := completeServer(t, nil, nil, withJudges(jevJudge.decider(t, ""), clefJudge.decider(t, jev.EndpointDecisions)))
	started := time.Now()
	if status, _, raw := decideTwiceBody(t, srv); status != http.StatusOK {
		t.Fatalf("status %d: %s", status, raw)
	}
	if took := time.Since(started); took >= 550*time.Millisecond {
		t.Fatalf("two 300 ms judges took %s: they ran one after the other", took)
	}
}
