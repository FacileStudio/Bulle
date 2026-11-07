package runner

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

const (
	// ChatsSearchName is the read-only tool over the user's past chats.
	ChatsSearchName = "chats_search"
	// ChatSearchMin is the shortest query chat search accepts, in characters.
	ChatSearchMin = 2
	// ChatSearchMax is the most chats one search answers.
	ChatSearchMax    = 20
	chatsSearchHits  = 10
	chatsSearchUsage = "Search the user's past chats with you in this space by title and by what the user and you said. " +
		"Use it when the user refers to an earlier conversation. Answers session_id, title, a snippet around the first match and updated_at, newest first."
)

// ErrShortQuery answers a chat search under ChatSearchMin characters.
var ErrShortQuery = errors.New("the query needs at least 2 characters")

type chatsSearchInput struct {
	Query string `json:"query" jsonschema:"required,description=Words to look for in past chats"`
}

// chatsSearchTool is chats_search for one conversation: it only ever sees the
// session owner's web and matrix chats of the session's space.
func (r *Runner) chatsSearchTool(userID int64, session store.Session) []nacelle.Tool {
	if r.cfg.Store == nil {
		return nil
	}
	st := r.cfg.Store
	tool, err := nacelle.NewToolWithOptions(ChatsSearchName, chatsSearchUsage, func(ctx context.Context, in chatsSearchInput) (string, error) {
		hits, err := SearchChats(ctx, st, userID, session.SpaceID, in.Query)
		if err != nil {
			return "", err
		}
		data, err := json.Marshal(hits[:min(len(hits), chatsSearchHits)])
		return string(data), err
	}, nacelle.ToolOptions{ReadOnly: true})
	if err != nil {
		r.cfg.Logger.Error("building chats_search", "error", err)
		return nil
	}
	return []nacelle.Tool{tool}
}

// SearchChats is the one chat search the tool and the HTTP route share: a
// trimmed query of at least two characters, at most ChatSearchMax hits.
func SearchChats(ctx context.Context, st store.Chats, userID int64, spaceID, query string) ([]store.SessionHit, error) {
	query = strings.TrimSpace(query)
	if utf8.RuneCountInString(query) < ChatSearchMin {
		return nil, ErrShortQuery
	}
	hits, err := st.SearchSessions(ctx, userID, spaceID, query, ChatSearchMax)
	if hits == nil && err == nil {
		hits = []store.SessionHit{}
	}
	return hits, err
}
