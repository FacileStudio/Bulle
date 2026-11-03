package sessions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func postSession(t *testing.T, router http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/sessions", strings.NewReader(body)))
	return rec
}

func TestSessionsLiveInTheirSpace(t *testing.T) {
	router, st, _ := setupTestRouter(t)
	ctx := context.Background()
	home, _ := st.EnsureDefaultSpace(ctx, 1)
	work, _ := st.CreateSpace(ctx, store.Space{UserID: 1, Name: "Work"})
	theirs, _ := st.CreateSpace(ctx, store.Space{UserID: 2, Name: "Theirs"})

	if rec := postSession(t, router, `{"space_id":"`+theirs.ID+`"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("a session in another user's space = %d, want 404", rec.Code)
	}
	var created CreateSessionResponse
	_ = json.Unmarshal(postSession(t, router, `{"space_id":"`+work.ID+`"}`).Body.Bytes(), &created)
	var fallback CreateSessionResponse
	_ = json.Unmarshal(postSession(t, router, `{}`).Body.Bytes(), &fallback)
	if created.SpaceID != work.ID || fallback.SpaceID != home.ID {
		t.Fatalf("want work then the default space, got %q and %q", created.SpaceID, fallback.SpaceID)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/sessions?space_id="+work.ID, nil))
	var listed []SessionSummary
	_ = json.Unmarshal(rec.Body.Bytes(), &listed)
	if len(listed) != 1 || listed[0].ID != created.ID || listed[0].SpaceID != work.ID {
		t.Fatalf("GET /sessions?space_id= must list only that space: %+v", listed)
	}
}
