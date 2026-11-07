package suite

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

const (
	myceliumSpacesPath = "/api/spaces"
	myceliumSearchPath = "/api/memory/search"
	myceliumFilesPath  = "/api/sync/files/memory/"
	PersonalSpaceName  = "Bulle"
	memoryDefaultHits  = 8
	memoryMaxHits      = 20
)

type MemoryHit struct {
	Path    string `json:"path"`
	Heading string `json:"heading,omitempty"`
	Line    int    `json:"line,omitempty"`
	Excerpt string `json:"excerpt"`
	Date    string `json:"date,omitempty"`
}

type MemoryResults struct {
	Results  []MemoryHit `json:"results"`
	Degraded bool        `json:"degraded,omitempty"`
}

type myceliumSpace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// Mycelium reads and writes the memory pages of one Mycelium space. Paths are
// relative to the space's memory/ directory, the same paths search answers.
type Mycelium struct{ gw *Gateway }

func NewMycelium(gw *Gateway) Mycelium { return Mycelium{gw: gw} }

func (m Mycelium) Search(ctx context.Context, userID int64, space, query string, limit int) (MemoryResults, error) {
	body := map[string]any{"query": query, "space_id": space, "limit": min(clampLimit(limit, memoryDefaultHits), memoryMaxHits)}
	raw, err := m.gw.Do(ctx, userID, Request{App: AppMycelium, Method: http.MethodPost, Path: myceliumSearchPath, Body: body})
	if err != nil {
		return MemoryResults{}, err
	}
	var out MemoryResults
	if err := decodeInto(raw, &out); err != nil {
		return MemoryResults{}, fmt.Errorf("mycelium: reading search results: %w", err)
	}
	if out.Results == nil {
		out.Results = []MemoryHit{}
	}
	return out, nil
}

func (m Mycelium) Read(ctx context.Context, userID int64, space, page string) ([]byte, error) {
	return m.gw.Get(ctx, userID, AppMycelium, myceliumFilesPath+page, url.Values{"space_id": {space}})
}

func (m Mycelium) Write(ctx context.Context, userID int64, space, page, content string) error {
	path := myceliumFilesPath + page + "?" + url.Values{"space_id": {space}}.Encode()
	req := Request{App: AppMycelium, Method: http.MethodPut, Path: path, Raw: []byte(content), ContentType: "text/markdown; charset=utf-8"}
	_, err := m.gw.Do(ctx, userID, req)
	return err
}

// PersonalSpace answers the user's own "Bulle" space, creating it in Mycelium
// the first time and remembering its id. An existing space of that name the
// user owns is adopted rather than duplicated.
func (m Mycelium) PersonalSpace(ctx context.Context, userID int64, saved store.MyceliumSpaces) (string, error) {
	id, err := saved.MyceliumSpace(ctx, userID)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return "", err
	}
	value, err := m.gw.spaces.do(fmt.Sprint("space/", userID), func() (string, error) {
		return m.ensureSpace(context.WithoutCancel(ctx), userID, saved)
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

func (m Mycelium) ensureSpace(ctx context.Context, userID int64, saved store.MyceliumSpaces) (string, error) {
	if id, err := saved.MyceliumSpace(ctx, userID); err == nil {
		return id, nil
	}
	id, err := m.findSpace(ctx, userID)
	if err == nil && id == "" {
		id, err = m.createSpace(ctx, userID)
	}
	if err != nil {
		return "", err
	}
	return id, saved.SaveMyceliumSpace(ctx, userID, id)
}

func (m Mycelium) findSpace(ctx context.Context, userID int64) (string, error) {
	spaces, err := m.spaces(ctx, userID)
	if err != nil {
		return "", err
	}
	for _, space := range spaces {
		if space.Name == PersonalSpaceName && space.Role == "owner" {
			return space.ID, nil
		}
	}
	return "", nil
}

// SpaceNamed answers the id of the user's Mycelium space with this name,
// ignoring case, which is how a shared space is found: Registre gives it the
// same name in every app.
func (m Mycelium) SpaceNamed(ctx context.Context, userID int64, name string) (string, error) {
	spaces, err := m.spaces(ctx, userID)
	if err != nil {
		return "", err
	}
	for _, space := range spaces {
		if strings.EqualFold(space.Name, name) {
			return space.ID, nil
		}
	}
	return "", fmt.Errorf("mycelium: you have no space named %q", name)
}

func (m Mycelium) spaces(ctx context.Context, userID int64) ([]myceliumSpace, error) {
	raw, err := m.gw.Get(ctx, userID, AppMycelium, myceliumSpacesPath, nil)
	if err != nil {
		return nil, err
	}
	var listing struct {
		Spaces []myceliumSpace `json:"spaces"`
	}
	if err := decodeInto(raw, &listing); err != nil {
		return nil, fmt.Errorf("mycelium: reading spaces: %w", err)
	}
	return listing.Spaces, nil
}

func (m Mycelium) createSpace(ctx context.Context, userID int64) (string, error) {
	body := map[string]string{"name": PersonalSpaceName, "description": "What Bulle remembers about you."}
	raw, err := m.gw.Do(ctx, userID, Request{App: AppMycelium, Method: http.MethodPost, Path: myceliumSpacesPath, Body: body})
	if err != nil {
		return "", err
	}
	var created myceliumSpace
	if err := decodeInto(raw, &created); err != nil || created.ID == "" {
		return "", fmt.Errorf("mycelium: the new space has no id")
	}
	return created.ID, nil
}
