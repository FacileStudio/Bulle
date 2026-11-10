package credentials_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	internal "github.com/FacileStudio/Bulle/apps/api/internal/credentials"
	"github.com/FacileStudio/Bulle/apps/api/modules/credentials"
	"github.com/go-chi/chi/v5"
)

func router(t *testing.T, userID int64, urls map[string]string) chi.Router {
	t.Helper()
	svc, err := internal.NewMemory(bytes.Repeat([]byte{3}, 32))
	if err != nil {
		t.Fatal(err)
	}
	h := credentials.NewHandler(svc, urls)
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := authcontext.WithIdentity(req.Context(), authcontext.Identity{UserID: userID})
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	r.Get("/credentials", h.List)
	r.Put("/credentials/{backend}", h.Put)
	r.Delete("/credentials/{backend}", h.Remove)
	r.Post("/credentials/{backend}/test", h.Test)
	return r
}

func do(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCredentialsLifecycleNeverReturnsKey(t *testing.T) {
	for _, name := range []string{"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY", "GOOGLE_API_KEY"} {
		t.Setenv(name, "")
	}
	t.Setenv("OPENROUTER_API_KEY", "sk-instance-key")
	vendor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "sk-ant-abcdef1234" {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer vendor.Close()
	rt := router(t, 5, map[string]string{"anthropic": vendor.URL})

	if rec := do(rt, http.MethodPut, "/credentials/nope", `{"api_key":"x"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown backend: %d", rec.Code)
	}
	rec := do(rt, http.MethodPut, "/credentials/anthropic", `{"api_key":"sk-ant-abcdef1234"}`)
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "abcdef") {
		t.Fatalf("put: %d %s", rec.Code, rec.Body.String())
	}

	rec = do(rt, http.MethodGet, "/credentials", "")
	var list []credentials.CredentialResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || len(list) != 4 {
		t.Fatalf("list: %s", rec.Body.String())
	}
	if list[0].Backend != "anthropic" || list[0].Source != "user" || list[0].Last4 != "1234" {
		t.Fatalf("anthropic row: %+v", list[0])
	}
	if list[3].Source != "instance" || list[1].Source != "none" {
		t.Fatalf("sources: %+v", list)
	}

	var result credentials.TestCredentialResponse
	rec = do(rt, http.MethodPost, "/credentials/anthropic/test", "")
	_ = json.Unmarshal(rec.Body.Bytes(), &result)
	if !result.OK {
		t.Fatalf("test: %s", rec.Body.String())
	}

	if rec = do(rt, http.MethodDelete, "/credentials/anthropic", ""); rec.Code != http.StatusOK {
		t.Fatalf("delete: %d", rec.Code)
	}
	rec = do(rt, http.MethodGet, "/credentials", "")
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if list[0].Source != "none" {
		t.Fatalf("after delete: %+v", list[0])
	}
}

func TestSearchKeysLiveInSonarNotBulle(t *testing.T) {
	rt := router(t, 5, nil)
	if rec := do(rt, http.MethodPut, "/credentials/tavily", `{"api_key":"tvly-secret-9876"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("a search engine key must be refused: %d %s", rec.Code, rec.Body.String())
	}
	var list []credentials.CredentialResponse
	if err := json.Unmarshal(do(rt, http.MethodGet, "/credentials", "").Body.Bytes(), &list); err != nil || len(list) != 4 {
		t.Fatalf("only the model backends are listed: %v %+v", err, list)
	}
}
