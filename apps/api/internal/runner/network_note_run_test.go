package runner_test

import (
	"context"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
)

func TestTheSystemPromptNamesTheComputersNetwork(t *testing.T) {
	for network, want := range map[string]string{
		"allowlist": "limited to a list of hosts",
		"offline":   "has no network",
		"open":      "",
	} {
		backend := &promptBackend{}
		f := newFixture(t, backend, nil, func(c *runner.Config) {
			c.Computers = networkComputers{fakeComputers: &fakeComputers{}, mode: network}
		})
		runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "hello")
		if err != nil {
			t.Fatal(err)
		}
		f.waitDone(t, runID)
		system := backend.prompt()
		if want == "" && (strings.Contains(system, "limited to a list") || strings.Contains(system, "no network")) {
			t.Errorf("%s: an open computer got a network note: %q", network, system)
		}
		if want != "" && !strings.Contains(system, want) {
			t.Errorf("%s: the prompt misses %q: %q", network, want, system)
		}
	}
}
