package llm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

type Resolver interface {
	Resolve(ctx context.Context, userID int64, backend string) (string, error)
}

type backend struct {
	base   *url.URL
	header string
	scheme string
}

func defaultBackends() map[string]backend {
	mustParse := func(raw string) *url.URL {
		parsed, err := url.Parse(raw)
		if err != nil {
			panic(err)
		}
		return parsed
	}
	return map[string]backend{
		"anthropic":  {base: mustParse("https://api.anthropic.com"), header: "X-Api-Key"},
		"openai":     {base: mustParse("https://api.openai.com/v1"), header: "Authorization", scheme: "Bearer "},
		"google":     {base: mustParse("https://generativelanguage.googleapis.com"), header: "X-Goog-Api-Key"},
		"openrouter": {base: mustParse("https://openrouter.ai/api/v1"), header: "Authorization", scheme: "Bearer "},
	}
}

var credentialHeaders = []string{
	"Authorization",
	"X-Api-Key",
	"X-Goog-Api-Key",
	"Cookie",
	"Proxy-Authorization",
}

type Handler struct {
	resolver Resolver
	errNoKey error
	backends map[string]backend
	reverse  *httputil.ReverseProxy
}

func NewHandler(resolver Resolver, errNoKey error) *Handler {
	h := &Handler{resolver: resolver, errNoKey: errNoKey, backends: defaultBackends()}
	h.reverse = &httputil.ReverseProxy{
		Rewrite:       rewrite,
		FlushInterval: -1,
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, _ error) {
			writeJSON(w, http.StatusBadGateway, "upstream request failed")
		},
	}
	return h
}

type target struct {
	backend backend
	key     string
	rest    string
}

type targetKey struct{}

func rewrite(pr *httputil.ProxyRequest) {
	t := pr.In.Context().Value(targetKey{}).(target)
	out := pr.Out
	for _, name := range credentialHeaders {
		out.Header.Del(name)
	}
	out.Header.Set(t.backend.header, t.backend.scheme+t.key)
	out.URL.Scheme = t.backend.base.Scheme
	out.URL.Host = t.backend.base.Host
	out.URL.Path = strings.TrimRight(t.backend.base.Path, "/") + "/" + t.rest
	out.URL.RawPath = ""
	query := pr.In.URL.Query()
	query.Del("key")
	out.URL.RawQuery = query.Encode()
	out.Host = t.backend.base.Host
}

func (h *Handler) proxy(w http.ResponseWriter, r *http.Request) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return
	}
	name := chiParam(r, "backend")
	b, found := h.backends[name]
	if !found {
		writeJSON(w, http.StatusNotFound, "unknown backend "+name)
		return
	}
	key, err := h.resolver.Resolve(r.Context(), identity.UserID, name)
	if errors.Is(err, h.errNoKey) || (err == nil && key == "") {
		writeJSON(w, http.StatusConflict, "no key for "+name)
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "could not resolve key for "+name)
		return
	}
	ctx := context.WithValue(r.Context(), targetKey{}, target{backend: b, key: key, rest: chiParam(r, "*")})
	h.reverse.ServeHTTP(w, r.WithContext(ctx))
}

func writeJSON(w http.ResponseWriter, status int, message string) {
	httpjson.WriteJSON(w, status, ErrorResponse{Error: message})
}

func chiParam(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}
