package runner

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"unicode"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

const (
	pageNotePrefix = "[Open page: "
	maxPageField   = 300
)

// ErrBadPage is the refusal of a page context that names no embedding app or
// carries no usable id or address.
var ErrBadPage = errors.New("runner: the open page is not a docs, toile or feuilles page")

// pageApps are the apps that embed the chat beside one of their pages.
var pageApps = []string{suite.AppDocs, suite.AppToile, suite.AppFeuilles}

// PageContext is the page a suite app has open beside an embedded chat. The
// zero value means no page.
type PageContext struct {
	App   string `json:"app"`
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// IsZero reports whether no page is given.
func (p PageContext) IsZero() bool { return p == PageContext{} }

// Note is the line the model reads in front of the message: which page the
// user has open and that it starts from it. Every field is folded to one
// line without brackets, so the page cannot end the note early or forge a
// second one.
func (p PageContext) Note() (nacelle.Text, error) {
	app := strings.ToLower(strings.TrimSpace(p.App))
	id, title, address := pageField(p.ID), pageField(p.Title), strings.TrimSpace(p.URL)
	if !slices.Contains(pageApps, app) || id == "" || !pageURL(address) {
		return nacelle.Text{}, ErrBadPage
	}
	label := suite.Label(app)
	text := fmt.Sprintf("%sThe user has %s page %q (%s, %s) open. Start from it: read it with the %s tools before answering. "+
		"Use the suite tools for other pages when asked.]", pageNotePrefix, label, title, id, pageField(address), label)
	return nacelle.Text{Text: text}, nil
}

// IsNote reports whether a stored text part is one the runner wrote for the
// model alone, the clock or the open page, which a person should not see.
func IsNote(text string) bool {
	return isStamp(text) || (strings.HasPrefix(text, pageNotePrefix) && strings.HasSuffix(text, "]") && !strings.Contains(text, "\n"))
}

func pageField(value string) string {
	folded := strings.Map(func(r rune) rune {
		switch {
		case r == '[' || r == ']':
			return -1
		case unicode.IsControl(r):
			return ' '
		}
		return r
	}, strings.TrimSpace(value))
	if runes := []rune(folded); len(runes) > maxPageField {
		folded = string(runes[:maxPageField])
	}
	return strings.TrimSpace(folded)
}

func pageURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return false
	}
	local := parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1"
	return parsed.Scheme == "https" || (parsed.Scheme == "http" && local)
}
