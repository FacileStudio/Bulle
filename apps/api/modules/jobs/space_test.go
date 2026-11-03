package jobs_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/modules/jobs"
)

func postJob(t *testing.T, router http.Handler, spaceID string) *httptest.ResponseRecorder {
	t.Helper()
	raw, _ := json.Marshal(jobs.CreateJobRequest{SpaceID: spaceID, Name: "n", Schedule: "0 9 * * *", Prompt: "p"})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewReader(raw)))
	return rec
}

func TestJobsLiveInTheirSpace(t *testing.T) {
	router, st, _ := setupTestRouter(t, 1)
	ctx := context.Background()
	home, _ := st.EnsureDefaultSpace(ctx, 1)
	work, _ := st.CreateSpace(ctx, store.Space{UserID: 1, Name: "Work"})
	theirs, _ := st.CreateSpace(ctx, store.Space{UserID: 2, Name: "Theirs"})

	if rec := postJob(t, router, theirs.ID); rec.Code != http.StatusNotFound {
		t.Fatalf("a job in another user's space = %d, want 404", rec.Code)
	}
	var inWork, fallback jobs.JobResponse
	_ = json.Unmarshal(postJob(t, router, work.ID).Body.Bytes(), &inWork)
	_ = json.Unmarshal(postJob(t, router, "").Body.Bytes(), &fallback)
	if inWork.SpaceID != work.ID || fallback.SpaceID != home.ID {
		t.Fatalf("want work then the default space, got %q and %q", inWork.SpaceID, fallback.SpaceID)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/jobs?space_id="+work.ID, nil))
	var listed []jobs.JobResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &listed)
	if len(listed) != 1 || listed[0].ID != inWork.ID {
		t.Fatalf("GET /jobs?space_id= must list only that space: %+v", listed)
	}
}
