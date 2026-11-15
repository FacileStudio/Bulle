package artifacts_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/shares"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/Bulle/apps/api/migrations"
	"github.com/FacileStudio/Bulle/apps/api/modules/artifacts"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/testdb"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

var testDB *gorm.DB

var dbConfig = testdb.Config{
	Prefix: "bulle_share_test",
	Migrate: func(db *gorm.DB) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return migrate.Run(context.Background(), migrate.Config{
			DB: sqlDB, FS: migrations.FS, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), LockID: 7_300_341,
		})
	},
}

func TestMain(m *testing.M) {
	if url, ok := testdb.URL(); ok {
		db, err := testdb.Open(url, dbConfig)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		testDB = db
	}
	os.Exit(m.Run())
}

func shareRouter(t *testing.T, gateway *suite.Gateway) chi.Router {
	t.Helper()
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testdb.Truncate(testDB, dbConfig); err != nil {
		t.Fatal(err)
	}
	if err := testDB.Exec(`INSERT INTO users (id, email) VALUES (1, 'owner@x.test'), (2, 'other@x.test')`).Error; err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	requireAuth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			id, _ := strconv.ParseInt(req.Header.Get("X-Test-User"), 10, 64)
			next.ServeHTTP(w, req.WithContext(authcontext.WithIdentity(req.Context(), authcontext.Identity{UserID: id})))
		})
	}
	artifacts.RegisterRoutes(r, store.NewMemory(), gateway, artifacts.Sharing{Store: shares.NewStore(testDB), Origin: "https://bulle.example.test"}, requireAuth)
	return r
}

func as(h http.Handler, user, method, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("X-Test-User", user)
	h.ServeHTTP(rec, req)
	return rec
}

func decodeShare(t *testing.T, rec *httptest.ResponseRecorder) artifacts.ShareResponse {
	t.Helper()
	var out artifacts.ShareResponse
	if rec.Code != http.StatusOK {
		t.Fatalf("share = %d %s", rec.Code, rec.Body)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func publicGet(h http.Handler, token string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/public/shares/"+token, nil))
	return rec
}

func ownerShare(t *testing.T) (chi.Router, string) {
	t.Helper()
	gw, fake := searchWorld(t)
	fake.MakeAdmin("1")
	id := fake.AddArtifact("", "Tea notes", "# Tea\n\nSteep three minutes.")
	return shareRouter(t, gw), "/artifacts/" + id + "/share"
}

func TestShareCreateIsIdempotentAndPrivate(t *testing.T) {
	h, path := ownerShare(t)
	first := decodeShare(t, as(h, "1", http.MethodPost, path))
	again := decodeShare(t, as(h, "1", http.MethodPost, path))
	if len(first.Token) != 43 || first.Token != again.Token || first.URL != "https://bulle.example.test/share/"+first.Token {
		t.Fatalf("create must be idempotent: %+v then %+v", first, again)
	}
	if got := decodeShare(t, as(h, "1", http.MethodGet, path)); got.Token != first.Token {
		t.Fatalf("get = %+v", got)
	}
	if rec := as(h, "2", http.MethodGet, path); rec.Code != http.StatusNotFound {
		t.Fatalf("another user must not see the link: %d", rec.Code)
	}
}

func TestPublicShareAnswersOnlyTitleContentUpdated(t *testing.T) {
	h, path := ownerShare(t)
	link := decodeShare(t, as(h, "1", http.MethodPost, path))
	public := publicGet(h, link.Token)
	var body map[string]any
	if err := json.Unmarshal(public.Body.Bytes(), &body); err != nil || public.Code != http.StatusOK {
		t.Fatalf("public = %d %s", public.Code, public.Body)
	}
	if len(body) != 3 || body["title"] != "Tea notes" || body["content"] == "" || body["updated"] == nil {
		t.Fatalf("public must answer only title, content and updated: %v", body)
	}
	for header, want := range map[string]string{"Cache-Control": "no-store", "X-Robots-Tag": "noindex, nofollow", "Referrer-Policy": "no-referrer"} {
		if got := public.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

func TestRevokedShareStaysDead(t *testing.T) {
	h, path := ownerShare(t)
	first := decodeShare(t, as(h, "1", http.MethodPost, path))
	for range 2 {
		if rec := as(h, "1", http.MethodDelete, path); rec.Code != http.StatusNoContent {
			t.Fatalf("revoke must be idempotent: %d", rec.Code)
		}
	}
	if rec := publicGet(h, first.Token); rec.Code != http.StatusNotFound {
		t.Fatalf("revoked token = %d", rec.Code)
	}
	if renewed := decodeShare(t, as(h, "1", http.MethodPost, path)); renewed.Token == first.Token {
		t.Fatal("a revoked token must never come back")
	}
	if rec := publicGet(h, first.Token); rec.Code != http.StatusNotFound {
		t.Fatalf("old token after a new link = %d", rec.Code)
	}
}

func TestErasingTheOwnerRemovesTheirShares(t *testing.T) {
	h, path := ownerShare(t)
	decodeShare(t, as(h, "1", http.MethodPost, path))
	if err := testDB.Exec(`DELETE FROM users WHERE id = 1`).Error; err != nil {
		t.Fatal(err)
	}
	var left int64
	if err := testDB.Raw(`SELECT count(*) FROM artifact_shares`).Scan(&left).Error; err != nil || left != 0 {
		t.Fatalf("erasing the user left %d shares, %v", left, err)
	}
}

func TestShareRefusesWhatTheCallerCannotRead(t *testing.T) {
	gw, fake := searchWorld(t)
	fake.MakeAdmin("1")
	id := fake.AddArtifact("", "Private", "# Mine")
	h := shareRouter(t, gw)
	if rec := as(h, "2", http.MethodPost, "/artifacts/"+id+"/share?mycelium_space=common"); rec.Code == http.StatusOK {
		t.Fatalf("a user must not share an artifact they cannot read: %d %s", rec.Code, rec.Body)
	}
	var made int64
	if err := testDB.Raw(`SELECT count(*) FROM artifact_shares`).Scan(&made).Error; err != nil || made != 0 {
		t.Fatalf("shares made = %d, %v", made, err)
	}
}

func TestPublicShareHidesUnknownTokensAndDeletedArtifacts(t *testing.T) {
	gw, fake := searchWorld(t)
	fake.MakeAdmin("1")
	id := fake.AddArtifact("", "Soon gone", "# Bye")
	h := shareRouter(t, gw)
	link := decodeShare(t, as(h, "1", http.MethodPost, "/artifacts/"+id+"/share"))

	unknown := publicGet(h, "nope")
	if unknown.Code != http.StatusNotFound || unknown.Header().Get("Cache-Control") != "no-store" || unknown.Header().Get("X-Robots-Tag") != "noindex, nofollow" {
		t.Fatalf("unknown token = %d %v", unknown.Code, unknown.Header())
	}

	if rec := as(h, "1", http.MethodDelete, "/artifacts/"+id+"?mycelium_space=common"); rec.Code != http.StatusNoContent {
		t.Fatalf("delete = %d %s", rec.Code, rec.Body)
	}
	deleted := publicGet(h, link.Token)
	if deleted.Code != http.StatusNotFound || deleted.Body.String() != unknown.Body.String() {
		t.Fatalf("deleted artifact = %d %s, want the unknown token's answer %s", deleted.Code, deleted.Body, unknown.Body)
	}
}
