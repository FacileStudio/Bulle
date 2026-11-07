package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestPostgresSeeMailInOneInsert(t *testing.T) {
	if testDB == nil {
		t.Skip("no TEST_DATABASE_URL")
	}
	if err := testDB.Exec("TRUNCATE users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatal(err)
	}
	if err := testDB.Exec("INSERT INTO users (id, email) VALUES (1, 'm1@facile.studio')").Error; err != nil {
		t.Fatal(err)
	}
	st, ctx := store.NewPostgres(testDB), context.Background()
	first := time.Date(2026, 11, 7, 9, 0, 0, 0, time.UTC)
	if seen, err := st.SeeMail(ctx, 1, []int64{10, 11}, first); err != nil || len(seen) != 0 {
		t.Fatalf("first sight %v %v", seen, err)
	}
	seen, err := st.SeeMail(ctx, 1, []int64{11, 12}, first.Add(time.Hour))
	if err != nil || len(seen) != 1 || !seen[11].Equal(first) {
		t.Fatalf("second sight %v %v", seen, err)
	}
	if again, _ := st.SeeMail(ctx, 1, []int64{12}, first.Add(2*time.Hour)); len(again) != 1 {
		t.Fatalf("12 was recorded by the batch insert: %v", again)
	}
}
