package runner

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/FacileStudio/nacelle"
)

var urlPattern = regexp.MustCompile(`https?://[^\s<>"'` + "`" + `\])}]+`)

// provenance is the set of URLs a run has already seen: in what the user wrote
// and in what a tool returned. A web_fetch of any other URL was invented by the
// model, maybe on a page's say-so with conversation data in its query string,
// so it goes to the owner first.
type provenance struct {
	mu   sync.Mutex
	seen map[string]bool
}

func newProvenance(history []nacelle.Message) *provenance {
	p := &provenance{seen: map[string]bool{}}
	for _, message := range history {
		if message.Role != nacelle.RoleUser {
			continue
		}
		for _, part := range message.Parts {
			switch part := part.(type) {
			case nacelle.Text:
				p.add(part.Text)
			case nacelle.ToolResult:
				p.add(part.Result)
			}
		}
	}
	return p
}

func (p *provenance) add(text string) {
	found := urlPattern.FindAllString(text, -1)
	if len(found) == 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, raw := range found {
		if key, ok := urlKey(strings.TrimRight(raw, ".,;:!?*_")); ok {
			p.seen[key] = true
		}
	}
}

func (p *provenance) known(raw string) bool {
	key, ok := urlKey(raw)
	if !ok {
		return false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.seen[key]
}

// unknownFetch reports whether a web_fetch input names a URL this run has not
// seen, which is when the owner must approve it.
func (p *provenance) unknownFetch(input json.RawMessage) bool {
	var in struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return true
	}
	return !p.known(in.URL)
}

func urlKey(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", false
	}
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment, parsed.RawFragment = "", ""
	if parsed.Path == "/" {
		parsed.Path, parsed.RawPath = "", ""
	}
	return parsed.String(), true
}

// hostTools builds the run's host tools with the web limits applied, and puts
// every web_fetch or browser_navigate of a URL the run has not seen through
// the owner's approval.
func (r *Runner) hostTools(lr *liveRun) ([]nacelle.Tool, error) {
	host, err := r.cfg.HostTools(lr.userID)
	if err != nil {
		return nil, err
	}
	lr.gate.asks = map[string]func(json.RawMessage) bool{
		"web_fetch":        lr.seen.unknownFetch,
		"browser_navigate": lr.seen.unknownFetch,
	}
	return capWeb(host), nil
}
