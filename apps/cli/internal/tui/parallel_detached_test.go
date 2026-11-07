package tui

import (
	"testing"
	"time"

	"github.com/FacileStudio/nacelle"
)

// The fan-out goroutine never touches the model: it ends with a done marker on
// the detached channel and the update loop drops the cancel func. Run under
// -race, the old in-goroutine delete fails here.
func TestDetachedFanOutDropsCancelOnTheLoop(t *testing.T) {
	m := sized()
	m.agent = answering(t)
	m.delegate = nacelle.Config{Backend: silent{}, System: "s", MaxIterations: 1}
	defer func() {
		if m.run.cancel != nil {
			m.run.cancel()
		}
	}()
	m.launchDetached([]string{"one"})
	awaitDetached(t, m, "detach1")
	if _, ok := m.parallelCancels["detach1"]; ok {
		t.Fatal("cancel func survived the done marker")
	}
}

// awaitDetached feeds the detached channel to the update loop until the batch's
// done marker lands.
func awaitDetached(t *testing.T, m *Model, id string) {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case r := <-detached:
			m.recordDetached(r)
			if r.batch == id && r.done {
				return
			}
		case <-deadline:
			t.Fatal("no done marker for the fan-out")
		}
	}
}

// The user's concurrency cap wins; zero falls back to the default.
func TestConcurrencyCapHonoursTheUser(t *testing.T) {
	if got := concurrencyCap(3); got != 3 {
		t.Errorf("concurrencyCap(3) = %d, want 3", got)
	}
	if got := concurrencyCap(0); got != defaultConcurrency {
		t.Errorf("concurrencyCap(0) = %d, want %d", got, defaultConcurrency)
	}
}
