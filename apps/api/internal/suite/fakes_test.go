package suite

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type fakePage struct {
	space string
	path  string
}

type fakeApps struct {
	mu      sync.Mutex
	owners  map[string]string
	pages   map[fakePage]string
	creates int
	arts    fakeArtifacts
}

func (f *fakeApps) serve(rw http.ResponseWriter, r *http.Request, user string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.init()
	f.mycelium(rw, r, user)
}

func (f *fakeApps) init() {
	if f.owners == nil {
		f.owners, f.pages = map[string]string{}, map[fakePage]string{}
	}
}

func (f *fakeApps) own(space, user string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.init()
	f.owners[space] = user
}

func (f *fakeApps) mycelium(rw http.ResponseWriter, r *http.Request, user string) {
	switch {
	case r.URL.Path == myceliumSpacesPath && r.Method == http.MethodGet:
		var spaces []myceliumSpace
		for id, owner := range f.owners {
			if owner == user && id != "" {
				spaces = append(spaces, myceliumSpace{ID: id, Name: PersonalSpaceName, Role: "owner"})
			}
		}
		_ = json.NewEncoder(rw).Encode(map[string]any{"spaces": spaces, "common_artifacts": f.owners[""] == user})
	case r.URL.Path == myceliumSpacesPath:
		f.creates++
		id := fmt.Sprintf("sp-%s-%d", user, f.creates)
		f.owners[id] = user
		rw.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(rw).Encode(myceliumSpace{ID: id, Name: PersonalSpaceName, Role: "owner"})
	case r.URL.Path == myceliumSearchPath:
		var req struct {
			Query   string `json:"query"`
			SpaceID string `json:"space_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		f.search(rw, user, req.SpaceID, req.Query)
	case r.URL.Path == myceliumTreePath:
		f.tree(rw, r, user)
	case strings.HasPrefix(r.URL.Path, myceliumArtifactsPath):
		f.artifact(rw, r, user)
	default:
		f.file(rw, r, user)
	}
}

func (f *fakeApps) search(rw http.ResponseWriter, user, space, query string) {
	if f.owners[space] != user {
		rw.WriteHeader(http.StatusForbidden)
		return
	}
	hits := []MemoryHit{}
	for page, body := range f.pages {
		if page.space == space && strings.Contains(body, query) {
			hits = append(hits, MemoryHit{Path: page.path, Excerpt: body})
		}
	}
	_ = json.NewEncoder(rw).Encode(map[string]any{"results": hits, "degraded": true})
}
