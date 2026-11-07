package runner

import (
	"encoding/json"
	"net/url"
	"strings"
	"unicode"
)

const (
	tokenMixed = 16
	tokenPlain = 32
)

// unseenData reports whether a browser_navigate would carry data out that the
// owner never wrote: a query value, a fragment, or an opaque path token
// (a long run of letters and digits) not found verbatim in the owner's own
// messages. A plain page address never asks, so reading pages stays unasked.
func unseenData(input json.RawMessage, own string) bool {
	var in struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return true
	}
	raw := strings.TrimSpace(in.URL)
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" {
		return true
	}
	if strings.Contains(own, raw) {
		return false
	}
	for _, token := range carried(parsed) {
		if !strings.Contains(own, token) {
			return true
		}
	}
	return false
}

// carried is every piece of a URL that can smuggle data: query values,
// opaque query keys, the fragment, and the opaque tokens in its path.
func carried(u *url.URL) []string {
	var out []string
	for key, values := range u.Query() {
		if opaque(key) {
			out = append(out, key)
		}
		for _, value := range values {
			if value != "" {
				out = append(out, value)
			}
		}
	}
	if u.Fragment != "" {
		out = append(out, u.Fragment)
	}
	return append(out, pathTokens(u.Path)...)
}

func pathTokens(p string) []string {
	var out []string
	runs := strings.FieldsFunc(p, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
	for _, run := range runs {
		if opaque(run) {
			out = append(out, run)
		}
	}
	return out
}

func opaque(run string) bool {
	if len(run) >= tokenPlain {
		return true
	}
	if len(run) < tokenMixed {
		return false
	}
	return strings.ContainsFunc(run, unicode.IsDigit) && strings.ContainsFunc(run, unicode.IsLetter)
}
