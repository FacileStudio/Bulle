package sessions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func firstMessage(t *testing.T, router http.Handler, sessionID string) StoredMessage {
	t.Helper()
	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/sessions/"+sessionID, nil))
	var detail SessionDetailResponse
	_ = json.Unmarshal(get.Body.Bytes(), &detail)
	if len(detail.Messages) == 0 {
		t.Fatalf("no messages: %s", get.Body.String())
	}
	return detail.Messages[0]
}

func TestSendAttachmentsWithoutTextAndReadThemBack(t *testing.T) {
	router, st, rn := setupTestRouter(t)
	ctx := context.Background()
	session, _ := st.CreateSession(ctx, store.Session{ID: "s1", UserID: 1, Channel: store.ChannelWeb})
	file, _ := st.CreateAttachment(ctx, store.Attachment{ID: "f1", UserID: 1, Name: "a.png", MediaType: "image/png", Size: 4, Data: []byte("\x89PNG")})
	foreign, _ := st.CreateAttachment(ctx, store.Attachment{ID: "f2", UserID: 2, Name: "b.png", MediaType: "image/png", Size: 4, Data: []byte("\x89PNG")})

	send := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/sessions/"+session.ID+"/messages", strings.NewReader(body)))
		return w
	}
	if w := send(`{"text":"","attachments":["` + foreign.ID + `"]}`); w.Code != http.StatusBadRequest {
		t.Fatalf("foreign attachment: %d", w.Code)
	}
	w := send(`{"text":"","attachments":["` + file.ID + `"]}`)
	if w.Code != http.StatusAccepted {
		t.Fatalf("attachments only: %d %s", w.Code, w.Body.String())
	}
	var sent SendMessageResponse
	_ = json.Unmarshal(w.Body.Bytes(), &sent)
	deadline := time.Now().Add(5 * time.Second)
	for rn.IsLive(sent.RunID) && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	first := firstMessage(t, router, session.ID)
	want := AttachmentInfo{ID: "f1", Name: "a.png", MediaType: "image/png", Size: 4}
	if first.Role != "user" || len(first.Attachments) != 1 || first.Attachments[0] != want {
		t.Fatalf("user message: %+v", first)
	}
}
