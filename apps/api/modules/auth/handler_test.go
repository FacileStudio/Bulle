package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/migrations"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/testdb"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

var testDB *gorm.DB

var dbConfig = testdb.Config{
	Prefix: "bulle_auth_test",
	Migrate: func(db *gorm.DB) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return migrate.Run(context.Background(), migrate.Config{
			DB:     sqlDB,
			FS:     migrations.FS,
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			LockID: 7_300_302,
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

func TestMeEndpointWithAssistantProfile(t *testing.T) {
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testdb.Truncate(testDB, dbConfig); err != nil {
		t.Fatal(err)
	}
	if err := testDB.Exec("INSERT INTO users (id, email, name, picture_url) VALUES (1, 'alice@facile.studio', 'Alice', 'https://sso.facile.studio/avatars/alice')").Error; err != nil {
		t.Fatal(err)
	}

	st := store.NewMemory()
	_ = st.SaveAssistantProfile(context.Background(), store.AssistantProfile{
		UserID:    1,
		Name:      "Bulle",
		AvatarURL: "https://cdn.facile.studio/avatars/bulle.png",
		Archetype: "pragmatic_senior",
		Traits:    []string{"concise"},
	})

	service := NewService(testDB)
	router := chi.NewRouter()
	requireAuth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := authcontext.WithIdentity(r.Context(), authcontext.Identity{UserID: 1, Email: "alice@facile.studio"})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	RegisterRoutes(router, service, st, requireAuth)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /auth/me = %d, want 200", rec.Code)
	}

	var res MeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if res.User.ID != 1 || res.User.Name != "Alice" || res.User.AvatarURL != "https://sso.facile.studio/avatars/alice" {
		t.Fatalf("user mismatch: %+v", res.User)
	}
	if res.User.Assistant == nil || res.User.Assistant.Name != "Bulle" || res.User.Assistant.Archetype != "pragmatic_senior" {
		t.Fatalf("assistant mismatch: %+v", res.User.Assistant)
	}
}

func TestPictureStoredOnSignInAndBackfilled(t *testing.T) {
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testdb.Truncate(testDB, dbConfig); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	const issuer = "https://sso.facile.studio"
	const avatar = issuer + "/avatars/sub-alice"

	users := NewUserStore(testDB, nil)
	alice, err := users.UpsertFromOIDC(ctx, porte.Claims{Provider: issuer, Subject: "sub-alice", Email: "alice@facile.studio", EmailVerified: true, Picture: avatar})
	if err != nil {
		t.Fatal(err)
	}
	if err := testDB.Exec("INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, 'sub-alice')", alice, issuer).Error; err != nil {
		t.Fatal(err)
	}
	if err := testDB.Exec("INSERT INTO users (id, email) VALUES (90, 'bob@facile.studio'), (91, 'carol@facile.studio')").Error; err != nil {
		t.Fatal(err)
	}
	if err := testDB.Exec("INSERT INTO porte_identities (user_id, provider, subject) VALUES (90, ?, 'sub-bob'), (91, 'https://elsewhere.example', 'sub-carol')", issuer).Error; err != nil {
		t.Fatal(err)
	}

	for run, want := range []int64{1, 0} {
		filled, err := BackfillPictures(ctx, testDB, issuer)
		if err != nil {
			t.Fatal(err)
		}
		if filled != want {
			t.Fatalf("backfill run %d filled %d users, want %d", run, filled, want)
		}
	}

	pictures := map[int64]string{}
	var rows []struct {
		ID         int64
		PictureURL string
	}
	if err := testDB.Raw("SELECT id, picture_url FROM users").Scan(&rows).Error; err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		pictures[row.ID] = row.PictureURL
	}
	want := map[int64]string{alice: avatar, 90: issuer + "/avatars/sub-bob", 91: ""}
	for id, url := range want {
		if pictures[id] != url {
			t.Errorf("user %d picture_url = %q, want %q", id, pictures[id], url)
		}
	}

	if _, err := users.UpsertFromOIDC(ctx, porte.Claims{Provider: issuer, Subject: "sub-alice", Email: "alice@facile.studio", EmailVerified: true, Picture: "data:image/svg+xml;base64,AAAA"}); err != nil {
		t.Fatal(err)
	}
	var stored string
	if err := testDB.Raw("SELECT picture_url FROM users WHERE id = ?", alice).Scan(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored != "" {
		t.Errorf("a non-https picture claim was stored as %q", stored)
	}
}
