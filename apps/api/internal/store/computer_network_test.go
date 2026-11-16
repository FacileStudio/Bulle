package store_test

import (
	"context"
	"slices"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

type preferenceStore interface {
	store.ComputerNetworks
	GetPreferences(ctx context.Context, userID int64) (store.Preferences, error)
	SavePreferences(ctx context.Context, prefs store.Preferences) error
}

func TestMemoryComputerNetworkKeepsTheModel(t *testing.T) {
	checkComputerNetwork(t, store.NewMemory())
	checkComputerNetAllow(t, store.NewMemory())
}

func TestPostgresComputerNetworkKeepsTheModel(t *testing.T) {
	if testDB == nil {
		t.Skip("no TEST_DATABASE_URL")
	}
	if err := testDB.Exec("TRUNCATE users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatal(err)
	}
	if err := testDB.Exec("INSERT INTO users (id, email) VALUES (1, 'n1@facile.studio')").Error; err != nil {
		t.Fatal(err)
	}
	checkComputerNetwork(t, store.NewPostgres(testDB))
	checkComputerNetAllow(t, store.NewPostgres(testDB))
}

func checkComputerNetwork(t *testing.T, st preferenceStore) {
	t.Helper()
	ctx := context.Background()
	if got, allow, err := st.ComputerNetwork(ctx, 1); err != nil || got != "" || len(allow) != 0 {
		t.Fatalf("never chosen = %q %v %v, want empty", got, allow, err)
	}
	if err := st.SavePreferences(ctx, store.Preferences{UserID: 1, Model: "m"}); err != nil {
		t.Fatal(err)
	}
	if err := st.SaveComputerNetwork(ctx, 1, "offline", nil); err != nil {
		t.Fatal(err)
	}
	if err := st.SavePreferences(ctx, store.Preferences{UserID: 1, Model: "n"}); err != nil {
		t.Fatal(err)
	}
	prefs, err := st.GetPreferences(ctx, 1)
	if err != nil || prefs.Model != "n" || prefs.ComputerNetwork != "offline" {
		t.Fatalf("prefs = %+v %v", prefs, err)
	}
	if got, _, _ := st.ComputerNetwork(ctx, 1); got != "offline" {
		t.Fatalf("network = %q, want offline", got)
	}
}

func checkComputerNetAllow(t *testing.T, st preferenceStore) {
	t.Helper()
	ctx := context.Background()
	hosts := []string{"api.example.com", "*.cdn.example.org:8443"}
	if err := st.SaveComputerNetwork(ctx, 1, "allowlist", hosts); err != nil {
		t.Fatal(err)
	}
	if err := st.SavePreferences(ctx, store.Preferences{UserID: 1, Model: "o"}); err != nil {
		t.Fatal(err)
	}
	if got, allow, err := st.ComputerNetwork(ctx, 1); err != nil || got != "allowlist" || !slices.Equal(allow, hosts) {
		t.Fatalf("allowlist = %q %v %v", got, allow, err)
	}
	if err := st.SaveComputerNetwork(ctx, 1, "open", nil); err != nil {
		t.Fatal(err)
	}
	if got, allow, err := st.ComputerNetwork(ctx, 1); err != nil || got != "open" || len(allow) != 0 {
		t.Fatalf("open = %q %v %v", got, allow, err)
	}
}
