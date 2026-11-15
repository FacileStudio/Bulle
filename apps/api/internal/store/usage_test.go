package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/tronc/testdb"
)

func TestPostgresUsageRunsIncludeAppCalls(t *testing.T) {
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	ctx := context.Background()
	st := store.NewPostgres(testDB)
	for _, query := range []string{
		`DELETE FROM users`,
		`INSERT INTO users (id, email) VALUES (1, 'u1@facile.studio'), (2, 'u2@facile.studio')`,
		`INSERT INTO ai_usage (user_id, source, model, input_tokens, cost_usd) VALUES (1, 'complete', 'mimo', 12, 0.25), (2, 'decide', 'other', 1, 9)`,
	} {
		if err := testDB.Exec(query).Error; err != nil {
			t.Fatal(err)
		}
	}
	runs, err := st.UsageRuns(ctx, 1, time.Now().Add(-time.Hour))
	if err != nil || len(runs) != 1 {
		t.Fatalf("runs %+v, %v", runs, err)
	}
	if got := runs[0]; got.Channel != store.ChannelApps || got.Model != "mimo" || got.CostUSD != 0.25 || string(got.Usage) == "" {
		t.Fatalf("apps row %+v", got)
	}
}
