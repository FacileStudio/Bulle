package runner_test

import (
	"errors"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/nacelle"
)

func TestRunWhoseWriteFailedIsNotBooked(t *testing.T) {
	sablier := newFakeSablier(t, false)
	f := mailFixture(t, sablier, func(string) (string, error) { return "", errors.New("smtp down") })
	f.runApproved(t, "mail bob", "m1")
	time.Sleep(100 * time.Millisecond)
	if posts := len(sablier.sent(bookSession)); posts != 0 {
		t.Fatalf("want no agent session for a run that changed nothing, got %d", posts)
	}
}

func mailFixture(t *testing.T, sablier *fakeSablier, send func(string) (string, error)) *fixture {
	t.Helper()
	backend := &fakeBackend{calls: []call{{id: "m1", name: "send_mail", input: `{"text":"to bob"}`}}}
	return newFixture(t, backend, []nacelle.Tool{newTool(t, "send_mail", send)}, func(c *runner.Config) {
		sablier.tune(c)
		c.Outward = []string{"send_mail"}
	})
}
