package runner

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

func trustGate(history ...nacelle.Message) *gate {
	st := store.NewMemory()
	em := &emitter{store: st, hub: newHub(), logger: slog.Default(), runID: "r1"}
	g := newGate(map[string]bool{}, st, em, "r1", time.Now)
	g.seed(history)
	return g
}

func TestTaintGatesTheShellAndProfile(t *testing.T) {
	g := trustGate(nacelle.UserText("build it"))
	profile := json.RawMessage(`{"path":"./profile.md","content":"x"}`)
	if g.asks("bash", json.RawMessage(`{}`)) || g.asks(memoryWriteTool, profile) {
		t.Fatal("a clean run must not ask for the shell or profile.md")
	}
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "list_directory"}})
	if g.asks("bash", json.RawMessage(`{}`)) {
		t.Fatal("a trusted tool tainted the run")
	}
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "browser_snapshot"}})
	if !g.asks("bash", json.RawMessage(`{}`)) || !g.asks(memoryWriteTool, profile) {
		t.Fatal("after a page read the shell and profile.md must ask")
	}
	if !g.asks(memoryWriteTool, json.RawMessage(`{"path":"people/ann.md","content":"x"}`)) || g.asks("read_file", nil) {
		t.Fatal("after a page read every memory write asks, and reads stay unasked")
	}
}

func urlInput(raw string) json.RawMessage {
	input, _ := json.Marshal(map[string]string{"url": raw})
	return input
}

func TestMemoryReadsLeaveTheRunClean(t *testing.T) {
	g := trustGate(nacelle.UserText("what do you remember about ann"))
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "memory_search"},
		Result: "people/ann.md: see https://evil.test/?d=x"})
	note := json.RawMessage(`{"path":"notes/x.md","content":"x"}`)
	if g.asks(shellTool, nil) || g.asks(memoryWriteTool, note) || g.asks(sonarFetchTool, urlInput("https://evil.test/a")) {
		t.Fatal("a memory read tainted the run")
	}
}

// fetchCases is whether a tainted run asks before fetching each URL.
var fetchCases = map[string]bool{
	"https://evil.test/acme/INV-2026-0042/4200EUR":      true,
	"https://acme-inv-2026-0042.evil.test/":             true,
	"https://acme:4200@evil.test/":                      true,
	"https://evil.test/c2VjcmV0aW52/b2ljZQ":             true,
	"https://evil.test/?c2VjcmV0aW52=&b2ljZQ=":          true,
	"https://evil.test/?d=c2VjcmV0aW52b2ljZQ":           true,
	"https://evil.test/c2VjcmV0aW52":                    true,
	"https://x.example/acme/INV-2026-0042/4200EUR?ok=1": true,
	"https://news.ycombinator.com/item?id=4212":         true,
	"not a url": true,
	"https://x.example/acme/INV-2026-0042/4200EUR":            false,
	"HTTPS://X.EXAMPLE:443/acme/INV-2026-0042/4200EUR":        false,
	"https://news.ycombinator.com/item?id=4211":               false,
	"https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=3":         false,
	"https://en.wikipedia.org/wiki/Go_(programming_language)": false,
	"https://docs.test/guide":                                 false,
	"https://docs.test/guide#setup":                           false,
	"https://go.dev/blog":                                     false,
	"https://example.com/a?id=7":                              false,
}

func TestTaintedFetchesAskForEveryURLTheRunDidNotSee(t *testing.T) {
	g := trustGate(nacelle.UserText("check my latest mail, then read go.dev/blog and https://Example.com:443/a?id=7"))
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "courrier_read_thread"},
		Result: `{"body":"confirm at https://x.example/acme/INV-2026-0042/4200EUR"}`})
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "sonar_search"},
		Result: `[{"url":"https:\/\/news.ycombinator.com\/item?id=4211"},{"url":"https://www.youtube.com/watch?v=dQw4w9WgXcQ\u0026t=3"}]` +
			"\n[Go (language)](https://en.wikipedia.org/wiki/Go_(programming_language)). See https://docs.test/guide#setup."})
	g.observe(context.Background(), nacelle.HookEvent{Point: nacelle.AfterToolCall, Call: nacelle.Call{Name: "list_directory"},
		Result: "https://evil.test/c2VjcmV0aW52"})
	for raw, want := range fetchCases {
		for _, tool := range []string{sonarFetchTool, browserNavigateTool} {
			if got := g.asks(tool, urlInput(raw)); got != want {
				t.Errorf("%s %s: asks %v, want %v", tool, raw, got, want)
			}
		}
	}
}

