package ai

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/porte"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

// keyed serves a request that carries a BULLE_AI_CALLERS service key, and
// hands every other request to other. Deprecated: delete this function, caller,
// hasSessionCookie, actingForKey and Config.Callers once no app sends a key.
func (h *handler) keyed(next, other http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, ok := h.caller(r)
		if !ok {
			other.ServeHTTP(w, r)
			return
		}
		h.cfg.Logger.Warn("ai caller key used", "caller", name)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), callerKey{}, name)))
	})
}

func (h *handler) caller(r *http.Request) (string, bool) {
	key, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || key == "" || hasSessionCookie(r) {
		return "", false
	}
	sum := sha256.Sum256([]byte(key))
	found := ""
	for name, digest := range h.cfg.Callers {
		if subtle.ConstantTimeCompare(sum[:], digest) == 1 {
			found = name
		}
	}
	return found, found != ""
}

func hasSessionCookie(r *http.Request) bool {
	for _, cookie := range r.Cookies() {
		if strings.TrimPrefix(cookie.Name, "__Host-") == porte.SessionCookieName {
			return true
		}
	}
	return false
}

func (h *handler) actingForKey(r *http.Request) (authcontext.Identity, error) {
	email := strings.TrimSpace(r.Header.Get(actingForHeader))
	if email == "" {
		return authcontext.Identity{}, troncerrors.Forbidden(actingForHeader + " must name a Bulle user")
	}
	identity, err := h.cfg.Users.IdentityForEmail(r.Context(), email)
	if troncerrors.Status(err) == http.StatusNotFound {
		return authcontext.Identity{}, troncerrors.Forbidden(actingForHeader + " names no Bulle user")
	}
	return identity, err
}
