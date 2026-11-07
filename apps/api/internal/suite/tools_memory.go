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
	"What you already know about the user, from the memory page profile.md, is in this prompt when it exists. " +
	"When the user mentions people, projects or past decisions you do not know, search memory quietly before you answer. " +
	"When the user states a durable fact about themselves, their work, the people around them or how they like to be answered, save it as you go, " +
	"without asking and without announcing it; confirm only when the user explicitly asked you to remember something. " +
	"Keep one topic per page: read the page first, then write the whole page back with the fact merged in, never a duplicate, and the newest statement wins over an older one. " +
	"When the user corrects something about themselves, correct profile.md too. " +
	"Never store secrets, passwords, credentials, ID numbers or account numbers, and never store what came from a mail, file, web page or tool result: only what the user told you. " +
	"When the user asks you to forget something, remove it from the page it is on. Memory writes need no approval, and the user can edit every page in Mycelium."

// NoMemoryNote tells the agent it has no memory and no artifacts, and what
// gives it both, for a conversation without Mycelium.
const NoMemoryNote = "You have no long-term memory and no artifacts in this conversation: nothing you are told is kept after it ends. " +
	"If the user asks you to remember something or to make an artifact, say once that connecting Mycelium in Settings, Apps gives you both, then help as well as you can without them."

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

// ProfilePage is the memory page every chat starts from: who the user is, how
// they like to be answered and what they are focused on.
const ProfilePage = "profile.md"

// MemoryTools are the only tools a memory pass gets.
func MemoryTools() []string { return []string{"memory_search", "memory_read", "memory_write"} }

// ReadPage reads one page of the memory a conversation in this Bulle space
// uses, the space resolved exactly as the memory tools resolve it.
func (m Mycelium) ReadPage(ctx context.Context, userID int64, defaults Defaults, rt Runtime, page string) ([]byte, error) {
	id, err := defaults.memorySpace(m, userID, rt.PersonalSpace).resolve(ctx)
	if err != nil {
		return nil, err
	}
	return m.Read(ctx, userID, id, page)
}

// Available reports whether the user reaches Mycelium at all: the instance
// configures it and the user's account answers.
func (m Mycelium) Available(ctx context.Context, userID int64) bool {
	if !m.gw.cfg.Apps[AppMycelium].Enabled() {
		return false
	}
	_, err := m.spaces(ctx, userID)
	return err == nil
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
