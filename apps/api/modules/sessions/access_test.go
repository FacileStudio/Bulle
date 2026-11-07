package sessions

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/plans"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

type denyAI struct{ calls int }

func (d *denyAI) CheckAI(context.Context, int64) error {
	d.calls++
	return troncerrors.New("payment_required", "AI features need a paid plan.", plans.ErrAIAccess).
		WithUpgradeURL("https://sso.example.test/spaces")
}

func TestSendMessageWithoutAIAccess(t *testing.T) {
	deny := &denyAI{}
	router, st, _ := setupRouterAccess(t, &fakeBackend{}, deny)
	if _, err := st.CreateSession(context.Background(), store.Session{ID: "s9", UserID: 1, Title: "t", CreatedAt: time.Now(), UpdatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/sessions/s9/messages", bytes.NewBufferString(`{"text":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusPaymentRequired {
		t.Fatalf("expected 402, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Error struct {
			Code       string `json:"code"`
			UpgradeURL string `json:"upgrade_url"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "payment_required" || body.Error.UpgradeURL != "https://sso.example.test/spaces" {
		t.Fatalf("unexpected envelope: %s", w.Body.String())
	}
	if deny.calls != 1 {
		t.Fatalf("the gate must run once, ran %d times", deny.calls)
	}
	runs, err := st.Messages(context.Background(), "s9")
	if err != nil || len(runs) != 0 {
		t.Fatalf("a refused message must leave no history: %v %d", err, len(runs))
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/sessions", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("non-AI routes must keep working, got %d", w.Code)
	}
}
