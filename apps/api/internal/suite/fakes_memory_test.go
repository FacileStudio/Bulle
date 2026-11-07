package suite

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func (f *fakeApps) file(rw http.ResponseWriter, r *http.Request, user string) {
	space := r.URL.Query().Get("space_id")
	path, ok := strings.CutPrefix(r.URL.Path, myceliumFilesPath)
	if !ok || f.owners[space] != user {
		rw.WriteHeader(http.StatusForbidden)
		return
	}
	key := fakePage{space: space, path: path}
	switch r.Method {
	case http.MethodPut:
		body, _ := io.ReadAll(r.Body)
		f.pages[key] = string(body)
		rw.WriteHeader(http.StatusNoContent)
		return
	case http.MethodDelete:
		delete(f.pages, key)
		rw.WriteHeader(http.StatusNoContent)
		return
	}
	body, found := f.pages[key]
	if !found {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	_, _ = io.WriteString(rw, body)
}

func (f *fakeApps) tree(rw http.ResponseWriter, r *http.Request, user string) {
	space := r.URL.Query().Get("space_id")
	if f.owners[space] != user {
		rw.WriteHeader(http.StatusForbidden)
		return
	}
	entries := []map[string]any{
		{"path": "artifacts/a.md", "size": 1, "mod_time": "2026-11-01T00:00:00Z"},
		{"path": "memory/.secret.md", "size": 1, "mod_time": "2026-11-01T00:00:00Z"},
		{"path": "memory/notes.txt", "size": 1, "mod_time": "2026-11-01T00:00:00Z"},
	}
	for page, body := range f.pages {
		if page.space == space {
			entries = append(entries, map[string]any{"path": memoryTreePrefix + page.path, "size": len(body), "mod_time": "2026-11-06T09:00:00Z"})
		}
	}
	_ = json.NewEncoder(rw).Encode(entries)
}
