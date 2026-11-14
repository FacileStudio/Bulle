package runner

import (
	"cmp"
	"encoding/json"
	"net/url"
	"regexp"
)

var markdownLink = regexp.MustCompile(`\]\(\s*<?([^)\s>]+)`)

// addResult records the URLs in a source's result. A fetched page's links
// that are relative to it are resolved against the page's URL, the request a
// browser following them sends, so lifting one is no more the model's choice
// than lifting an absolute link.
func (s seenURLs) addResult(text string) {
	s.addText(text)
	var page struct {
		FinalURL string `json:"final_url"`
		URL      string `json:"url"`
		Markdown string `json:"markdown"`
	}
	if json.Unmarshal([]byte(text), &page) != nil {
		return
	}
	base, err := url.Parse(cmp.Or(page.FinalURL, page.URL))
	if err != nil || base.Host == "" {
		return
	}
	for _, match := range markdownLink.FindAllStringSubmatch(page.Markdown, -1) {
		if ref, err := url.Parse(match[1]); err == nil {
			s.add(base.ResolveReference(ref).String())
		}
	}
}
