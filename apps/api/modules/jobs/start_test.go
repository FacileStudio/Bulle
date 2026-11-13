package jobs_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"iter"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/modules/jobs"
	"github.com/FacileStudio/nacelle"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/go-chi/chi/v5"
)

type refusingAccess struct{}

func (refusingAccess) CheckAIIn(context.Context, int64, string) error {
	return troncerrors.New("payment_required", "AI features need a paid plan.", errors.New("no ai.access"))
}

type failingBackend struct{ fakeBackend }

func (*failingBackend) Stream(context.Context, nacelle.Request) iter.Seq2[nacelle.Event, error] {
	return func(yield func(nacelle.Event, error) bool) {
		yield(nacelle.Event{}, errors.New("provider exploded: key sk-secret"))
	}
}

func runNow(t *testing.T, router chi.Router, st store.Store, prompt string) *httptest.ResponseRecorder {
	t.Helper()
	created, err := st.CreateJob(context.Background(), store.Job{UserID: 1, Name: "digest", Schedule: "0 0 * * *", Prompt: prompt, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/jobs/"+created.ID+"/run", nil))
	return w
}

func conversations(t *testing.T, st store.Store) []store.Session {
	t.Helper()
	sessions, err := st.ListSessions(context.Background(), 1, "")
	if err != nil {
		t.Fatal(err)
	}
	return sessions
}

func TestRunNowWithoutAIAccessLeavesNoConversation(t *testing.T) {
	router, st, _ := setupRouterWithAccess(t, 1, &fakeBackend{}, refusingAccess{})
	w := runNow(t, router, st, "sum up")
	if w.Code != http.StatusPaymentRequired || !bytes.Contains(w.Body.Bytes(), []byte("paid plan")) {
		t.Fatalf("expected 402 with the plan refusal, got %d %s", w.Code, w.Body.String())
	}
	if n := len(conversations(t, st)); n != 0 {
		t.Fatalf("expected no conversation after a refused run, got %d", n)
	}
}

func TestRunNowThatCannotStartRemovesItsConversation(t *testing.T) {
	router, st, _ := setupTestRouter(t, 1)
	w := runNow(t, router, st, "   ")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d %s", w.Code, w.Body.String())
	}
	if n := len(conversations(t, st)); n != 0 {
		t.Fatalf("expected no conversation after a failed start, got %d", n)
	}
}

func waitForStatus(t *testing.T, st store.Store, runID, status string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		run, err := st.GetRun(context.Background(), 1, runID)
		if err != nil {
			t.Fatal(err)
		}
		if run.Status == status {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("run %s never reached %s", runID, status)
}

func lastMessage(t *testing.T, st store.Store, sessionID string) (nacelle.Role, string) {
	t.Helper()
	stored, err := st.Messages(context.Background(), sessionID)
	if err != nil || len(stored) == 0 {
		t.Fatalf("expected messages, got %d (%v)", len(stored), err)
	}
	last, err := runner.DecodeMessage(stored[len(stored)-1].Payload)
	if err != nil {
		t.Fatal(err)
	}
	var text strings.Builder
	for _, part := range last.Parts {
		if p, ok := part.(nacelle.Text); ok {
			text.WriteString(p.Text)
		}
	}
	return last.Role, text.String()
}

func TestRunNowThatFailsExplainsItInItsConversation(t *testing.T) {
	router, st, _ := setupRouterWithBackend(t, 1, &failingBackend{})
	w := runNow(t, router, st, "sum up")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}
	var res jobs.RunJobResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	waitForStatus(t, st, res.RunID, store.RunError)
	sessions := conversations(t, st)
	if len(sessions) != 1 {
		t.Fatalf("expected the run's conversation, got %d", len(sessions))
	}
	role, text := lastMessage(t, st, sessions[0].ID)
	if role != nacelle.RoleAssistant || !strings.Contains(text, "failed") || strings.Contains(text, "sk-secret") {
		t.Fatalf("expected a plain failure note from the assistant, got %s %q", role, text)
	}
}
