package push_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	pushsender "github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/FacileStudio/Bulle/apps/api/modules/push"
	"github.com/go-chi/chi/v5"
)

func router(sender *pushsender.Sender) chi.Router {
	r := chi.NewRouter()
	auth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := authcontext.WithIdentity(req.Context(), authcontext.Identity{UserID: 7})
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
	push.RegisterRoutes(r, pushsender.NewHub(pushsender.HubConfig{Ntfy: sender}), auth)
	return r
}

func TestDisabledPush(t *testing.T) {
	r := router(nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/push", nil))
	var status push.StatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &status); err != nil || status.Enabled || rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/push/test", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("the desktop stream is always on: want 204, got %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/push/devices", strings.NewReader(`{"kind":"fcm","token":"t"}`)))
	if rec.Code != http.StatusConflict {
		t.Fatalf("fcm device without FCM: want 409, got %d", rec.Code)
	}
}

func TestEnabledPush(t *testing.T) {
	hits := 0
	ntfy := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits++ }))
	defer ntfy.Close()
	sender := pushsender.New(pushsender.Config{URL: ntfy.URL, TopicSecret: "s"}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	r := router(sender)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/push", nil))
	var status push.StatusResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &status)
	if !status.Enabled || status.Topic != sender.Topic(7) || status.SubscribeURL != ntfy.URL+"/"+status.Topic || status.Server != ntfy.URL {
		t.Fatalf("unexpected status %+v", status)
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/push/test", nil))
	if rec.Code != http.StatusNoContent || hits != 1 {
		t.Fatalf("code %d hits %d", rec.Code, hits)
	}
}
