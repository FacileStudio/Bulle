package assistant_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/modules/assistant"
	"github.com/go-chi/chi/v5"
)

func setupTestRouter(st store.Store, authenticated bool) chi.Router {
	router := chi.NewRouter()
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !authenticated {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ctx := authcontext.WithIdentity(r.Context(), authcontext.Identity{UserID: 7, Email: "user@facile.studio"})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	assistant.RegisterRoutes(router, st, authMiddleware)
	return router
}

func TestGetAssistantDefault(t *testing.T) {
	st := store.NewMemory()
	router := setupTestRouter(st, true)

	req := httptest.NewRequest(http.MethodGet, "/assistant", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /assistant = %d, want 200", rec.Code)
	}

	var res assistant.ProfileResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if res.UserID != 7 || res.Name != "" || res.AvatarURL != "" || res.Archetype != "" {
		t.Fatalf("unexpected defaults: %+v", res)
	}
	if res.Traits == nil || len(res.Traits) != 0 {
		t.Fatalf("expected empty non-nil traits, got %+v", res.Traits)
	}
}

func TestPutAndGetAssistant(t *testing.T) {
	st := store.NewMemory()
	router := setupTestRouter(st, true)

	body, _ := json.Marshal(assistant.UpdateProfileRequest{
		Name:      "Atlas",
		AvatarURL: "https://cdn.facile.studio/avatars/atlas.png",
		Archetype: "witty_hacker",
		Traits:    []string{"concise", "strict"},
	})
	putReq := httptest.NewRequest(http.MethodPut, "/assistant", bytes.NewReader(body))
	putRec := httptest.NewRecorder()
	router.ServeHTTP(putRec, putReq)

	if putRec.Code != http.StatusOK {
		t.Fatalf("PUT /assistant = %d, want 200", putRec.Code)
	}

	var putRes assistant.ProfileResponse
	if err := json.Unmarshal(putRec.Body.Bytes(), &putRes); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if putRes.Name != "Atlas" || putRes.Archetype != "witty_hacker" || len(putRes.Traits) != 2 {
		t.Fatalf("unexpected put response: %+v", putRes)
	}

	saved, err := st.GetAssistantProfile(context.Background(), 7)
	if err != nil {
		t.Fatalf("store get profile: %v", err)
	}
	if saved.Name != "Atlas" || saved.Archetype != "witty_hacker" || len(saved.Traits) != 2 {
		t.Fatalf("store saved mismatch: %+v", saved)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/assistant", nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GET /assistant = %d, want 200", getRec.Code)
	}
	var getRes assistant.ProfileResponse
	if err := json.Unmarshal(getRec.Body.Bytes(), &getRes); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}
	if getRes.Name != "Atlas" || getRes.Archetype != "witty_hacker" || len(getRes.Traits) != 2 {
		t.Fatalf("get mismatch after put: %+v", getRes)
	}
}

func TestPutAssistantInvalidJSON(t *testing.T) {
	st := store.NewMemory()
	router := setupTestRouter(st, true)

	req := httptest.NewRequest(http.MethodPut, "/assistant", bytes.NewReader([]byte("{invalid-json")))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("PUT /assistant with invalid json = %d, want 400", rec.Code)
	}
}

func TestAssistantUnauthenticated(t *testing.T) {
	st := store.NewMemory()
	router := setupTestRouter(st, false)

	getReq := httptest.NewRequest(http.MethodGet, "/assistant", nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated GET = %d, want 401", getRec.Code)
	}

	putReq := httptest.NewRequest(http.MethodPut, "/assistant", bytes.NewReader([]byte("{}")))
	putRec := httptest.NewRecorder()
	router.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated PUT = %d, want 401", putRec.Code)
	}
}
