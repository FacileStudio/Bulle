package suite

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	saved   []NuageFile
	uploads map[int64][]byte
	arts    fakeArtifacts
}

func (f *fakeApps) serve(rw http.ResponseWriter, r *http.Request, app, user string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.init()
	if app == AppMycelium {
		f.mycelium(rw, r, user)
		return
	}
	f.nuage(rw, r)
}

func (f *fakeApps) init() {
	if f.owners == nil {
		f.owners, f.pages, f.uploads = map[string]string{}, map[fakePage]string{}, map[int64][]byte{}
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

var nuageSeed = map[int64]NuageFile{
	1: {ID: 1, Name: "notes.txt", MimeType: "text/plain", Size: 11},
	2: {ID: 2, Name: "photo.png", MimeType: "image/png", Size: 4},
}

func (f *fakeApps) nuage(rw http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == nuageFoldersPath:
		_ = json.NewEncoder(rw).Encode(map[string]any{"folders": []NuageFolder{{ID: 3, Name: "Docs"}}})
	case path == nuageFilesPath && r.Method == http.MethodPost:
		f.upload(rw, r)
	case path == nuageFilesPath:
		_ = json.NewEncoder(rw).Encode(map[string]any{"files": []NuageFile{nuageSeed[1], nuageSeed[2]}})
	case path == nuageSearchPath:
		_ = json.NewEncoder(rw).Encode(map[string]any{"results": []NuageHit{{ID: 1, Name: "notes.txt", Type: "file", Path: "/notes.txt"}}, "total": 1})
	case strings.HasSuffix(path, "/download"):
		_, _ = io.WriteString(rw, "hello nuage")
	default:
		id, _ := strconv.ParseInt(strings.TrimPrefix(path, nuageFilesPath+"/"), 10, 64)
		_ = json.NewEncoder(rw).Encode(nuageSeed[id])
	}
}

func (f *fakeApps) upload(rw http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil || r.FormValue("origin_app") != "bulle" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	data, _ := io.ReadAll(file)
	folder, _ := strconv.ParseInt(r.FormValue("folder_id"), 10, 64)
	saved := NuageFile{ID: int64(100 + len(f.saved)), Name: header.Filename, MimeType: header.Header.Get("Content-Type"), Size: int64(len(data)), FolderID: &folder}
	f.saved = append(f.saved, saved)
	f.uploads[saved.ID] = data
	rw.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(rw).Encode(saved)
}
