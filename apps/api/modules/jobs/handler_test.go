package jobs_test

import (
	"bytes"
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
	"github.com/FacileStudio/Bulle/apps/api/modules/jobs"
	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/nacelle"
	"github.com/go-chi/chi/v5"
)

type fakeComputerProvider struct{}

func (f *fakeComputerProvider) Ensure(_ context.Context, _ string) (*sandbox.Target, error) {
	return &sandbox.Target{Name: "test-box", Status: "running", Backend: "boite", User: "boite"}, nil
}

func (f *fakeComputerProvider) Touch(string) {}

type fakeBackend struct{}

func (f *fakeBackend) Name() string { return "fake" }

func (f *fakeBackend) Capabilities() nacelle.Capabilities {
	return nacelle.Capabilities{}
}

func (f *fakeBackend) CountTokens(context.Context, nacelle.Request) (int64, error) { return 0, nil }

func (f *fakeBackend) Stream(_ context.Context, _ nacelle.Request) iter.Seq2[nacelle.Event, error] {
	return func(yield func(nacelle.Event, error) bool) {
		if !yield(nacelle.Event{Kind: nacelle.KindText, Text: "Done"}, nil) {
			return
		}
		yield(nacelle.Event{Kind: nacelle.KindDone, Stop: nacelle.StopEnd}, nil)
	}
}

func setupTestRouter(t *testing.T, userID int64) (chi.Router, store.Store, *runner.Runner) {
	t.Helper()
	st := store.NewMemory()
	rn := runner.New(runner.Config{
		Store:     st,
		Computers: &fakeComputerProvider{},
		Backends: func(b, m string) (nacelle.Backend, error) {
			return &fakeBackend{}, nil
		},
		VMTools: func(target *sandbox.Target) ([]nacelle.Tool, io.Closer, error) {
			return nil, io.NopCloser(nil), nil
		},
		HostTools: func() ([]nacelle.Tool, error) {
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
			if userID > 0 {
				identity := authcontext.Identity{UserID: userID, Email: "user@facile.studio"}
				next.ServeHTTP(w, req.WithContext(authcontext.WithIdentity(req.Context(), identity)))
				return
			}
			next.ServeHTTP(w, req)
		})
	}
	jobs.RegisterRoutes(r, st, rn, nil, requireAuth)
	return r, st, rn
}

func TestListAndCreateJobs(t *testing.T) {
	router, _, _ := setupTestRouter(t, 1)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var list []jobs.JobResponse
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil || len(list) != 0 {
		t.Fatalf("expected empty list, got: %v", list)
	}

	body := jobs.CreateJobRequest{
		Name:     "weekly-report",
		Schedule: "0 9 * * 1",
		Prompt:   "compile weekly report",
		Delivery: "push",
	}
	raw, _ := json.Marshal(body)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(raw))
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d %s", w.Code, w.Body.String())
	}

	var created jobs.JobResponse
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if created.ID == "" || created.Name != "weekly-report" || !created.Enabled {
		t.Fatalf("unexpected created job: %+v", created)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/jobs", nil)
	router.ServeHTTP(w, req)
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil || len(list) != 1 {
		t.Fatalf("expected 1 job, got: %v", list)
	}
}

func TestCreateJobValidation(t *testing.T) {
	router, _, _ := setupTestRouter(t, 1)

	cases := []jobs.CreateJobRequest{
		{Name: "", Schedule: "0 9 * * *", Prompt: "test"},
		{Name: "test", Schedule: "", Prompt: "test"},
		{Name: "test", Schedule: "invalid-cron", Prompt: "test"},
		{Name: "test", Schedule: "0 9 * * *", Prompt: ""},
	}
	for i, c := range cases {
		raw, _ := json.Marshal(c)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(raw))
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("case %d: expected 400, got %d %s", i, w.Code, w.Body.String())
		}
	}
}

