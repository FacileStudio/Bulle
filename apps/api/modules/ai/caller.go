package ai

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strconv"
	"strings"

	"github.com/FacileStudio/porte"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

type callerKey struct{}

func (h *handler) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		caller, ok := h.caller(r)
		if !ok {
			httpjson.WriteError(w, troncerrors.Unauthorized("a service key is required"))
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), callerKey{}, caller)))
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

func callerOf(r *http.Request) string {
	name, _ := r.Context().Value(callerKey{}).(string)
	return name
}

func callerAndUser(r *http.Request) string {
	return callerOf(r) + "\x00" + strconv.FormatInt(actingOf(r).UserID, 10)
}

func actorAndUser(r *http.Request) string {
	identity, _ := porte.From(r.Context())
	return identity.Actor + "\x00" + identity.Subject
}
