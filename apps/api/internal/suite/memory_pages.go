package suite

import (
	"context"
	"errors"
	"sort"
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
// the rest by path. Mycelium leaves out hidden files, index.md and log.md.
func (m Mycelium) ListPages(ctx context.Context, userID int64, space string) ([]MemoryPage, error) {
	pages := []MemoryPage{}
	cursor := ""
	for {
		var out struct {
			Pages []struct {
				Path    string `json:"path"`
				Size    int64  `json:"size"`
				Updated string `json:"updated"`
			} `json:"pages"`
			NextCursor string `json:"next_cursor"`
		}
		if err := m.call(ctx, userID, "mycelium_list_memory_pages", map[string]any{"space_id": space, "cursor": cursor}, &out); err != nil {
			return nil, err
		}
		for _, page := range out.Pages {
			if _, err := MemoryPath(page.Path, true); err == nil {
				pages = append(pages, MemoryPage{Path: page.Path, Size: page.Size, Updated: page.Updated})
			}
		}
		if out.NextCursor == "" || out.NextCursor == cursor {
			break
		}
		cursor = out.NextCursor
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

// Delete removes one page from a memory space. Mycelium asks the person
// first: the call's context answers through WithElicit, and without an
// answer the page is kept and Delete says so.
func (m Mycelium) Delete(ctx context.Context, userID int64, space, page string) error {
	var out deleteOutcome
	args := map[string]any{"path": rawPage(page), "log": "bulle: delete " + rawPage(page), "space_id": space}
	if err := m.call(ctx, userID, "mycelium_delete_memory_page", args, &out); err != nil {
		return err
	}
	return out.err()
}

// deleteOutcome is where one of Mycelium's confirmed deletes stands.
type deleteOutcome struct {
	Status     string `json:"status"`
	ConfirmURL string `json:"confirm_url"`
	Message    string `json:"message"`
}

// ErrDeclined is a delete the person did not confirm: nothing changed.
var ErrDeclined = errors.New("the user did not confirm, so nothing was deleted; do not retry")

// PendingError is a delete waiting on Mycelium's confirm page.
type PendingError struct{ URL string }

func (e *PendingError) Error() string {
	return "nothing was deleted yet: the user confirms it at " + e.URL
}

// err is nil once the delete ran.
func (o deleteOutcome) err() error {
	switch o.Status {
	case "deleted":
		return nil
	case "pending_confirmation":
		return &PendingError{URL: o.ConfirmURL}
	}
	return ErrDeclined
}

func memoryDeleteTool(set *toolSet, m Mycelium, resolve MemorySpace) {
	selfTool(set, "memory_delete", "Delete a whole page of the user's long-term memory. The user confirms every delete on a card first.",
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
