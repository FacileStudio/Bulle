package suite

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const opusColumnsBody = `[{"slug":"to-do","name":"To Do"},{"slug":"in-progress","name":"In Progress"},` +
	`{"slug":"in-review","name":"In Review"},{"slug":"done","name":"Done","isFinal":true}]`

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
	case strings.HasPrefix(path, "/api/agent/"):
		return AppGlouton
	case strings.HasPrefix(path, "/api/time-entries"), strings.HasPrefix(path, "/api/agents"), strings.HasPrefix(path, "/api/heartbeats"):
		return AppSablier
	case strings.HasPrefix(path, "/api/spaces"), strings.HasPrefix(path, "/api/memory"), strings.HasPrefix(path, "/api/sync"),
		strings.HasPrefix(path, myceliumArtifactsPath):
		return AppMycelium
	}
	return AppOpus
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
	case app == AppGlouton:
		serveGlouton(rw, r, user)
	default:
		_, _ = fmt.Fprint(rw, appBody(r, user))
	}
}

func appBody(r *http.Request, user string) string {
	path := r.URL.Path
	switch {
	case strings.HasPrefix(path, "/api/workspace"):
		return `[{"id":2,"name":"W","slug":"w"},{"id":3,"name":"Other","slug":"other"}]`
	case path == "/api/project" && r.Method == http.MethodGet && r.URL.Query().Get("workspaceId") == "3":
		return `[{"id":8,"name":"Elsewhere","slug":"elsewhere"}]`
	case path == "/api/project" && r.Method == http.MethodGet:
		return `[{"id":5,"name":"P","slug":"p"}]`
	case strings.HasPrefix(path, "/api/project"):
		return `{"id":5,"workspaceId":2,"name":"P","slug":"p","description":"","isPublic":false}`
	case strings.HasPrefix(path, "/api/column/"):
		return opusColumnsBody
	case path == "/api/task/me" && r.URL.Query().Get("includeDone") != "true":
		return fmt.Sprintf(`[{"id":9,"projectId":5,"title":"task of %s","status":"to-do","projectName":"P","actorName":"A"}]`, user)
	case path == "/api/task/me":
		return fmt.Sprintf(`[{"id":9,"projectId":5,"title":"task of %s","status":"to-do"},{"id":10,"projectId":5,"title":"done thing","status":"done"}]`, user)
	case strings.HasPrefix(path, "/api/task/tasks/"):
		return fmt.Sprintf(`{"data":{"columns":[{"isFinal":false,"tasks":[{"id":1,"projectId":5,"title":"task of %s","status":"to-do","actorName":"A"}]},{"isFinal":true,"tasks":[{"id":2,"title":"done"}]}]}}`, user)
	}
	return fmt.Sprintf(`{"entry":{"task_name":"timer of %s"}}`, user)
}
