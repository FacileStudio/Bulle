package suite

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	myceliumArtifactsPath = "/api/artifacts"
	ArtifactMachine       = "bulle"
	// CommonTree names Mycelium's common tree where a space id is expected.
	// Space ids are UUIDs, so it never collides with one.
	CommonTree = "common"
)

var ErrArtifactID = errors.New("artifact id is required")

// ArtifactSummary is one artifact as Mycelium lists it.
type ArtifactSummary struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Machine string    `json:"machine"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	// SpaceID and SpaceName are set by ListAllArtifacts only, which answers
	// CommonTree as the SpaceID of the common tree, an admin's personal tree.
	SpaceID   string `json:"space_id,omitempty"`
	SpaceName string `json:"space_name,omitempty"`
}

// ArtifactDetail is an artifact with its markdown, frontmatter removed.
type ArtifactDetail struct {
	ArtifactSummary
	Content string `json:"content"`
}

// ArtifactWrite is the body of a create or an update. Machine is set by
// CreateArtifact and never sent on an update.
type ArtifactWrite struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content"`
	Machine string `json:"machine,omitempty"`
}

// artifactPath builds the route for one artifact, or the collection when id
// is empty. The id is escaped as one path segment, and "." or ".." is
// refused so it can never walk to another route.
func artifactPath(space, id string, one bool) (string, error) {
	path := myceliumArtifactsPath
	if one {
		id = strings.TrimSpace(id)
		if id == "" || id == "." || id == ".." {
			return "", ErrArtifactID
		}
		path += "/" + url.PathEscape(id)
	}
	if space == CommonTree {
		return path, nil
	}
	return path + "?" + url.Values{"space_id": {space}}.Encode(), nil
}

// ListArtifacts lists the space's artifacts, newest first.
func (m Mycelium) ListArtifacts(ctx context.Context, userID int64, space string) ([]ArtifactSummary, error) {
	path, _ := artifactPath(space, "", false)
	raw, err := m.gw.Do(ctx, userID, Request{App: AppMycelium, Method: http.MethodGet, Path: path})
	if err != nil {
		return nil, err
	}
	out := []ArtifactSummary{}
	if err := decodeInto(raw, &out); err != nil {
		return nil, fmt.Errorf("mycelium: reading artifacts: %w", err)
	}
	return out, nil
}

// GetArtifact reads one artifact, its frontmatter removed.
func (m Mycelium) GetArtifact(ctx context.Context, userID int64, space, id string) (ArtifactDetail, error) {
	path, err := artifactPath(space, id, true)
	if err != nil {
		return ArtifactDetail{}, err
	}
	return m.artifactCall(ctx, userID, Request{App: AppMycelium, Method: http.MethodGet, Path: path})
}

// CreateArtifact saves a new markdown artifact in the space, signed as Bulle.
func (m Mycelium) CreateArtifact(ctx context.Context, userID int64, space string, body ArtifactWrite) (ArtifactDetail, error) {
	path, _ := artifactPath(space, "", false)
	body.Machine = ArtifactMachine
	return m.artifactCall(ctx, userID, Request{App: AppMycelium, Method: http.MethodPost, Path: path, Body: body})
}

// UpdateArtifact rewrites an artifact in place. An empty title keeps the
// current one.
func (m Mycelium) UpdateArtifact(ctx context.Context, userID int64, space, id string, body ArtifactWrite) (ArtifactDetail, error) {
	path, err := artifactPath(space, id, true)
	if err != nil {
		return ArtifactDetail{}, err
	}
	body.Machine = ""
	return m.artifactCall(ctx, userID, Request{App: AppMycelium, Method: http.MethodPut, Path: path, Body: body})
}

// DeleteArtifact removes an artifact from the space.
func (m Mycelium) DeleteArtifact(ctx context.Context, userID int64, space, id string) error {
	path, err := artifactPath(space, id, true)
	if err != nil {
		return err
	}
	_, err = m.gw.Do(ctx, userID, Request{App: AppMycelium, Method: http.MethodDelete, Path: path})
	return err
}

func (m Mycelium) artifactCall(ctx context.Context, userID int64, req Request) (ArtifactDetail, error) {
	raw, err := m.gw.Do(ctx, userID, req)
	if err != nil {
		return ArtifactDetail{}, err
	}
	var out ArtifactDetail
	if err := decodeInto(raw, &out); err != nil || out.ID == "" {
		return ArtifactDetail{}, fmt.Errorf("mycelium: the artifact answer has no id")
	}
	out.Content = stripFrontmatter(out.Content)
	return out, nil
}

// stripFrontmatter drops the leading YAML block Mycelium stamps on every
// markdown artifact, the same cut Mycelium's own reader makes.
func stripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return content
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return content
	}
	rest := content[4+end+1:]
	if cut := strings.Index(rest, "\n"); cut >= 0 {
		rest = rest[cut+1:]
	} else {
		rest = ""
	}
	return strings.TrimLeft(rest, "\r\n")
}
