package auth

import (
	"context"
	"testing"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/testdb"
)

func TestSecondSubjectOfTheSameIssuerIsRefused(t *testing.T) {
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testdb.Truncate(testDB, dbConfig); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	const issuer = "https://sso.facile.studio"
	users := NewUserStore(testDB, nil)

	owner, err := users.UpsertFromOIDC(ctx, porte.Claims{Provider: issuer, Subject: "sub-owner", Email: "Alice@Facile.studio", EmailVerified: true, Locale: "fr"})
	if err != nil {
		t.Fatal(err)
	}
	if err := testDB.Exec("INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, 'sub-owner')", owner, issuer).Error; err != nil {
		t.Fatal(err)
	}

	if _, err := users.UpsertFromOIDC(ctx, porte.Claims{Provider: issuer, Subject: "sub-squatter", Email: "alice@facile.studio", EmailVerified: true}); err == nil {
		t.Fatal("a second subject of the same issuer took over the account")
	}
	if _, err := users.UpsertFromOIDC(ctx, porte.Claims{Provider: issuer, Subject: "sub-squatter", Email: "alice@facile.studio", EmailVerified: false}); err == nil {
		t.Fatal("an unverified email was accepted")
	}

	again, err := users.UpsertFromOIDC(ctx, porte.Claims{Provider: issuer, Subject: "sub-owner", Email: "alice@facile.studio", EmailVerified: true})
	if err != nil || again != owner {
		t.Fatalf("the owner's next sign-in = %d, %v; want %d", again, err, owner)
	}

	linked, err := users.UpsertFromOIDC(ctx, porte.Claims{Provider: "https://other.example", Subject: "sub-other", Email: "alice@facile.studio", EmailVerified: true, Locale: "de"})
	if err != nil || linked != owner {
		t.Fatalf("a verified identity from another issuer = %d, %v; want %d", linked, err, owner)
	}

	var locale string
	if err := testDB.Raw("SELECT locale FROM users WHERE id = ?", owner).Scan(&locale).Error; err != nil {
		t.Fatal(err)
	}
	if locale != "de" {
		t.Fatalf("locale = %q, want de", locale)
	}
}

func TestNewAccountStartsInEnglishWithoutALocaleClaim(t *testing.T) {
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testdb.Truncate(testDB, dbConfig); err != nil {
		t.Fatal(err)
	}
	id, err := NewUserStore(testDB, nil).UpsertFromOIDC(context.Background(), porte.Claims{Provider: "https://sso.facile.studio", Subject: "sub-new", Email: "new@facile.studio", EmailVerified: true})
	if err != nil {
		t.Fatal(err)
	}
	var locale string
	if err := testDB.Raw("SELECT locale FROM users WHERE id = ?", id).Scan(&locale).Error; err != nil || locale != "en" {
		t.Fatalf("locale = %q, %v; want en", locale, err)
	}
}
