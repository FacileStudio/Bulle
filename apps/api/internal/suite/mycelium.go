package suite

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

const (
	PersonalSpaceName = "Bulle"
	memoryDefaultHits = 8
	memoryMaxHits     = 20
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

// myceliumTree is one tree Mycelium's index lists: a space, or the common
// tree (empty SpaceID), which only an admin reaches.
type myceliumTree struct {
	SpaceID string `json:"space_id"`
	Name    string `json:"name"`
	Role    string `json:"role"`
}

// Mycelium reads and writes one person's Mycelium memory and artifacts
// through Mycelium's MCP server, with that person's own token. Page paths are
// relative to a space's memory directory, the same paths search answers.
type Mycelium struct{ gw *Gateway }

func NewMycelium(gw *Gateway) Mycelium { return Mycelium{gw: gw} }

// call runs one Mycelium MCP tool as userID.
func (m Mycelium) call(ctx context.Context, userID int64, tool string, args, out any) error {
	return m.gw.MCP().Call(ctx, userID, AppMycelium, tool, args, out)
}

func (m Mycelium) Search(ctx context.Context, userID int64, space, query string, limit int) (MemoryResults, error) {
	args := map[string]any{"query": query, "space_id": space, "limit": min(clampLimit(limit, memoryDefaultHits), memoryMaxHits)}
	var out MemoryResults
	if err := m.call(ctx, userID, "mycelium_search_memory", args, &out); err != nil {
		return MemoryResults{}, err
	}
	if out.Results == nil {
		out.Results = []MemoryHit{}
	}
	return out, nil
}

// Read answers one page whole, frontmatter included, reading every window
// of a long page.
func (m Mycelium) Read(ctx context.Context, userID int64, space, page string) ([]byte, error) {
	var text strings.Builder
	from := 0
	for {
		var out struct {
			Content  string `json:"content"`
			NextLine int    `json:"next_line"`
		}
		args := map[string]any{"path": rawPage(page), "space_id": space, "from_line": from}
		if err := m.call(ctx, userID, "mycelium_read_memory_page", args, &out); err != nil {
			return nil, err
		}
		text.WriteString(out.Content)
		if out.NextLine == 0 || out.NextLine <= from {
			return []byte(text.String()), nil
		}
		from = out.NextLine
	}
}

// Write creates or replaces a page. A page without frontmatter gets the
// header Mycelium's wiki rules ask for, titled after its first heading.
func (m Mycelium) Write(ctx context.Context, userID int64, space, page, content string) error {
	page = rawPage(page)
	args := map[string]any{"path": page, "content": withFrontmatter(page, content), "log": "bulle: update " + page, "space_id": space}
	return m.call(ctx, userID, "mycelium_write_memory_page", args, nil)
}

// PersonalSpace answers the user's own "Bulle" space, creating it in Mycelium
// the first time and remembering its id. Mycelium reuses a space of that
// name the user owns rather than making a second one.
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
	var created struct {
		SpaceID string `json:"space_id"`
	}
	args := map[string]any{"name": PersonalSpaceName, "description": "What Bulle remembers about you."}
	if err := m.call(ctx, userID, "mycelium_create_space", args, &created); err != nil {
		return "", err
	}
	if created.SpaceID == "" {
		return "", fmt.Errorf("mycelium: the new space has no id")
	}
	return created.SpaceID, saved.SaveMyceliumSpace(ctx, userID, created.SpaceID)
}

// SpaceNamed answers the id of the user's Mycelium space with this name,
// ignoring case, which is how a shared space is found: Registre gives it the
// same name in every app.
func (m Mycelium) SpaceNamed(ctx context.Context, userID int64, name string) (string, error) {
	trees, err := m.trees(ctx, userID)
	if err != nil {
		return "", err
	}
	for _, tree := range trees {
		if tree.SpaceID != "" && strings.EqualFold(tree.Name, name) {
			return tree.SpaceID, nil
		}
	}
	return "", fmt.Errorf("mycelium: you have no space named %q", name)
}

// trees lists every tree the user reaches in Mycelium.
func (m Mycelium) trees(ctx context.Context, userID int64) ([]myceliumTree, error) {
	var out struct {
		Spaces []myceliumTree `json:"spaces"`
	}
	err := m.call(ctx, userID, "mycelium_read_memory_index", map[string]any{"space_id": ""}, &out)
	return out.Spaces, err
}

// rawPage undoes MemoryPath's escaping: MCP takes the path as text.
func rawPage(page string) string {
	if raw, err := url.PathUnescape(page); err == nil {
		return raw
	}
	return page
}

// withFrontmatter gives a page the YAML header Mycelium requires when it has
// none: a title from its first heading, else its file name, and a type.
func withFrontmatter(page, content string) string {
	if strings.HasPrefix(content, "---\n") {
		return content
	}
	title := strings.TrimSuffix(page[strings.LastIndex(page, "/")+1:], ".md")
	for _, line := range strings.Split(content, "\n") {
		if heading, ok := strings.CutPrefix(strings.TrimSpace(line), "# "); ok && strings.TrimSpace(heading) != "" {
			title = strings.TrimSpace(heading)
			break
		}
	}
	kind := "synthesis"
	if page == ProfilePage || strings.HasPrefix(page, "people/") {
		kind = "actor"
	}
	return fmt.Sprintf("---\ntitle: %q\ntype: %s\n---\n\n%s", title, kind, strings.TrimLeft(content, "\n"))
}
