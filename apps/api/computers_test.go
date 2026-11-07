package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/computers"
	"github.com/FacileStudio/Bulle/apps/api/internal/middleware"
	computersmodule "github.com/FacileStudio/Bulle/apps/api/modules/computers"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/go-chi/chi/v5"
)

type recordingManager struct {
	users    []int64
	tags     []string
	networks []string
}

func (r *recordingManager) view(userID int64) (*computers.Computer, error) {
	r.users = append(r.users, userID)
	return &computers.Computer{UserID: userID, Status: computers.StatusStopped, CPUs: 2}, nil
}

func (r *recordingManager) Status(_ context.Context, id int64) (*computers.Computer, error) {
	return r.view(id)
}

func (r *recordingManager) Launch(_ context.Context, id int64) (*computers.Computer, error) {
	return r.view(id)
}

func (r *recordingManager) Stop(_ context.Context, id int64) (*computers.Computer, error) {
	return r.view(id)
}

func (r *recordingManager) Snapshot(_ context.Context, id int64, tag string) (*computers.Computer, error) {
	r.tags = append(r.tags, tag)
	return r.view(id)
}

func (r *recordingManager) Rollback(_ context.Context, id int64, tag string) (*computers.Computer, error) {
	r.tags = append(r.tags, tag)
	return r.view(id)
}

func (r *recordingManager) SetNetwork(_ context.Context, id int64, network string) (*computers.Computer, error) {
	r.networks = append(r.networks, network)
	return r.view(id)
}

func computersRouter(t *testing.T) (chi.Router, *session.Manager, *recordingManager) {
	t.Helper()
	manager, err := session.New(porte.Config{SuccessURL: "http://bulle.test/"}, session.Deps{
		Sessions: newMemoryStore(),
		Logger:   slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatal(err)
	}
	recorder := &recordingManager{}
	router := chi.NewRouter()
	guard := middleware.RequireAuth(manager, fixedResolver{})
	router.Route("/api", func(r chi.Router) {
		computersmodule.RegisterRoutes(r, recorder, guard)
	})
	return router, manager, recorder
}

func doJSON(router chi.Router, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestComputerRoutesNeedASession(t *testing.T) {
	router, _, _ := computersRouter(t)
	for _, route := range [][2]string{
		{http.MethodGet, "/api/computer"}, {http.MethodPost, "/api/computer/start"}, {http.MethodPost, "/api/computer/stop"},
		{http.MethodPost, "/api/computer/snapshots"}, {http.MethodPost, "/api/computer/rollback"},
		{http.MethodPut, "/api/computer/network"},
	} {
		if code := doJSON(router, route[0], route[1], `{"tag":"v1"}`, nil).Code; code != http.StatusUnauthorized {
			t.Errorf("%s %s = %d, want 401", route[0], route[1], code)
		}
	}
}

func TestComputerCookieMutationsNeedCSRFAndBearerDoesNot(t *testing.T) {
	router, manager, _ := computersRouter(t)
	token, _, err := manager.Issue(context.Background(), 7, "")
	if err != nil {
		t.Fatal(err)
	}
	cookie := map[string]string{"Cookie": porte.SessionCookieName + "=" + token}
	if code := doJSON(router, http.MethodGet, "/api/computer", "", cookie).Code; code != http.StatusOK {
		t.Fatalf("cookie GET = %d", code)
	}
	if code := doJSON(router, http.MethodPost, "/api/computer/stop", "", cookie).Code; code != http.StatusForbidden {
		t.Fatalf("cookie POST without CSRF = %d", code)
	}
	cookie[porte.CSRFHeaderName] = "1"
	if code := doJSON(router, http.MethodPost, "/api/computer/stop", "", cookie).Code; code != http.StatusOK {
		t.Fatalf("cookie POST with CSRF = %d", code)
	}
	bearer := map[string]string{"Authorization": "Bearer " + token}
	if code := doJSON(router, http.MethodPost, "/api/computer/start", "", bearer).Code; code != http.StatusOK {
		t.Fatalf("bearer POST = %d", code)
	}
}

func TestComputerRoutesActOnTheCallersOwnComputerOnly(t *testing.T) {
	router, manager, recorder := computersRouter(t)
	for _, id := range []int64{7, 8} {
		token, _, err := manager.Issue(context.Background(), id, "")
		if err != nil {
			t.Fatal(err)
		}
		bearer := map[string]string{"Authorization": "Bearer " + token}
		doJSON(router, http.MethodGet, "/api/computer", "", bearer)
		doJSON(router, http.MethodPost, "/api/computer/stop", `{"user_id":99}`, bearer)
	}
	want := []int64{7, 7, 8, 8}
	if len(recorder.users) != len(want) {
		t.Fatalf("users = %v", recorder.users)
	}
	for i, id := range want {
		if recorder.users[i] != id {
			t.Fatalf("users = %v, want %v", recorder.users, want)
		}
	}
}

func TestComputerResponseShape(t *testing.T) {
	router, manager, _ := computersRouter(t)
	token, _, _ := manager.Issue(context.Background(), 7, "")
	response := doJSON(router, http.MethodGet, "/api/computer", "", map[string]string{"Authorization": "Bearer " + token})
	body := response.Body.String()
	for _, want := range []string{`"status":"stopped"`, `"cpus":2`, `"memory_mb":0`, `"disk_gb":0`} {
		if !strings.Contains(body, want) {
			t.Errorf("body %s lacks %s", body, want)
		}
	}
	if strings.Contains(body, "ssh") || strings.Contains(body, "box") {
		t.Errorf("body leaks internals: %s", body)
	}
}

func TestSnapshotBodyIsDecoded(t *testing.T) {
	router, manager, recorder := computersRouter(t)
	token, _, _ := manager.Issue(context.Background(), 7, "")
	response := doJSON(router, http.MethodPost, "/api/computer/snapshots", `{"tag":"v1"}`, map[string]string{"Authorization": "Bearer " + token})
	if response.Code != http.StatusOK || len(recorder.tags) != 1 || recorder.tags[0] != "v1" {
		t.Fatalf("code = %d tags = %v", response.Code, recorder.tags)
	}
}

func TestNetworkBodyIsDecoded(t *testing.T) {
	router, manager, recorder := computersRouter(t)
	token, _, _ := manager.Issue(context.Background(), 7, "")
	bearer := map[string]string{"Authorization": "Bearer " + token}
	response := doJSON(router, http.MethodPut, "/api/computer/network", `{"network":"offline"}`, bearer)
	if response.Code != http.StatusOK || len(recorder.networks) != 1 || recorder.networks[0] != "offline" {
		t.Fatalf("code = %d networks = %v", response.Code, recorder.networks)
	}
}

func TestComputerRoutesAreDocumented(t *testing.T) {
	router, _, _ := computersRouter(t)
	config := apiref.Config{
		Title:    "t",
		Servers:  []string{"/api"},
		Registry: apiref.Registry{Modules: []apiref.Module{computersmodule.Documentation}},
	}
	if missing := apiref.Undocumented(router, config); len(missing) > 0 {
		t.Errorf("undocumented: %v", missing)
	}
}
