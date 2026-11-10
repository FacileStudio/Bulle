package matrixbot

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

func storeExists(t *testing.T, m *Manager, userID int64, device string) bool {
	t.Helper()
	var exists bool
	err := m.cfg.SQL.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", cryptoName(cryptoBase(t, m), userID, device)).Scan(&exists)
	if err != nil {
		t.Fatal(err)
	}
	return exists
}

func TestEraseDropsEveryStoreOfThePerson(t *testing.T) {
	m := cryptoManager(t)
	m.cancel, m.status = map[int64]context.CancelFunc{}, map[int64]string{}
	openCrypto(t, m, 947, "AAAA")
	openCrypto(t, m, 947, "BBBB")
	openCrypto(t, m, 948, "CCCC")
	if err := m.Erase(context.Background(), 947); err != nil {
		t.Fatal(err)
	}
	if storeExists(t, m, 947, "AAAA") || storeExists(t, m, 947, "BBBB") {
		t.Fatal("a store of the erased person survived")
	}
	if !storeExists(t, m, 948, "CCCC") {
		t.Fatal("another person's store was dropped")
	}
}

func TestSweepDropsTheStoresOfVanishedUsersOnly(t *testing.T) {
	m := cryptoManager(t)
	m.cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	if _, err := m.cfg.SQL.Exec(`INSERT INTO users (id, email) VALUES (945, 'kept-945@x.test') ON CONFLICT DO NOTHING`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = m.cfg.SQL.Exec(`DELETE FROM users WHERE id = 945`) })
	openCrypto(t, m, 945, "KEPT")
	openCrypto(t, m, 946, "GONE")
	m.sweepOrphans(context.Background())
	if !storeExists(t, m, 945, "KEPT") {
		t.Fatal("the store of a live user was dropped")
	}
	if storeExists(t, m, 946, "GONE") {
		t.Fatal("the store of a vanished user survived")
	}
}
