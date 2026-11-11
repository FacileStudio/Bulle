package suite

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type recorded struct {
	method string
	path   string
	body   string
}

func scopeAllows(r *http.Request, app, scope string) bool {
	granted := strings.Split(scope, "+")
	for _, item := range granted {
		if item == app+".write" || (r.Method == http.MethodGet && item == app+".read") {
			return true
		}
	}
	return false
}

func appForPath(path string) string {
	switch {
	case strings.HasPrefix(path, "/api/spaces"), strings.HasPrefix(path, "/api/memory"), strings.HasPrefix(path, "/api/sync"),
		strings.HasPrefix(path, myceliumArtifactsPath):
		return AppMycelium
	}
	return AppSablier
}

func (w *world) serveApp(rw http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), ":")
	if r.Method != http.MethodGet {
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(body))
		w.mu.Lock()
		w.writes = append(w.writes, recorded{method: r.Method, path: r.URL.Path, body: string(body)})
		w.mu.Unlock()
	}
	if len(parts) < 5 || parts[0] != "jwt" {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	user, audience, scope := parts[1], parts[2], parts[3]
	app := appForPath(r.URL.Path)
	w.mu.Lock()
	w.appCalls++
	unknown := w.unknown[user]
	w.mu.Unlock()
	switch {
	case audience != "aud-"+app || !scopeAllows(r, app, scope):
		rw.WriteHeader(http.StatusForbidden)
	case unknown:
		rw.WriteHeader(http.StatusUnauthorized)
	case app == AppMycelium:
		w.fakes.serve(rw, r, user)
	default:
		_, _ = fmt.Fprint(rw, appBody(user))
	}
}

func appBody(user string) string {
	return fmt.Sprintf(`{"entry":{"task_name":"timer of %s"}}`, user)
}
