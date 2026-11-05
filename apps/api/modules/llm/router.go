package llm

import (
	"net/http"
	"slices"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, resolver Resolver, errNoKey error, offer Offer, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(resolver, errNoKey)
	router.With(requireAuth).Get("/llm", offer.serve)
	router.With(requireAuth).HandleFunc("/llm/{backend}", handler.proxy)
	router.With(requireAuth).HandleFunc("/llm/{backend}/*", handler.proxy)
}

func (o Offer) serve(w http.ResponseWriter, r *http.Request) {
	out := o.Config
	if identity, ok := authcontext.IdentityFromContext(r.Context()); ok {
		if o.Persona != nil {
			out.Persona = o.Persona(r.Context(), identity.UserID)
		}
		// The user's own default wins over the instance's, but only while the
		// instance still offers it: a model that left BULLE_LLM_MODELS must not
		// be handed back to a client that would then be refused.
		if o.ModelFor != nil {
			if model := o.ModelFor(r.Context(), identity.UserID); model != "" && slices.Contains(out.Models, model) {
				out.Model = model
			}
		}
	}
	httpjson.WriteJSON(w, http.StatusOK, out)
}
