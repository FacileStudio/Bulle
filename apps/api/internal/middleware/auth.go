// Package middleware holds the HTTP middleware that authenticates a request and
// puts the caller's identity on its context.
package middleware

import (
	"context"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

type Authenticator interface {
	RequireAuth(http.Handler) http.Handler
}

type IdentityResolver interface {
	IdentityForUser(ctx context.Context, userID int64) (authcontext.Identity, error)
}

func RequireAuth(sessions Authenticator, resolver IdentityResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return sessions.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			authenticated, ok := porte.From(request.Context())
			if !ok {
				httpjson.WriteError(w, errors.Unauthorized("missing auth token"))
				return
			}
			identity, err := resolver.IdentityForUser(request.Context(), authenticated.UserID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			next.ServeHTTP(w, request.WithContext(authcontext.WithIdentity(request.Context(), identity)))
		}))
	}
}
