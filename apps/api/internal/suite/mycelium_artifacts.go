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
)

var errArtifactID = errors.New("artifact id is required")

// ArtifactSummary is one artifact as Mycelium lists it.
type ArtifactSummary struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Machine string    `json:"machine"`
	Format  string    `json:"format,omitempty"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

// ArtifactDetail is an artifact with its markdown, frontmatter removed.
type ArtifactDetail struct {
	ArtifactSummary
	Content string `json:"content"`
}

type artifactWrite struct {
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
			return "", errArtifactID
		}
		path += "/" + url.PathEscape(id)
	}
	return path + "?" + url.Values{"space_id": {space}}.Encode(), nil
}

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

func (m Mycelium) GetArtifact(ctx context.Context, userID int64, space, id string) (ArtifactDetail, error) {
	path, err := artifactPath(space, id, true)
	if err != nil {
		return ArtifactDetail{}, err
	}
	return m.artifactCall(ctx, userID, Request{App: AppMycelium, Method: http.MethodGet, Path: path})
}

// CreateArtifact saves a new markdown artifact in the space, signed as Bulle.
func (m Mycelium) CreateArtifact(ctx context.Context, userID int64, space, title, content string) (ArtifactDetail, error) {
	path, _ := artifactPath(space, "", false)
	body := artifactWrite{Title: title, Content: content, Machine: ArtifactMachine}
	return m.artifactCall(ctx, userID, Request{App: AppMycelium, Method: http.MethodPost, Path: path, Body: body})
}

// UpdateArtifact rewrites an artifact in place. An empty title keeps the
// current one.
func (m Mycelium) UpdateArtifact(ctx context.Context, userID int64, space, id, title, content string) (ArtifactDetail, error) {
	path, err := artifactPath(space, id, true)
	if err != nil {
		return ArtifactDetail{}, err
	}
	body := artifactWrite{Title: title, Content: content}
	return m.artifactCall(ctx, userID, Request{App: AppMycelium, Method: http.MethodPut, Path: path, Body: body})
}

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
