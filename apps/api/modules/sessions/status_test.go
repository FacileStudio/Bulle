package sessions

import (
	"context"
	"encoding/json"
	"io"
	"iter"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/nacelle"
	"github.com/go-chi/chi/v5"
)

type blockingBackend struct{ release chan struct{} }

func (b *blockingBackend) Name() string                                                { return "blocking" }
func (b *blockingBackend) Capabilities() nacelle.Capabilities                          { return nacelle.Capabilities{} }
func (b *blockingBackend) CountTokens(context.Context, nacelle.Request) (int64, error) { return 0, nil }

func (b *blockingBackend) Stream(ctx context.Context, _ nacelle.Request) iter.Seq2[nacelle.Event, error] {
	return func(yield func(nacelle.Event, error) bool) {
		select {
		case <-b.release:
		case <-ctx.Done():
		}
		yield(nacelle.Event{Kind: nacelle.KindDone, Stop: nacelle.StopEnd}, nil)
	}
}

func listStatuses(t *testing.T, router http.Handler) map[string]string {
	t.Helper()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sessions", nil))
	var items []SessionSummary
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, item := range items {
		out[item.ID] = item.Status
	}
	return out
}

func TestListShowsWhichSessionsAreRunning(t *testing.T) {
	backend := &blockingBackend{release: make(chan struct{})}
	router, st, rn := setupRouterWith(t, backend)
	busy := createTestSession(t, st, "busy")
	quiet := createTestSession(t, st, "quiet")

	runID, err := rn.Start(context.Background(), 1, busy, "work")
	if err != nil {
		t.Fatal(err)
	}
	if got := listStatuses(t, router); got[busy] != "running" || got[quiet] != "idle" {
		t.Fatalf("statuses while running: %v", got)
	}
	close(backend.release)
	deadline := time.Now().Add(5 * time.Second)
	for rn.IsLive(runID) && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if got := listStatuses(t, router); got[busy] != "idle" {
		t.Fatalf("statuses after the run: %v", got)
	}
}

func createTestSession(t *testing.T, st store.Store, title string) string {
	t.Helper()
	session, err := st.CreateSession(context.Background(), store.Session{ID: store.NewID(), UserID: 1, Title: title})
	if err != nil {
		t.Fatal(err)
	}
	return session.ID
}

func setupRouterWith(t *testing.T, backend nacelle.Backend) (chi.Router, store.Store, *runner.Runner) {
	t.Helper()
	st := store.NewMemory()
	rn := runner.New(runner.Config{
		Store:     st,
		Computers: &fakeComputerProvider{},
		Backends: func(b, m string) (nacelle.Backend, error) {
			return backend, nil
		},
		VMTools: func(target *sandbox.Target) ([]nacelle.Tool, io.Closer, error) {
			return nil, io.NopCloser(nil), nil
		},
		HostTools: func(int64) ([]nacelle.Tool, error) {
			return nil, nil
		},
	})
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		rn.Shutdown(ctx)
	})

	r := chi.NewRouter()
	requireAuth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity := authcontext.Identity{UserID: 1, Email: "user@example.com"}
			next.ServeHTTP(w, req.WithContext(authcontext.WithIdentity(req.Context(), identity)))
		})
	}
	RegisterRoutes(r, st, rn, nil, requireAuth)
	return r, st, rn
}
