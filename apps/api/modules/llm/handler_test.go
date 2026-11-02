package llm

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/go-chi/chi/v5"
)

var errNoKey = errors.New("no key")

type fakeResolver struct {
	keys map[string]string
}

func (f fakeResolver) Resolve(_ context.Context, _ int64, backend string) (string, error) {
	key, ok := f.keys[backend]
	if !ok {
		return "", errNoKey
	}
	return key, nil
}

func fakeAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer bulle-token" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(authcontext.WithIdentity(r.Context(), authcontext.Identity{UserID: 7})))
	})
}

func setup(t *testing.T, upstream http.Handler, keys map[string]string) *httptest.Server {
	t.Helper()
	up := httptest.NewServer(upstream)
	t.Cleanup(up.Close)
	base, _ := url.Parse(up.URL)
	handler := NewHandler(fakeResolver{keys: keys}, errNoKey)
	for name, b := range handler.backends {
		b.base = base
		handler.backends[name] = b
	}
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		r.With(fakeAuth).HandleFunc("/llm/{backend}/*", handler.proxy)
	})
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, srv *httptest.Server, path, token string, hdr map[string]string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, srv.URL+path, strings.NewReader(`{"a":1}`))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestAuthSwap(t *testing.T) {
	cases := map[string]struct{ header, want string }{
		"anthropic":  {"X-Api-Key", "real"},
		"openai":     {"Authorization", "Bearer real"},
		"openrouter": {"Authorization", "Bearer real"},
		"google":     {"X-Goog-Api-Key", "real"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var got *http.Request
			var body string
			srv := setup(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = r
				b, _ := io.ReadAll(r.Body)
				body = string(b)
				w.WriteHeader(http.StatusOK)
			}), map[string]string{name: "real"})
			resp := do(t, srv, "/api/llm/"+name+"/v1/messages?key=leak&x=1", "bulle-token", map[string]string{
				"X-Api-Key":      "bulle-token",
				"X-Goog-Api-Key": "bulle-token",
				"Cookie":         "session=abc",
				"Connection":     "close, X-Hop",
				"X-Hop":          "1",
			})
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status %d", resp.StatusCode)
			}
			if v := got.Header.Get(tc.header); v != tc.want {
				t.Fatalf("%s = %q, want %q", tc.header, v, tc.want)
			}
			if got.URL.Path != "/v1/messages" && !strings.HasSuffix(got.URL.Path, "/v1/messages") {
				t.Fatalf("path %q", got.URL.Path)
			}
			if got.URL.Query().Get("key") != "" || got.URL.Query().Get("x") != "1" {
				t.Fatalf("query %q", got.URL.RawQuery)
			}
			if got.Header.Get("Cookie") != "" || got.Header.Get("X-Hop") != "" {
				t.Fatal("hop or cookie header leaked")
			}
			for h, vs := range got.Header {
				for _, v := range vs {
					if strings.Contains(v, "bulle-token") {
						t.Fatalf("bulle token leaked in %s", h)
					}
				}
			}
			if body != `{"a":1}` {
				t.Fatalf("body %q", body)
			}
		})
	}
}

func TestStreamingFlush(t *testing.T) {
	release := make(chan struct{})
	srv := setup(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte("data: one\n\n"))
		w.(http.Flusher).Flush()
		<-release
		w.Write([]byte("data: two\n\n"))
	}), map[string]string{"anthropic": "real"})
	resp := do(t, srv, "/api/llm/anthropic/v1/messages", "bulle-token", nil)
	buf := make([]byte, len("data: one\n\n"))
	done := make(chan error, 1)
	go func() {
		_, err := io.ReadFull(resp.Body, buf)
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil || string(buf) != "data: one\n\n" {
			t.Fatalf("read %q %v", buf, err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("first chunk not flushed before upstream finished")
	}
	close(release)
	rest, _ := io.ReadAll(resp.Body)
	if string(rest) != "data: two\n\n" {
		t.Fatalf("rest %q", rest)
	}
}

func TestNoKey(t *testing.T) {
	srv := setup(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("upstream must not be called")
	}), map[string]string{})
	resp := do(t, srv, "/api/llm/openai/v1/chat/completions", "bulle-token", nil)
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusConflict || !strings.Contains(string(body), `"error":"no key for openai"`) {
		t.Fatalf("got %d %s", resp.StatusCode, body)
	}
}

func TestBadToken(t *testing.T) {
	srv := setup(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("upstream must not be called")
	}), map[string]string{"openai": "real"})
	resp := do(t, srv, "/api/llm/openai/v1/models", "wrong", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status %d", resp.StatusCode)
	}
}

func TestUnknownBackend(t *testing.T) {
	srv := setup(t, http.NotFoundHandler(), map[string]string{})
	resp := do(t, srv, "/api/llm/nope/x", "bulle-token", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", resp.StatusCode)
	}
}
