package spaces_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/modules/spaces"
	"github.com/go-chi/chi/v5"
)

func newRouter(st store.Store) chi.Router {
	router := chi.NewRouter()
	asUser := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := strconv.ParseInt(r.Header.Get("X-User"), 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(authcontext.WithIdentity(r.Context(), authcontext.Identity{UserID: id})))
		})
	}
	spaces.RegisterRoutes(router, st, asUser)
	return router
}

type client struct {
	t      *testing.T
	router http.Handler
}

func (c client) call(user, method, path, body string) *httptest.ResponseRecorder {
	c.t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("X-User", user)
	rec := httptest.NewRecorder()
	c.router.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decoding %q: %v", rec.Body.String(), err)
	}
	return out
}

func TestListCreatesTheDefaultSpace(t *testing.T) {
	api := client{t, newRouter(store.NewMemory())}
	rec := api.call("7", http.MethodGet, "/spaces", "")
	listed := decode[[]spaces.SpaceResponse](t, rec)
	if rec.Code != http.StatusOK || len(listed) != 1 || listed[0].Name != store.DefaultSpaceName {
		t.Fatalf("GET /spaces = %d %+v", rec.Code, listed)
	}
	if rec := api.call("", http.MethodGet, "/spaces", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated GET = %d", rec.Code)
	}
}

func TestCreateAndUpdateSpace(t *testing.T) {
	api := client{t, newRouter(store.NewMemory())}
	if rec := api.call("7", http.MethodPost, "/spaces", `{"name":"  "}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("a blank name must be refused, got %d", rec.Code)
	}
	rec := api.call("7", http.MethodPost, "/spaces", `{"name":" Work ","courrier_account":4}`)
	created := decode[spaces.SpaceResponse](t, rec)
	if rec.Code != http.StatusCreated || created.Name != "Work" || created.CourrierAccount != 4 {
		t.Fatalf("POST /spaces = %d %+v", rec.Code, created)
	}
	rec = api.call("7", http.MethodPut, "/spaces/"+created.ID, `{"opus_workspace":"Acme"}`)
	updated := decode[spaces.SpaceResponse](t, rec)
	if rec.Code != http.StatusOK || updated.Name != "Work" || updated.OpusWorkspace != "Acme" || updated.CourrierAccount != 4 {
		t.Fatalf("PUT must change only the fields sent: %d %+v", rec.Code, updated)
	}
}

func TestSpacesOfAnotherUserAreHidden(t *testing.T) {
	st := store.NewMemory()
	api := client{t, newRouter(st)}
	other, _ := st.CreateSpace(context.Background(), store.Space{UserID: 8, Name: "Theirs"})
	if _, err := st.EnsureDefaultSpace(context.Background(), 8); err != nil {
		t.Fatal(err)
	}
	if rec := api.call("7", http.MethodPut, "/spaces/"+other.ID, `{"name":"Mine"}`); rec.Code != http.StatusNotFound {
		t.Fatalf("PUT on another user's space = %d, want 404", rec.Code)
	}
	if rec := api.call("7", http.MethodDelete, "/spaces/"+other.ID, ""); rec.Code != http.StatusNotFound {
		t.Fatalf("DELETE on another user's space = %d, want 404", rec.Code)
	}
	listed := decode[[]spaces.SpaceResponse](t, api.call("7", http.MethodGet, "/spaces", ""))
	if len(listed) != 1 || listed[0].ID == other.ID {
		t.Fatalf("user 7 sees another user's space: %+v", listed)
	}
	if got, _ := st.GetSpace(context.Background(), 8, other.ID); got.Name != "Theirs" {
		t.Fatalf("another user's space changed: %+v", got)
	}
}

func TestDeleteRefusesTheLastSpace(t *testing.T) {
	st := store.NewMemory()
	api := client{t, newRouter(st)}
	home, _ := st.EnsureDefaultSpace(context.Background(), 7)
	work, _ := st.CreateSpace(context.Background(), store.Space{UserID: 7, Name: "Work"})
	if _, err := st.CreateSession(context.Background(), store.Session{ID: store.NewID(), UserID: 7, SpaceID: work.ID}); err != nil {
		t.Fatal(err)
	}
	if rec := api.call("7", http.MethodDelete, "/spaces/"+work.ID, ""); rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE = %d, want 204", rec.Code)
	}
	if sessions, _ := st.ListSessions(context.Background(), 7, ""); len(sessions) != 0 {
		t.Fatalf("the deleted space's sessions survived: %+v", sessions)
	}
	if rec := api.call("7", http.MethodDelete, "/spaces/"+home.ID, ""); rec.Code != http.StatusConflict {
		t.Fatalf("deleting the last space = %d, want 409", rec.Code)
	}
}
