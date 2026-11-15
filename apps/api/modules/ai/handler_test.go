package ai_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/modules/ai"
	"github.com/FacileStudio/Bulle/core/jev"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

const (
	member   = "sara@facile.studio"
	question = `{"state":{"title":"fix the invoice export"},"questions":{"task":{"type":"choice","instructions":"Pick the task","options":["Acme › Billing › Export","Acme › Site › Header"]}}}`
)

func stubJev(t *testing.T, calls *atomic.Int32) ai.Decider {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var body struct {
			State     map[string]string       `json:"state"`
			Questions map[string]jev.Question `json:"questions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.State["title"] != "fix the invoice export" || len(body.Questions["task"].Options) != 2 {
			t.Errorf("jev got an unexpected body: %+v, %v", body, err)
		}
		_, _ = io.WriteString(w, `{"model":"jev-1.13.0","answers":{"task":{"type":"choice","choice":"Acme › Billing › Export","confidence":0.91,`+
			`"probabilities":{"Acme › Billing › Export":0.91,"Acme › Site › Header":0.09}}},"usage":{"input_tokens":120,"output_tokens":0}}`)
	}))
	t.Cleanup(srv.Close)
	return ai.NewDecider(jev.Config{BaseURL: srv.URL, APIKey: "jev-key", Attempts: 1})
}

func withDecider(decider ai.Decider) func(*ai.Config) {
	return func(c *ai.Config) { c.Decider = decider }
}

func bearer(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

func post(t *testing.T, srv *httptest.Server, headers map[string]string, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/ai/decide", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = res.Body.Close() })
	return res
}

func TestDecideRefusesWhatIsNotAnExchangedTokenOrAChoice(t *testing.T) {
	var calls atomic.Int32
	srv := completeServer(t, nil, nil, withDecider(stubJev(t, &calls)))
	for name, tc := range map[string]struct {
		headers map[string]string
		body    string
		want    int
	}{
		"signed out":   {nil, question, http.StatusUnauthorized},
		"no actor":     {bearer(userToken), question, http.StatusForbidden},
		"not a choice": {bearer(toileWrite), `{"state":"x","questions":{"a":{"type":"noul","instructions":"?"}}}`, http.StatusBadRequest},
	} {
		if res := post(t, srv, tc.headers, tc.body); res.StatusCode != tc.want {
			t.Errorf("%s: status %d, want %d", name, res.StatusCode, tc.want)
		}
	}
	if calls.Load() != 0 {
		t.Fatalf("a refused call reached jev %d times", calls.Load())
	}
}

func TestDecideAnswersForAnApp(t *testing.T) {
	var calls atomic.Int32
	srv := completeServer(t, nil, nil, withDecider(stubJev(t, &calls)))
	res := post(t, srv, bearer(toileWrite), question)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200", res.StatusCode)
	}
	var got ai.DecideResponse
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	task := got.Answers["task"]
	if task.Choice != "Acme › Billing › Export" || task.Confidence != 0.91 || task.Probabilities["Acme › Site › Header"] != 0.09 {
		t.Fatalf("unexpected answer: %+v", got)
	}
	if got.Model != "jev-1.13.0" || got.Usage.InputTokens != 120 || calls.Load() != 1 {
		t.Fatalf("unexpected model, usage or call count: %+v, %d calls", got, calls.Load())
	}
}

func TestDecideWithoutAJevKeyIsUnavailable(t *testing.T) {
	t.Setenv("TYPESAFE_API_KEY", "")
	srv := completeServer(t, nil, nil, withDecider(ai.NewDecider(jev.Config{})))
	res := post(t, srv, bearer(toileWrite), question)
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status %d, want 503", res.StatusCode)
	}
}

type deniedPlan struct{}

func (deniedPlan) CheckAIIn(context.Context, int64, string) error {
	return troncerrors.New("payment_required", "upgrade to use AI", nil)
}

func TestDecideRefusesAUserWithoutAIAccess(t *testing.T) {
	var calls atomic.Int32
	srv := completeServer(t, nil, deniedPlan{}, withDecider(stubJev(t, &calls)))
	res := post(t, srv, bearer(toileWrite), question)
	if res.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("status %d, want 402", res.StatusCode)
	}
	if calls.Load() != 0 {
		t.Fatalf("a refused call reached jev %d times", calls.Load())
	}
}
