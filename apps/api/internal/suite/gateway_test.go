package suite

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGatewayReadsAsTheCaller(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	gw := NewGateway(w.tokens(), w.cfg, nil)
	body, err := gw.Get(context.Background(), 1, AppSablier, "/api/time-entries/running", nil)
	if err != nil || !strings.Contains(string(body), "timer of 1") {
		t.Fatalf("got %q, %v", body, err)
	}
}

func TestGatewayUnknownUserIsNoAccount(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.unknown["1"] = true
	gw := NewGateway(w.tokens(), w.cfg, nil)
	_, err := gw.Get(context.Background(), 1, AppSablier, "/api/time-entries/running", nil)
	var appErr *AppError
	if !errors.As(err, &appErr) || !errors.Is(err, ErrNoAccount) {
		t.Fatalf("want ErrNoAccount, got %v", err)
	}
	if !strings.Contains(err.Error(), w.app.URL) || w.count(&w.exchanges) != 2 {
		t.Fatalf("message must link the app and a 401 must retry once with a fresh token: %v, exchanges %d", err, w.count(&w.exchanges))
	}
}

func TestGatewayAppDown(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	down := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusBadGateway)
	}))
	defer down.Close()
	w.cfg.Apps[AppSablier] = AppConfig{URL: down.URL, Audience: "aud-sablier"}
	_, err := NewGateway(w.tokens(), w.cfg, nil).Get(context.Background(), 1, AppSablier, "/api/time-entries/running", nil)
	if !errors.Is(err, ErrAppDown) {
		t.Fatalf("want ErrAppDown, got %v", err)
	}
}

func TestGatewayForbiddenIsNotAuthFailure(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.cfg.Apps[AppSablier] = AppConfig{URL: w.app.URL, Audience: "wrong"}
	_, err := NewGateway(w.tokens(), w.cfg, nil).Get(context.Background(), 1, AppSablier, "/api/time-entries/running", nil)
	if err == nil || errors.Is(err, ErrNoAccount) || !strings.Contains(err.Error(), "403") {
		t.Fatalf("a refused audience must surface as a 403, got %v", err)
	}
}

func TestStatuses(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.unknown["1"] = true
	w.cfg.Apps[AppMycelium] = AppConfig{}
	statuses := NewGateway(w.tokens(), w.cfg, nil).Statuses(context.Background(), 1)
	if len(statuses) != 1 || statuses[0].App != AppGlouton || statuses[0].State != StateNoAccount {
		t.Fatalf("unexpected statuses: %+v", statuses)
	}
	w.unknown["1"] = false
	if got := NewGateway(w.tokens(), w.cfg, nil).Statuses(context.Background(), 1); got[0].State != StateConnected {
		t.Fatalf("want connected, got %+v", got)
	}
	if got := NewGateway(w.tokens(), w.cfg, nil).Statuses(context.Background(), 99); got[0].State != StateReconnect {
		t.Fatalf("want reconnect for a user who never signed in, got %+v", got)
	}
}