func TestGetUpdateDeleteJob(t *testing.T) {
	router, st, _ := setupTestRouter(t, 1)

	created, err := st.CreateJob(context.Background(), store.Job{
		UserID:   1,
		Name:     "daily-standup",
		Schedule: "@daily",
		Prompt:   "daily summary",
		Enabled:  true,
	})
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs/"+created.ID, nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var got jobs.JobResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got.ID != created.ID {
		t.Fatalf("got mismatch: %+v", got)
	}

	disabled := false
	updateBody := jobs.UpdateJobRequest{
		Name:     "renamed-standup",
		Schedule: "0 8 * * *",
		Prompt:   "updated prompt",
		Enabled:  &disabled,
	}
	raw, _ := json.Marshal(updateBody)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/jobs/"+created.ID, bytes.NewReader(raw))
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}
	var updated jobs.JobResponse
	if err := json.Unmarshal(w.Body.Bytes(), &updated); err != nil || updated.Name != "renamed-standup" || updated.Enabled {
		t.Fatalf("updated mismatch: %+v", updated)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/jobs/"+created.ID, nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/jobs/"+created.ID, nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRunJobAndListRuns(t *testing.T) {
	router, st, _ := setupTestRouter(t, 1)

	created, err := st.CreateJob(context.Background(), store.Job{
		UserID:   1,
		Name:     "manual-job",
		Schedule: "0 0 * * *",
		Prompt:   "run now",
		Enabled:  true,
	})
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/jobs/"+created.ID+"/run", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}
	var runRes jobs.RunJobResponse
	if err := json.Unmarshal(w.Body.Bytes(), &runRes); err != nil || runRes.RunID == "" || runRes.JobRun.JobID != created.ID {
		t.Fatalf("invalid run response: %+v", runRes)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/jobs/"+created.ID+"/runs", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var runs []jobs.JobRunResponse
	if err := json.Unmarshal(w.Body.Bytes(), &runs); err != nil || len(runs) != 1 {
		t.Fatalf("expected 1 run, got %v", runs)
	}
}

func TestJobsOwnership(t *testing.T) {
	routerStrang, st, _ := setupTestRouter(t, 2)

	created, err := st.CreateJob(context.Background(), store.Job{
		UserID:   1,
		Name:     "user1-job",
		Schedule: "@hourly",
		Prompt:   "ping",
		Enabled:  true,
	})
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/jobs/"+created.ID, nil)
	routerStrang.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("stranger read: expected 404, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/jobs/"+created.ID, nil)
	routerStrang.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("stranger delete: expected 404, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/jobs/"+created.ID+"/run", nil)
	routerStrang.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("stranger run: expected 404, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/jobs/"+created.ID+"/runs", nil)
	routerStrang.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("stranger list runs: expected 404, got %d", w.Code)
	}
}

func TestDeliveryIsValidated(t *testing.T) {
	router, _, _ := setupTestRouter(t, 5)
	send := func(method, path, delivery string) int {
		raw, _ := json.Marshal(jobs.CreateJobRequest{Name: "n", Schedule: "0 9 * * 1", Prompt: "p", Delivery: delivery})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewReader(raw)))
		return w.Code
	}
	if code := send(http.MethodPost, "/jobs", "web"); code != http.StatusBadRequest {
		t.Fatalf("create web: %d", code)
	}
	if code := send(http.MethodPost, "/jobs", ""); code != http.StatusCreated {
		t.Fatalf("create empty: %d", code)
	}
	list := httptest.NewRecorder()
	router.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/jobs", nil))
	var all []jobs.JobResponse
	_ = json.Unmarshal(list.Body.Bytes(), &all)
	if len(all) != 1 {
		t.Fatalf("jobs: %d", len(all))
	}
	if code := send(http.MethodPut, "/jobs/"+all[0].ID, "matrix"); code != http.StatusBadRequest {
		t.Fatalf("update matrix: %d", code)
	}
	if code := send(http.MethodPut, "/jobs/"+all[0].ID, "push"); code != http.StatusOK {
		t.Fatalf("update push: %d", code)
	}
}

func TestUpdateKeepsLegacyDelivery(t *testing.T) {
	router, st, _ := setupTestRouter(t, 5)
	job := store.Job{ID: store.NewID(), UserID: 5, Name: "n", Schedule: "0 9 * * 1", Prompt: "p", Delivery: "antenne", Enabled: true}
	created, err := st.CreateJob(context.Background(), job)
	if err != nil {
		t.Fatalf("seed job: %v", err)
	}
	job = created
	send := func(delivery string) int {
		raw, _ := json.Marshal(jobs.CreateJobRequest{Name: "n", Schedule: "0 9 * * 1", Prompt: "p", Delivery: delivery})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodPut, "/jobs/"+job.ID, bytes.NewReader(raw)))
		return w.Code
	}
	if code := send("antenne"); code != http.StatusOK {
		t.Fatalf("keep antenne: %d", code)
	}
	if code := send("matrix"); code != http.StatusBadRequest {
		t.Fatalf("switch to matrix: %d", code)
	}
}
