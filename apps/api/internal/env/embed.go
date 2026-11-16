package env

import (
	"fmt"
	"net/url"
)

// embedDefaults are the hosted suite apps that frame the chat beside a page.
var embedDefaults = []string{
	"https://docs.facile.studio", "https://feuilles.facile.studio",
	"https://diapo.facile.studio", "https://clap.facile.studio",
}

// loadEmbedOrigins reads BULLE_EMBED_ORIGINS, the pages allowed to frame the
// embedded chat. Each must be a bare https origin, or http on localhost.
func loadEmbedOrigins() ([]string, error) {
	origins := csvOrDefault("BULLE_EMBED_ORIGINS", embedDefaults)
	for _, origin := range origins {
		if !embedOrigin(origin) {
			return nil, fmt.Errorf("BULLE_EMBED_ORIGINS contains invalid origin %q: use https://host or http://localhost", origin)
		}
	}
	return origins, nil
}

func embedOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil || parsed.Opaque != "" {
		return false
	}
	return parsed.Scheme == "https" || (parsed.Scheme == "http" && parsed.Hostname() == "localhost")
}
