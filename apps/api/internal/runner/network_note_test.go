package runner

import (
	"strings"
	"testing"
)

func TestNetworkNoteSaysWhatTheComputerCannotDo(t *testing.T) {
	if got := networkNote(NetworkOpen, nil); got != "" {
		t.Fatalf("an open computer needs no note, got %q", got)
	}
	offline := networkNote(NetworkOffline, nil)
	if !strings.Contains(offline, "no network") {
		t.Fatalf("offline note = %q", offline)
	}
	allow := networkNote(NetworkAllowlist, []string{"api.example.com", "*.cdn.example.org"})
	for _, want := range []string{"limited to a list of hosts", "https", "api.example.com, *.cdn.example.org", "git@", "https:// URLs"} {
		if !strings.Contains(allow, want) {
			t.Errorf("allowlist note misses %q: %s", want, allow)
		}
	}
	if strings.Contains(networkNote(NetworkAllowlist, nil), "plus the hosts") {
		t.Error("no extra hosts must not mention them")
	}
}
