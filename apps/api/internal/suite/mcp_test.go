package suite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/FacileStudio/nacelle"
)

func TestThrottledSendsARefusedCallAgain(t *testing.T) {
	waits := mcpThrottleWaits
	mcpThrottleWaits = []time.Duration{time.Millisecond, time.Millisecond}
	t.Cleanup(func() { mcpThrottleWaits = waits })
	refused := errors.New("calling glouton_contacts_search: rejected: Too Many Requests")
	calls := 0
	result, err := throttled(context.Background(), func() (nacelle.Result, error) {
		calls++
		if calls < 3 {
			return nacelle.Result{}, refused
		}
		return nacelle.Result{Text: "ok"}, nil
	})
	if err != nil || result.Text != "ok" || calls != 3 {
		t.Fatalf("a 429 must be sent again until it passes: calls=%d, %v", calls, err)
	}
	calls = 0
	if _, err := throttled(context.Background(), func() (nacelle.Result, error) { calls++; return nacelle.Result{}, refused }); !errors.Is(err, refused) || calls != 3 {
		t.Fatalf("the retries must stop after the last wait: calls=%d, %v", calls, err)
	}
	calls = 0
	other := errors.New("calling glouton_company_get: not found")
	if _, err := throttled(context.Background(), func() (nacelle.Result, error) { calls++; return nacelle.Result{}, other }); !errors.Is(err, other) || calls != 1 {
		t.Fatalf("any other failure must not be sent again: calls=%d, %v", calls, err)
	}
}
