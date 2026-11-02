package suite

import (
	"fmt"
	"net/http"
	"strings"
)

func appForPath(path string) string {
	switch {
	case strings.HasPrefix(path, "/api/accounts"):
		return AppCourrier
	case strings.HasPrefix(path, "/api/calendars"):
		return AppAgenda
	case strings.HasPrefix(path, "/api/time-entries"):
		return AppSablier
	}
	return AppOpus
}

func (w *world) serveApp(rw http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), ":")
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
	case audience != "aud-"+app || scope != app+".read":
		rw.WriteHeader(http.StatusForbidden)
	case unknown:
		rw.WriteHeader(http.StatusUnauthorized)
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
