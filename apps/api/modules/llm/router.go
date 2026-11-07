package llm

import (
	"net/http"
	"slices"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, resolver Resolver, errNoKey error, offer Offer, catalog CatalogFunc, requireAuth, requireAI func(http.Handler) http.Handler) {
	handler := NewHandler(resolver, errNoKey)
	router.With(requireAuth).Get("/llm", offer.serve)
	router.With(requireAuth).Get("/models", catalogHandler(catalog))
	router.With(requireAuth, requireAI).HandleFunc("/llm/{backend}", handler.proxy)
	router.With(requireAuth, requireAI).HandleFunc("/llm/{backend}/*", handler.proxy)
}

func (o Offer) serve(w http.ResponseWriter, r *http.Request) {
	out := o.Config
	if identity, ok := authcontext.IdentityFromContext(r.Context()); ok {
		if o.Persona != nil {
			out.Persona = o.Persona(r.Context(), identity.UserID)
		}
		// The catalogue is the same list the web picker searches, so a session
		// in either surface switches between the same models.
		if o.Catalog != nil {
			ids := make([]string, 0)
			for _, model := range o.Catalog(r.Context(), identity.UserID) {
				ids = append(ids, model.ID)
			}
			out.Models = merged(out.Models, ids)
		}
		// The user's own default wins over the instance's, but only while the
		// backend still offers it: a model that left the catalogue must not be
		// handed back to a client that would then be refused.
		if o.ModelFor != nil {
			if model := o.ModelFor(r.Context(), identity.UserID); model != "" && slices.Contains(out.Models, model) {
				out.Model = model
			}
		}
	}
	httpjson.WriteJSON(w, http.StatusOK, out)
}

// merged keeps the operator's list first and appends what the vendor adds, so
// a deployment that pins a short list still leads with it.
func merged(pinned, found []string) []string {
	out := make([]string, 0, len(pinned)+len(found))
	seen := make(map[string]bool, len(pinned)+len(found))
	for _, group := range [][]string{pinned, found} {
		for _, id := range group {
			if id == "" || seen[id] {
				continue
			}
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}
