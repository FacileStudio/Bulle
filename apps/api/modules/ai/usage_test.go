package ai_test

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/aiusage"
	"github.com/FacileStudio/Bulle/apps/api/modules/ai"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

type recorded struct {
	userID         int64
	source, caller string
	call           aiusage.Call
}

type fakeLedger struct {
	mu   sync.Mutex
	rows []recorded
}

func (f *fakeLedger) Record(_ context.Context, userID int64, source, caller string, call aiusage.Call) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rows = append(f.rows, recorded{userID, source, caller, call})
}

func withLedger(ledger ai.Ledger) func(*ai.Config) {
	return func(cfg *ai.Config) { cfg.Ledger = ledger }
}

func TestCompleteRecordsWhatItCost(t *testing.T) {
	ledger := &fakeLedger{}
	srv := completeServer(t, &fakeCompleter{answer: "ok"}, nil, withLedger(ledger))
	if res, raw := complete(t, srv, toileWrite, `{"prompt":"hi"}`); res.StatusCode != http.StatusOK {
		t.Fatalf("status %d: %s", res.StatusCode, raw)
	}
	if len(ledger.rows) != 1 {
		t.Fatalf("rows %+v", ledger.rows)
	}
	row := ledger.rows[0]
	if row.userID != 7 || row.source != aiusage.SourceComplete || row.caller != "toile-client" || row.call.Usage.InputTokens != 7 || row.call.Usage.Cost != 0.002 {
		t.Fatalf("row %+v", row)
	}
}

func TestCompleteRecordsPartialUsageOnProviderError(t *testing.T) {
	ledger := &fakeLedger{}
	srv := completeServer(t, &fakeCompleter{err: context.DeadlineExceeded}, nil, withLedger(ledger))
	if res, _ := complete(t, srv, toileWrite, `{"prompt":"hi"}`); res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status %d", res.StatusCode)
	}
	if len(ledger.rows) != 1 {
		t.Fatalf("rows %+v", ledger.rows)
	}
}

func TestCompleteRefusedByBudgetRecordsNothing(t *testing.T) {
	refusal := troncerrors.New("rate_limited", "This month's AI allowance is used up. It resets on 1 December.", aiusage.ErrBudget)
	ledger := &fakeLedger{}
	srv := completeServer(t, &fakeCompleter{err: refusal}, nil, withLedger(ledger))
	res, raw := complete(t, srv, toileWrite, `{"prompt":"hi"}`)
	if res.StatusCode != http.StatusTooManyRequests || len(ledger.rows) != 0 {
		t.Fatalf("status %d, rows %+v: %s", res.StatusCode, ledger.rows, raw)
	}
}
