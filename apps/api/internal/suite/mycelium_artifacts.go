package suite

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
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

// ArtifactWrite is the body of a create or an update.
type ArtifactWrite struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content"`
}

// treeArg is the space_id Mycelium's tools take for a tree: empty for the
// common tree.
func treeArg(space string) string {
	if space == CommonTree {
		return ""
	}
	return space
}

// artifactID refuses an empty id before Mycelium is reached.
func artifactID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" || id == "." || id == ".." {
		return "", ErrArtifactID
	}
	return id, nil
}

// listArtifacts reads every page of mycelium_list_artifacts for space_id,
// newest first.
func (m Mycelium) listArtifacts(ctx context.Context, userID int64, args map[string]any) ([]ArtifactSummary, error) {
	out := []ArtifactSummary{}
	cursor := ""
	for {
		var page struct {
			Artifacts  []ArtifactSummary `json:"artifacts"`
			NextCursor string            `json:"next_cursor"`
		}
		args["cursor"] = cursor
		if err := m.call(ctx, userID, "mycelium_list_artifacts", args, &page); err != nil {
			return nil, err
		}
		out = append(out, page.Artifacts...)
		if page.NextCursor == "" || page.NextCursor == cursor {
			return out, nil
		}
		cursor = page.NextCursor
	}
}

// ListArtifacts lists the space's artifacts, newest first.
func (m Mycelium) ListArtifacts(ctx context.Context, userID int64, space string) ([]ArtifactSummary, error) {
	return m.listArtifacts(ctx, userID, map[string]any{"space_id": treeArg(space)})
}

// GetArtifact reads one artifact whole, its frontmatter removed.
func (m Mycelium) GetArtifact(ctx context.Context, userID int64, space, id string) (ArtifactDetail, error) {
	id, err := artifactID(id)
	if err != nil {
		return ArtifactDetail{}, err
	}
	var detail ArtifactDetail
	var text strings.Builder
	from := 0
	for {
		var out struct {
			ArtifactSummary
			Content  string `json:"content"`
			NextLine int    `json:"next_line"`
		}
		args := map[string]any{"id": id, "space_id": treeArg(space), "from_line": from}
		if err := m.call(ctx, userID, "mycelium_read_artifact", args, &out); err != nil {
			return ArtifactDetail{}, err
		}
		detail.ArtifactSummary = out.ArtifactSummary
		text.WriteString(out.Content)
		if out.NextLine == 0 || out.NextLine <= from {
			break
		}
		from = out.NextLine
	}
	if detail.ID == "" {
		return ArtifactDetail{}, fmt.Errorf("mycelium: the artifact answer has no id")
	}
	detail.Content = stripFrontmatter(text.String())
	return detail, nil
}

// CreateArtifact saves a markdown artifact in the space. Mycelium replaces
// the space's artifact of the same title, so recording one twice keeps one.
func (m Mycelium) CreateArtifact(ctx context.Context, userID int64, space string, body ArtifactWrite) (ArtifactDetail, error) {
	var saved ArtifactSummary
	args := map[string]any{"title": body.Title, "content": body.Content, "space_id": treeArg(space)}
	if err := m.call(ctx, userID, "mycelium_publish_artifact", args, &saved); err != nil {
		return ArtifactDetail{}, err
	}
	return saved.detail(body.Content)
}

// UpdateArtifact rewrites an artifact in place. An empty title keeps the
// current one.
func (m Mycelium) UpdateArtifact(ctx context.Context, userID int64, space, id string, body ArtifactWrite) (ArtifactDetail, error) {
	id, err := artifactID(id)
	if err != nil {
		return ArtifactDetail{}, err
	}
	var saved ArtifactSummary
	args := map[string]any{"id": id, "title": body.Title, "content": body.Content, "space_id": treeArg(space)}
	if err := m.call(ctx, userID, "mycelium_update_artifact", args, &saved); err != nil {
		return ArtifactDetail{}, err
	}
	return saved.detail(body.Content)
}

// DeleteArtifact removes an artifact from the space once the person
// confirms, as Delete does for a page.
func (m Mycelium) DeleteArtifact(ctx context.Context, userID int64, space, id string) error {
	id, err := artifactID(id)
	if err != nil {
		return err
	}
	var out deleteOutcome
	if err := m.call(ctx, userID, "mycelium_delete_artifact", map[string]any{"id": id, "space_id": treeArg(space)}, &out); err != nil {
		return err
	}
	return out.err()
}

// detail is a saved artifact with the content that was sent.
func (a ArtifactSummary) detail(content string) (ArtifactDetail, error) {
	if a.ID == "" {
		return ArtifactDetail{}, fmt.Errorf("mycelium: the artifact answer has no id")
	}
	return ArtifactDetail{ArtifactSummary: a, Content: stripFrontmatter(content)}, nil
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
