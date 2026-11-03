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
	case strings.HasPrefix(path, "/api/agent/"):
		return AppGlouton
	case strings.HasPrefix(path, "/api/accounts"):
		return AppCourrier
	case strings.HasPrefix(path, "/api/calendars"), strings.HasPrefix(path, "/api/events"):
		return AppAgenda
	case strings.HasPrefix(path, "/api/time-entries"), strings.HasPrefix(path, "/api/agents"), strings.HasPrefix(path, "/api/heartbeats"):
		return AppSablier
	case strings.HasPrefix(path, "/api/spaces"), strings.HasPrefix(path, "/api/memory"), strings.HasPrefix(path, "/api/sync"):
		return AppMycelium
	case strings.HasPrefix(path, "/api/files"), strings.HasPrefix(path, "/api/folders"), strings.HasPrefix(path, "/api/search"):
		return AppNuage
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
	case strings.Contains(r.URL.RawQuery, "from=") && !strings.Contains(r.URL.RawQuery, "from=20"):
		rw.WriteHeader(http.StatusBadRequest)
	case audience != "aud-"+app || !scopeAllows(r, app, scope):
		rw.WriteHeader(http.StatusForbidden)
	case unknown:
		rw.WriteHeader(http.StatusUnauthorized)
	case app == AppMycelium || app == AppNuage:
		w.fakes.serve(rw, r, app, user)
	case app == AppGlouton:
		serveGlouton(rw, r, user)
	default:
		_, _ = fmt.Fprint(rw, appBody(r, user))
	}
}

func appBody(r *http.Request, user string) string {
	path := r.URL.Path
	switch {
	case path == "/api/accounts/":
		return `{"accounts":[{"id":7,"name":"n","email":"e","is_default":true,"imap_host":"hidden"}]}`
	case strings.HasPrefix(path, "/api/accounts/"):
		return fmt.Sprintf(`{"emails":[{"subject":"mail of %s"}]}`, user)
	case strings.HasPrefix(path, "/api/events/") && r.Method == http.MethodGet:
		return `{"id":4,"calendar_id":3,"title":"standup","start_at":"2026-11-04T09:00:00+01:00","end_at":"2026-11-04T09:30:00+01:00","timezone":"Europe/Paris","etag":"hidden"}`
	case strings.HasPrefix(path, "/api/events/"):
		return `{"id":4,"title":"moved"}`
	case path == "/api/calendars/":
		return `[{"id":3,"name":"c","is_personal":true,"role":"owner"}]`
	case strings.HasPrefix(path, "/api/calendars/"):
		return fmt.Sprintf(`[{"id":1,"title":"event of %s","etag":"hidden"}]`, user)
	case strings.HasPrefix(path, "/api/workspace"):
		return `[{"id":2,"name":"W","slug":"w"}]`
	case strings.HasPrefix(path, "/api/project"):
		return `[{"id":5,"name":"P","slug":"p"}]`
	case path == "/api/task/me" && r.URL.Query().Get("includeDone") != "true":
		return fmt.Sprintf(`[{"id":9,"projectId":5,"title":"task of %s","status":"todo","projectName":"P","actorName":"A"}]`, user)
	case path == "/api/task/me":
		return fmt.Sprintf(`[{"id":9,"projectId":5,"title":"task of %s","status":"todo"},{"id":10,"title":"done thing","status":"done"}]`, user)
	case strings.HasPrefix(path, "/api/task/tasks/"):
		return fmt.Sprintf(`{"data":{"columns":[{"isFinal":false,"tasks":[{"id":1,"projectId":5,"title":"task of %s","status":"todo","actorName":"A"}]},{"isFinal":true,"tasks":[{"id":2,"title":"done"}]}]}}`, user)
	}
	return fmt.Sprintf(`{"entry":{"task_name":"timer of %s"}}`, user)
}
