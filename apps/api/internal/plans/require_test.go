package plans_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/plans"
)

func TestRequire(t *testing.T) {
	var gate *plans.Gate
	reached := false
	handler := gate.Require(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusNoContent)
	}))

	anonymous := httptest.NewRecorder()
	handler.ServeHTTP(anonymous, httptest.NewRequest(http.MethodPost, "/", nil))
	if anonymous.Code != http.StatusUnauthorized || reached {
		t.Fatalf("anonymous: want 401 without reaching the handler, got %d reached=%v", anonymous.Code, reached)
	}

	signedIn := httptest.NewRequest(http.MethodPost, "/", nil)
	signedIn = signedIn.WithContext(authcontext.WithIdentity(signedIn.Context(), authcontext.Identity{UserID: 1}))
	allowed := httptest.NewRecorder()
	handler.ServeHTTP(allowed, signedIn)
	if allowed.Code != http.StatusNoContent || !reached {
		t.Fatalf("signed in, gate inactive: want the handler, got %d", allowed.Code)
	}
}
