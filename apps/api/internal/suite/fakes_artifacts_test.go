package suite

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type fakeArtifacts struct {
	items  map[string]map[string]ArtifactDetail
	paths  []string
	bodies []ArtifactWrite
}

func (f *fakeApps) artifact(rw http.ResponseWriter, r *http.Request, user string) {
	space := r.URL.Query().Get("space_id")
	f.arts.paths = append(f.arts.paths, r.URL.EscapedPath()+"?"+r.URL.RawQuery)
	if r.URL.Query().Get("all") == "true" {
		f.allArtifacts(rw, user)
		return
	}
	if f.owners[space] != user {
		rw.WriteHeader(http.StatusForbidden)
		return
	}
	if f.arts.items == nil {
		f.arts.items = map[string]map[string]ArtifactDetail{}
	}
	if f.arts.items[space] == nil {
		f.arts.items[space] = map[string]ArtifactDetail{}
	}
	id, one := strings.CutPrefix(r.URL.Path, myceliumArtifactsPath+"/")
	switch {
	case !one && r.Method == http.MethodGet:
		out := []ArtifactSummary{}
		for _, art := range f.arts.items[space] {
			out = append(out, art.ArtifactSummary)
		}
		_ = json.NewEncoder(rw).Encode(out)
	case !one:
		f.saveArtifact(rw, r, space, "")
	default:
		f.artifactOne(rw, r, space, id)
	}
}

func (f *fakeApps) saveArtifact(rw http.ResponseWriter, r *http.Request, space, id string) {
	var body ArtifactWrite
	_ = json.NewDecoder(r.Body).Decode(&body)
	f.arts.bodies = append(f.arts.bodies, body)
	now := time.Date(2026, 11, 6, 9, 0, 0, 0, time.UTC)
	art, found := f.arts.items[space][id]
	status := http.StatusOK
	if !found {
		id, status = fmt.Sprintf("2026-11-06-art-%d", len(f.arts.bodies)), http.StatusCreated
		art = ArtifactDetail{ArtifactSummary: ArtifactSummary{ID: id, Machine: body.Machine, Created: now}}
	}
	if body.Title != "" {
		art.Title = body.Title
	}
	art.Updated = now
	art.Content = "---\ntitle: \"" + art.Title + "\"\ntype: artifact\n---\n\n" + body.Content
	f.arts.items[space][id] = art
	rw.WriteHeader(status)
	_ = json.NewEncoder(rw).Encode(art)
}

func (f *fakeApps) artifactOne(rw http.ResponseWriter, r *http.Request, space, id string) {
	art, found := f.arts.items[space][id]
	switch {
	case !found:
		rw.WriteHeader(http.StatusNotFound)
	case r.Method == http.MethodPut:
		f.saveArtifact(rw, r, space, id)
	case r.Method == http.MethodDelete:
		delete(f.arts.items[space], id)
		rw.WriteHeader(http.StatusNoContent)
	default:
		_ = json.NewEncoder(rw).Encode(art)
	}
}

func (f *fakeApps) allArtifacts(rw http.ResponseWriter, user string) {
	out := []ArtifactSummary{}
	for space, items := range f.arts.items {
		for _, art := range items {
			if f.owners[space] == user {
				art.SpaceID, art.SpaceName = space, "Name of "+space
				out = append(out, art.ArtifactSummary)
			}
		}
	}
	_ = json.NewEncoder(rw).Encode(out)
}
