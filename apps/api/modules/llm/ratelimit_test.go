package llm

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestProxyRateLimitsPerUser(t *testing.T) {
	router := chi.NewRouter()
	passThrough := func(next http.Handler) http.Handler { return next }
	RegisterRoutes(router, fakeResolver{keys: map[string]string{}}, errNoKey, Offer{}, nil, fakeAuth, passThrough)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	for i := 0; i < proxyBurst; i++ {
		if resp := do(t, srv, "/llm/nope/v1/x", "bulle-token", nil); resp.StatusCode == http.StatusTooManyRequests {
			t.Fatalf("request %d was limited inside the burst", i+1)
		}
	}
	if resp := do(t, srv, "/llm/nope/v1/x", "bulle-token", nil); resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("request past the burst answered %d, want 429", resp.StatusCode)
	}
}
