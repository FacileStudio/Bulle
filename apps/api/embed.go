package main

import (
	"net/http"

	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/go-chi/chi/v5"
)

// embedAncestors are the only pages allowed to frame an embed route. Every
// other route keeps frame-ancestors 'none'.
var embedAncestors = []string{"https://docs.facile.studio", "https://feuilles.facile.studio"}

// mountClient serves the web client. Under /embed/, where Docs and Feuilles
// frame Bulle's chat beside their pages, those two origins may frame it;
// everywhere else the client refuses every frame. Toile has its own composer
// and frames nothing.
func mountClient(router chi.Router, client http.Handler) {
	framing := troncmiddleware.SecurityOverride(troncmiddleware.SecurityOverrideConfig{
		CSPSources: map[string][]string{"frame-ancestors": embedAncestors},
	})
	router.With(framing).Handle("/embed/*", client)
	router.Handle("/*", client)
}
