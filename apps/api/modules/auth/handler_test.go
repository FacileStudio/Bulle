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
	if err := testDB.Exec("INSERT INTO users (id, email, name) VALUES (1, 'alice@facile.studio', 'Alice')").Error; err != nil {
		t.Fatal(err)
	}

	st := store.NewMemory()
	_ = st.SaveAssistantProfile(context.Background(), store.AssistantProfile{
		UserID:    1,
		Name:      "Bulle",
		AvatarURL: "https://blobatar.dev/avatar/bulle",
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
	if res.User.ID != 1 || res.User.Name != "Alice" {
		t.Fatalf("user mismatch: %+v", res.User)
	}
	if res.User.Assistant == nil || res.User.Assistant.Name != "Bulle" || res.User.Assistant.Archetype != "pragmatic_senior" {
		t.Fatalf("assistant mismatch: %+v", res.User.Assistant)
	}
}
