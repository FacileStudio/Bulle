package push_test

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	pushsender "github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/FacileStudio/Bulle/apps/api/modules/push"
	"github.com/go-chi/chi/v5"
)

func openStream(t *testing.T, server *httptest.Server, userID int64) *bufio.Reader {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/push/stream", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-User", strconv.FormatInt(userID, 10))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(res.Body)
	if frame := readFrame(t, reader); frame != ": ok\n\n" {
		t.Fatalf("first frame %q", frame)
	}
	return reader
}

func readFrame(t *testing.T, reader *bufio.Reader) string {
	t.Helper()
	frame := ""
	for !strings.HasSuffix(frame, "\n\n") {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		frame += line
	}
	return frame
}

func liveRouter(hub *pushsender.Hub) http.Handler {
	r := chi.NewRouter()
	auth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			id, err := strconv.ParseInt(req.Header.Get("X-User"), 10, 64)
			if err != nil {
				http.Error(w, "no user", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, req.WithContext(authcontext.WithIdentity(req.Context(), authcontext.Identity{UserID: id})))
		})
	}
	push.RegisterRoutes(r, hub, auth)
	return r
}

func TestStreamCarriesOnlyTheCallersEvents(t *testing.T) {
	hub := pushsender.NewHub(pushsender.HubConfig{})
	server := httptest.NewServer(liveRouter(hub))
	t.Cleanup(server.Close)

	mine := openStream(t, server, 1)
	theirs := openStream(t, server, 2)
	hub.Publish(2, "session.created", map[string]string{"session_id": "theirs"})
	hub.Publish(1, "session.created", map[string]string{"session_id": "mine"})

	if frame := readFrame(t, mine); frame != "event: session.created\ndata: {\"session_id\":\"mine\"}\n\n" {
		t.Fatalf("frame %q", frame)
	}
	got := make(chan string, 1)
	go func() {
		line, err := theirs.ReadString('\n')
		if err == nil {
			got <- line
		}
	}()
	select {
	case line := <-got:
		if line != "event: session.created\n" {
			t.Fatalf("the other user's first line %q", line)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the other user got nothing")
	}
}
