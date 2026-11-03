package llm

import (
	"net/http"

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
	if identity, ok := authcontext.IdentityFromContext(r.Context()); ok && o.Persona != nil {
		out.Persona = o.Persona(r.Context(), identity.UserID)
	}
	httpjson.WriteJSON(w, http.StatusOK, out)
}
