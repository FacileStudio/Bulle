package ide

import (
	"context"
	"testing"

	"github.com/FacileStudio/nacelle"
)

func TestTwoCallsOfOneToolKeepTheirOwnIDs(t *testing.T) {
	srv := startTestServer(t, Options{Root: t.TempDir()})
	_, dec := attach(t, srv)
	hooks := srv.Hooks()
	call := func(point nacelle.HookPoint, id string) map[string]any {
		hook := hooks[point][0]
		hook(context.Background(), nacelle.HookEvent{Point: point, Call: nacelle.Call{ID: id, Name: "bash"}})
		return readEvent(t, dec)
	}
	call(nacelle.BeforeToolCall, "a")
	call(nacelle.BeforeToolCall, "b")
	if done := call(nacelle.AfterToolCall, "a"); done["id"] != "a" {
		t.Fatalf("the first call's finish carries %v, want a", done["id"])
	}
	if done := call(nacelle.AfterToolCall, "b"); done["id"] != "b" {
		t.Fatalf("the second call's finish carries %v, want b", done["id"])
	}
}
