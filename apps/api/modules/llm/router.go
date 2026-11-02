package llm

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, resolver Resolver, errNoKey error, requireAuth func(http.Handler) http.Handler) {
	handler := NewHandler(resolver, errNoKey)
	router.With(requireAuth).HandleFunc("/llm/{backend}", handler.proxy)
	router.With(requireAuth).HandleFunc("/llm/{backend}/*", handler.proxy)
}
