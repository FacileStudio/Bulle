package suite

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

const maxMemoryPage = 64 << 10

var errNoMemorySpace = errors.New("memory is not available in this conversation")

type memorySearchInput struct {
	Query string `json:"query" jsonschema:"required,description=What to look for in memory"`
	Limit int    `json:"limit,omitempty" jsonschema:"description=Maximum hits (default 8, max 20)"`
}

type memoryReadInput struct {
	Path string `json:"path" jsonschema:"required,description=Page path as memory_search returns it, such as people/alice.md"`
}

type memoryWriteInput struct {
	Path    string `json:"path" jsonschema:"required,description=Page path ending in .md, such as preferences.md or projects/bulle.md"`
	Content string `json:"content" jsonschema:"required,description=The whole page in markdown; it replaces the page"`
}

const memoryNote = "You have a long-term memory (memory_search, memory_read, memory_write) that persists across conversations. " +
	"Before answering about the user, their preferences, people, projects or past decisions, search it. " +
	"When you learn a durable fact worth keeping, save it concisely: read the page first, then write the whole page back with the fact merged in, " +
	"one topic per page, short markdown, no secrets or passwords. Memory writes need no approval, and the user can edit them in Mycelium."

func (space MemorySpace) resolve(ctx context.Context) (string, error) {
	if space == nil {
		return "", errNoMemorySpace
	}
	id, err := space(ctx)
	if err == nil && id == "" {
		return "", errNoMemorySpace
	}
	return id, err
}

func memoryTools(set *toolSet, m Mycelium, space MemorySpace) {
	resolve := space.resolve
	readTool(set, "memory_search", "Search the user's long-term memory. Answers matching pages with an excerpt each.",
		func(ctx context.Context, in memorySearchInput) (string, error) {
			id, err := resolve(ctx)
			if err != nil {
				return "", err
			}
			return jsonResult(m.Search(ctx, set.userID, id, in.Query, in.Limit))
		})
	readTool(set, "memory_read", "Read one page of the user's long-term memory.",
		func(ctx context.Context, in memoryReadInput) (string, error) {
			page, err := memoryPath(in.Path, false)
			if err != nil {
				return "", err
			}
			id, err := resolve(ctx)
			if err != nil {
				return "", err
			}
			return rawResult(m.Read(ctx, set.userID, id, page))
		})
	memoryWriteTool(set, m, resolve)
}

func memoryWriteTool(set *toolSet, m Mycelium, resolve MemorySpace) {
	selfTool(set, "memory_write", "Create or replace a page of the user's long-term memory with markdown.",
		func(ctx context.Context, in memoryWriteInput) (string, error) {
			page, err := memoryPath(in.Path, true)
			if err != nil {
				return "", err
			}
			if strings.TrimSpace(in.Content) == "" || len(in.Content) > maxMemoryPage {
				return "", fmt.Errorf("content must be between 1 byte and %d KB", maxMemoryPage>>10)
			}
			id, err := resolve(ctx)
			if err != nil {
				return "", err
			}
			return done("Saved "+in.Path+" to memory.", m.Write(ctx, set.userID, id, page, in.Content))
		})
}

// memoryPath checks a page path and escapes it for the URL. It must stay
// inside the space's memory directory: no absolute path, no "..", no
// backslash, and a page that is written must be markdown.
func memoryPath(raw string, markdown bool) (string, error) {
	path := strings.TrimSpace(raw)
	switch {
	case path == "":
		return "", errors.New("path is required")
	case strings.HasPrefix(path, "/"), strings.ContainsAny(path, "\\:\x00"):
		return "", fmt.Errorf("path %q must be relative, like notes/topic.md", raw)
	case markdown && !strings.HasSuffix(path, ".md"):
		return "", fmt.Errorf("path %q must end in .md", raw)
	}
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if segment == "" || segment == "." || segment == ".." || strings.HasPrefix(segment, ".") {
			return "", fmt.Errorf("path %q must not contain empty, hidden or .. segments", raw)
		}
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/"), nil
}
