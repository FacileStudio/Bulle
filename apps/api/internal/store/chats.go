package store

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode"
)

const (
	// MemoryPassLookback bounds how far back a chat may have last changed and
	// still get a memory pass, so a user who never connects Mycelium cannot
	// pile up candidates forever and an old chat is not learned from at boot.
	MemoryPassLookback = 7 * 24 * time.Hour
	memoryPassBatch    = 100
	snippetBefore      = 60
	snippetAfter       = 120
)

// SessionHit is one past chat that matches a search, with a short excerpt
// around the first matching message text. Snippet is empty when only the
// title matched.
type SessionHit struct {
	SessionID string    `json:"session_id"`
	Title     string    `json:"title"`
	Snippet   string    `json:"snippet"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Chats finds the past chats a memory pass should read and searches the
// user's chats. Both only ever see web and matrix sessions: ideas, memory,
// cron and guest sessions are never candidates and never results.
type Chats interface {
	ListSessionsForMemoryPass(ctx context.Context, idleSince time.Time) ([]Session, error)
	MarkMemorySeen(ctx context.Context, sessionID string, seq int) error
	SearchSessions(ctx context.Context, userID int64, spaceID, query string, limit int) ([]SessionHit, error)
}

// Conversational reports whether a channel is a chat a person had with the
// agent, the only sessions memory passes and chat search read.
func Conversational(channel string) bool {
	return channel == ChannelWeb || channel == ChannelMatrix
}

type storedPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type storedMessage struct {
	Role  string       `json:"role"`
	Parts []storedPart `json:"parts"`
}

// ChatTexts reads the user and assistant text parts of a stored message, the
// only parts chat search and memory passes look at. Tool calls, tool results,
// reasoning and attachments are left out.
func ChatTexts(payload json.RawMessage) (role string, texts []string) {
	var message storedMessage
	if json.Unmarshal(payload, &message) != nil || (message.Role != "user" && message.Role != "assistant") {
		return "", nil
	}
	for _, part := range message.Parts {
		if part.Type == "text" && strings.TrimSpace(part.Text) != "" {
			texts = append(texts, part.Text)
		}
	}
	return message.Role, texts
}

// Snippet answers a short excerpt of text around the first case-insensitive
// occurrence of query, and false when query does not occur.
func Snippet(text, query string) (string, bool) {
	runes := []rune(text)
	needle := []rune(strings.TrimSpace(query))
	for i, r := range needle {
		needle[i] = unicode.ToLower(r)
	}
	at := indexFold(runes, needle)
	if at < 0 {
		return "", false
	}
	start := max(at-snippetBefore, 0)
	end := min(at+len(needle)+snippetAfter, len(runes))
	excerpt := strings.Join(strings.Fields(string(runes[start:end])), " ")
	if start > 0 {
		excerpt = "…" + excerpt
	}
	if end < len(runes) {
		excerpt += "…"
	}
	return excerpt, true
}

func indexFold(haystack, needle []rune) int {
	if len(needle) == 0 {
		return -1
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if matchesAt(haystack[i:], needle) {
			return i
		}
	}
	return -1
}

func matchesAt(haystack, needle []rune) bool {
	for j, r := range needle {
		if unicode.ToLower(haystack[j]) != r {
			return false
		}
	}
	return true
}

// FirstSnippet answers the excerpt of the first user or assistant text among
// payloads that contains query.
func FirstSnippet(payloads []json.RawMessage, query string) (string, bool) {
	for _, payload := range payloads {
		_, texts := ChatTexts(payload)
		for _, text := range texts {
			if excerpt, ok := Snippet(text, query); ok {
				return excerpt, true
			}
		}
	}
	return "", false
}

func likePattern(query string) string {
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(query)
	return "%" + escaped + "%"
}
