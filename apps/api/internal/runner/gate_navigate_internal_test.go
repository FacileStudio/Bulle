package runner

import (
	"context"
	"testing"

	"github.com/FacileStudio/nacelle"
)

func TestALinkRelativeToAFetchedPageCountsAsSeen(t *testing.T) {
	g := mailTaintedGate()
	page := `{"final_url":"https://alpinelinux.org/releases/","markdown":"[3.24.4](/posts/alpine-3.24.4-released/) [notes](../notes/x \"t\") [abs](https://go.dev/)"}`
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: sonarFetchTool}, Result: page})
	for _, link := range []string{"https://alpinelinux.org/posts/alpine-3.24.4-released/", "https://alpinelinux.org/notes/x", "https://go.dev/"} {
		if g.need(sonarFetchTool, urlInput(link), false) != askNo {
			t.Errorf("%s is linked from the fetched page and must not ask", link)
		}
	}
	if g.need(sonarFetchTool, urlInput("https://alpinelinux.org/posts/?d=FR76"), false) != askCard {
		t.Fatal("a URL the page never links must still ask")
	}
}