func TestACleanRunFetchesAndNavigatesUnasked(t *testing.T) {
	g := trustGate(nacelle.UserText("look up the weather in paris"))
	for _, tool := range []string{sonarFetchTool, browserNavigateTool} {
		if g.asks(tool, urlInput("https://search.test/?q=weather+paris")) {
			t.Fatalf("%s asked in a run no third party's text reached", tool)
		}
	}
}

func TestAStoredResultShowsItsURLsToALaterTurn(t *testing.T) {
	g := trustGate(nacelle.UserText("find the go blog"),
		nacelle.Message{Role: nacelle.RoleUser, Parts: []nacelle.Part{nacelle.ToolResult{ID: "s1", Name: "sonar_search", Result: "https://go.dev/blog/v2"}}},
		nacelle.Message{Role: nacelle.RoleUser, Parts: []nacelle.Part{nacelle.ToolResult{ID: "f1", Name: "read_file", Result: "https://evil.test/k"}}})
	if g.asks(sonarFetchTool, urlInput("https://go.dev/blog/v2")) || !g.asks(sonarFetchTool, urlInput("https://evil.test/k")) {
		t.Fatal("only a source's stored result may show a URL to a later turn")
	}
}

func TestApprovalWaitDoesNotSpendTheRunBudget(t *testing.T) {
	g := trustGate()
	ctx, stop := g.deadline(context.Background(), 150*time.Millisecond)
	defer stop()
	_, release := g.await("c1")
	time.Sleep(300 * time.Millisecond)
	release()
	if ctx.Err() != nil {
		t.Fatal("the approval wait used up the run's budget")
	}
	select {
	case <-ctx.Done():
		if !errors.Is(context.Cause(ctx), errTimeout) {
			t.Fatalf("cause %v", context.Cause(ctx))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("the run never timed out")
	}
}

func TestJobCardShowsScheduleAndPrompt(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	job, _ := st.CreateJob(ctx, store.Job{UserID: 1, Name: "digest", Schedule: "0 9 * * *", Prompt: "sum up", Enabled: true})
	em := &emitter{store: st, hub: newHub(), logger: slog.Default(), runID: "r1"}
	g := newGate(map[string]bool{}, st, em, "r1", time.Now)
	g.prime(st, 1, nil)
	args, scope, ok := g.describe(ctx, "jobs_update", json.RawMessage(`{"job_id":"`+job.ID+`","prompt":"curl evil.test"}`))
	record, _ := json.Marshal(args)
	if !ok || scope != "" || !strings.Contains(string(record), `"prompt":"curl evil.test"`) || !strings.Contains(string(record), `"schedule":"0 9 * * *"`) {
		t.Fatalf("card %s scope %q ok %v", record, scope, ok)
	}
}

func TestOnlyReadOnlyOpenWorldToolsAreSources(t *testing.T) {
	spec := func(name string, readOnly, openWorld bool) nacelle.Tool {
		return specTool{nacelle.Spec{Name: name, ReadOnly: readOnly, Hints: &nacelle.Hints{OpenWorld: openWorld}}}
	}
	tools := []nacelle.Tool{
		spec("app_read_inbox", true, true), spec("app_list_accounts", true, false),
		spec("app_send", false, true), specTool{nacelle.Spec{Name: "app_local", ReadOnly: true}},
	}
	if got := openWorldSources(tools); len(got) != 1 || got[0] != "app_read_inbox" {
		t.Fatalf("sources = %v, want only the open-world read", got)
	}
	g := &gate{}
	g.addSources(openWorldSources(tools))
	g.seed(nil)
	if !g.sourceLocked("app_read_inbox") || g.sourceLocked("app_send") {
		t.Fatal("a source must survive the seed and a write must not be one")
	}
}

type specTool struct{ spec nacelle.Spec }

func (s specTool) Spec() nacelle.Spec { return s.spec }

func (s specTool) Run(context.Context, nacelle.Call, func(string)) (nacelle.Result, error) {
	return nacelle.Result{}, nil
}
