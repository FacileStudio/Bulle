package ai

import (
	"context"
	"net/http"
	"strconv"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/porte"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

type callerKey struct{}

// authenticate lets in only a token an app exchanged for its user, and names
// that app as the caller.
func (h *handler) authenticate(next http.Handler) http.Handler {
	if h.cfg.RequireAuth == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			httpjson.WriteError(w, troncerrors.Unauthorized("a token exchanged by an app for its user is required"))
		})
	}
	return h.cfg.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, signedIn := authcontext.IdentityFromContext(r.Context())
		actor, acting := porte.ActorFrom(r.Context())
		if !signedIn || !acting {
			httpjson.WriteError(w, troncerrors.Forbidden("a token exchanged by an app for its user is required"))
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), callerKey{}, actor)))
	}))
}

func callerOf(r *http.Request) string {
	name, _ := r.Context().Value(callerKey{}).(string)
	return name
}

func callerAndUser(r *http.Request) string {
	return callerOf(r) + "\x00" + strconv.FormatInt(actingOf(r).UserID, 10)
}

// admit resolves the user a service call acts for and refuses the call when
// that user's plan lacks ai.access, before any rate limit is spent, so a
// refused user never takes a paying user's share of the caller's budget.
func (h *handler) admit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, err := h.actingFor(r)
		if err == nil && h.cfg.Access != nil {
			err = h.cfg.Access.CheckAIIn(r.Context(), identity.UserID, "")
		}
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), actingKey{}, identity)))
	})
}

func actorAndUser(r *http.Request) string {
	identity, _ := porte.From(r.Context())
	return identity.Actor + "\x00" + identity.Subject
}
