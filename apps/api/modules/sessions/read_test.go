package sessions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func listUnread(t *testing.T, router http.Handler) map[string]bool {
	t.Helper()
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sessions", nil))
	var items []SessionSummary
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("list: %d %s", w.Code, w.Body.String())
	}
	out := map[string]bool{}
	for _, item := range items {
		out[item.ID] = item.Unread
	}
	return out
}

func TestAJobConversationStaysUnreadUntilOpened(t *testing.T) {
	router, st, _ := setupTestRouter(t)
	ctx := context.Background()
	for id, channel := range map[string]string{"job": store.ChannelCron, "chat": store.ChannelWeb} {
		if _, err := st.CreateSession(ctx, store.Session{ID: id, UserID: 1, Channel: channel}); err != nil {
			t.Fatal(err)
		}
	}
	if got := listUnread(t, router); !got["job"] || got["chat"] {
		t.Fatalf("unread before opening: %v", got)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/sessions/job/read", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("mark read: %d %s", w.Code, w.Body.String())
	}
	if got := listUnread(t, router); got["job"] {
		t.Fatalf("still unread after opening: %v", got)
	}

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/sessions/missing/read", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing session: %d", w.Code)
	}
}
