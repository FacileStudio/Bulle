package suite

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

const (
	myceliumTreePath = "/api/sync/tree"
	memoryTreePrefix = "memory/"
)

// MemoryPage is one markdown page of a memory space, as the Settings list
// shows it. Path is relative to the space's memory directory.
type MemoryPage struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Updated string `json:"updated"`
}

type memoryDeleteInput struct {
	Path string `json:"path" jsonschema:"required,description=Page path ending in .md, as memory_search returns it, such as people/alice.md"`
}

// ListPages lists the markdown pages of a memory space, profile.md first and
// the rest by path. Hidden files and anything outside memory/ are left out.
func (m Mycelium) ListPages(ctx context.Context, userID int64, space string) ([]MemoryPage, error) {
	raw, err := m.gw.Get(ctx, userID, AppMycelium, myceliumTreePath, url.Values{"space_id": {space}})
	if err != nil {
		return nil, err
	}
	var tree []struct {
		Path    string `json:"path"`
		Size    int64  `json:"size"`
		ModTime string `json:"mod_time"`
	}
	if err := decodeInto(raw, &tree); err != nil {
		return nil, fmt.Errorf("mycelium: reading the file tree: %w", err)
	}
	pages := []MemoryPage{}
	for _, entry := range tree {
		page, ok := strings.CutPrefix(entry.Path, memoryTreePrefix)
		if _, err := MemoryPath(page, true); ok && err == nil {
			pages = append(pages, MemoryPage{Path: page, Size: entry.Size, Updated: entry.ModTime})
		}
	}
	sort.SliceStable(pages, func(i, j int) bool { return pageBefore(pages[i].Path, pages[j].Path) })
	return pages, nil
}

func pageBefore(a, b string) bool {
	if (a == ProfilePage) != (b == ProfilePage) {
		return a == ProfilePage
	}
	return a < b
}

// Delete removes one page from a memory space. The page path must already be
// checked and escaped by MemoryPath.
func (m Mycelium) Delete(ctx context.Context, userID int64, space, page string) error {
	path := myceliumFilesPath + page + "?" + url.Values{"space_id": {space}}.Encode()
	_, err := m.gw.Do(ctx, userID, Request{App: AppMycelium, Method: http.MethodDelete, Path: path})
	return err
}

func memoryDeleteTool(set *toolSet, m Mycelium, resolve MemorySpace) {
	selfTool(set, "memory_delete", "Delete a whole page of the user's long-term memory.",
		func(ctx context.Context, in memoryDeleteInput) (string, error) {
			page, err := MemoryPath(in.Path, true)
			if err != nil {
				return "", err
			}
			id, err := resolve(ctx)
			if err != nil {
				return "", err
			}
			return done("Deleted "+in.Path+" from memory.", m.Delete(ctx, set.userID, id, page))
		})
}
