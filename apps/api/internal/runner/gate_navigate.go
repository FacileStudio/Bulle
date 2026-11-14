package runner

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
)

const urlTrailing = ".,;:!?)]}'\"*"

var (
	urlPattern = regexp.MustCompile("(?i)https?://[^\\s<>\"'`\\\\\\[\\]{}|^]+")
	urlEscapes = strings.NewReplacer("\\u0026", "&", "\\u003d", "=", "\\u003c", "<", "\\u003e", ">", "\\/", "/", "&amp;", "&")
)

// seenURLs is a set of normalised URLs the run has read verbatim: in the
// owner's own messages or in a result of an untrusted source. A URL lifted
// from there carries only what its author already knew; a URL the model
// built or changed is how data leaves.
type seenURLs map[string]bool

// addText records every http(s) URL in text, with and without trailing
// punctuation and with and without its fragment, after undoing the JSON and
// HTML escapes a tool result wraps URLs in.
func (s seenURLs) addText(text string) {
	for _, match := range urlPattern.FindAllString(urlEscapes.Replace(text), -1) {
		for ; match != ""; match = match[:len(match)-1] {
			s.add(match)
			if !strings.ContainsAny(match[len(match)-1:], urlTrailing) {
				break
			}
		}
	}
}

// addOwn records the URLs in the owner's own text, and every bare address
// they typed, such as go.dev/blog, as both its https and http forms.
func (s seenURLs) addOwn(text string) {
	s.addText(text)
	for _, field := range strings.Fields(text) {
		bare := strings.Trim(field, urlTrailing+"([{<>")
		if strings.Contains(bare, ".") && !strings.Contains(bare, "://") {
			s.add("https://" + bare)
			s.add("http://" + bare)
		}
	}
}

func (s seenURLs) add(raw string) {
	normal, ok := normalURL(raw)
	if !ok {
		return
	}
	s[normal] = true
	if cut, _, found := strings.Cut(normal, "#"); found {
		s[cut] = true
	}
}

// has reports whether the url argument of a fetch or navigate input is one the
// run has seen, compared after normalisation.
func (s seenURLs) has(input json.RawMessage) bool {
	var in struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return false
	}
	normal, ok := normalURL(in.URL)
	return ok && s[normal]
}

// normalURL lowercases the scheme and host, drops the default port and
// gives an empty path its slash. Everything else stays byte for byte, so two
// URLs match only when they send the same request. Only http and https
// normalise.
func normalURL(raw string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" || u.Opaque != "" {
		return "", false
	}
	u.Scheme = strings.ToLower(u.Scheme)
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", false
	}
	u.Host = strings.ToLower(u.Host)
	defaultPort := map[string]string{"http": ":80", "https": ":443"}[u.Scheme]
	u.Host = strings.TrimSuffix(u.Host, defaultPort)
	if u.Path == "" && u.RawPath == "" {
		u.Path = "/"
	}
	return u.String(), true
}

// linkLocked is need for a fetch or a navigation. The caller holds mu.
func (g *gate) linkLocked(input json.RawMessage, spend bool) ask {
	switch {
	case g.trust.own.has(input):
		return askNo
	case !g.trust.seen.has(input):
		return askCard
	case g.grants[linkScope]:
		return askNo
	case g.trust.followed >= linkBudget:
		return askLinks
	}
	if spend {
		g.trust.followed++
	}
	return askNo
}
