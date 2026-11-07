package sessions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func search(t *testing.T, router http.Handler, query url.Values) (int, SearchResponse) {
	t.Helper()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/sessions/search?"+query.Encode(), nil))
	var out SearchResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func chatIn(t *testing.T, st store.Store, userID int64, spaceID, channel, said string) string {
	t.Helper()
	ctx := context.Background()
	session, err := st.CreateSession(ctx, store.Session{ID: store.NewID(), UserID: userID, SpaceID: spaceID, Channel: channel})
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(map[string]any{"role": "user", "parts": []map[string]string{{"type": "text", "text": said}}})
	if err := st.ReplaceMessages(ctx, session.ID, []json.RawMessage{payload}); err != nil {
		t.Fatal(err)
	}
	return session.ID
}

func TestSearchIsScopedToTheCallerAndTheSpace(t *testing.T) {
	router, st, _ := setupTestRouter(t)
	ctx := context.Background()
	home, _ := st.EnsureDefaultSpace(ctx, 1)
	work, _ := st.CreateSpace(ctx, store.Space{UserID: 1, Name: "Work"})
	theirs, _ := st.EnsureDefaultSpace(ctx, 2)
	homeChat := chatIn(t, st, 1, home.ID, store.ChannelWeb, "Book the dentist on Friday")
	workChat := chatIn(t, st, 1, work.ID, store.ChannelMatrix, "The dentist invoice is paid")
	chatIn(t, st, 1, home.ID, store.ChannelCron, "dentist reminder")
	chatIn(t, st, 1, home.ID, store.ChannelIdeas, "dentist idea")
	chatIn(t, st, 1, home.ID, store.ChannelMemory, "dentist memory")
	chatIn(t, st, 2, theirs.ID, store.ChannelWeb, "my dentist")

	code, got := search(t, router, url.Values{"q": {"DENTIST"}, "space_id": {work.ID}})
	if code != http.StatusOK || len(got.Results) != 1 || got.Results[0].SessionID != workChat {
		t.Fatalf("search in Work = %d %+v", code, got)
	}
	code, got = search(t, router, url.Values{"q": {"dentist"}})
	if code != http.StatusOK || len(got.Results) != 1 || got.Results[0].SessionID != homeChat || got.Results[0].Snippet != "Book the dentist on Friday" {
		t.Fatalf("search in the default space = %d %+v", code, got)
	}
	if code, _ := search(t, router, url.Values{"q": {"dentist"}, "space_id": {theirs.ID}}); code != http.StatusNotFound {
		t.Fatalf("search in another user's space = %d, want 404", code)
	}
	if code, _ := search(t, router, url.Values{"q": {" d "}}); code != http.StatusBadRequest {
		t.Fatalf("a one-character search = %d, want 400", code)
	}
	if code, got := search(t, router, url.Values{"q": {"nothing like it"}}); code != http.StatusOK || got.Results == nil {
		t.Fatalf("an empty search must answer an empty list: %d %+v", code, got)
	}
}
