package runner_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"iter"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/core/nacelletest"
	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/nacelle"
)

type call struct {
	id    string
	name  string
	input string
}

type fakeBackend struct {
	calls []call
	block chan struct{}
}

func (f *fakeBackend) Name() string { return "fake" }

func (f *fakeBackend) Capabilities() nacelle.Capabilities {
	return nacelle.Capabilities{Thinking: true}
}

func (f *fakeBackend) CountTokens(context.Context, nacelle.Request) (int64, error) { return 0, nil }

func (f *fakeBackend) Stream(ctx context.Context, req nacelle.Request) iter.Seq2[nacelle.Event, error] {
	return nacelle.NewBackend(&fakeRun{fakeBackend: f}).Stream(ctx, req)
}

// fakeRun is one run of a fakeBackend: a first turn that asks for its calls,
// waiting on block first when it is set, then a turn that answers.
type fakeRun struct {
	*fakeBackend
	turns int
}

func (r *fakeRun) Turn(ctx context.Context, _ nacelle.Request) iter.Seq2[nacelle.Event, error] {
	first := r.turns == 0
	r.turns++
	turn := nacelletest.Turn{Text: "All done.", Usage: nacelle.Usage{InputTokens: 10, OutputTokens: 5, Cost: 0.002}}
	if first && len(r.calls) > 0 {
		turn = nacelletest.Turn{Thinking: "hm", Text: "Let me look. "}
		for _, c := range r.calls {
			turn.Calls = append(turn.Calls, nacelle.ToolCall{ID: c.id, Name: c.name, Input: json.RawMessage(c.input)})
		}
	}
	return func(yield func(nacelle.Event, error) bool) {
		if first && !r.wait(ctx, yield) {
			return
		}
		for _, event := range turn.Events() {
			if !yield(event, nil) {
				return
			}
		}
	}
}

func (r *fakeRun) wait(ctx context.Context, yield func(nacelle.Event, error) bool) bool {
	if r.block == nil {
		return true
	}
	if !yield(nacelle.Event{Kind: nacelle.KindText, Text: "waiting"}, nil) {
		return false
	}
	select {
	case <-r.block:
		return true
	case <-ctx.Done():
		yield(nacelle.Event{}, ctx.Err())
		return false
	}
}

type fakeComputers struct {
	ensured atomic.Int32
	touched atomic.Int32
	lastID  atomic.Value
}

func (f *fakeComputers) Ensure(_ context.Context, userID string) (*sandbox.Target, error) {
	f.ensured.Add(1)
	f.lastID.Store(userID)
	return &sandbox.Target{Name: "vm-" + userID, Workdir: "/workspace"}, nil
}

func (f *fakeComputers) Touch(string) { f.touched.Add(1) }

type noClose struct{}

func (noClose) Close() error { return nil }

type input struct {
	Text string `json:"text"`
}

func newTool(t *testing.T, name string, fn func(string) (string, error)) nacelle.Tool {
	t.Helper()
	tool, err := nacelle.NewTool(name, name, func(_ context.Context, in input) (string, error) { return fn(in.Text) })
	if err != nil {
		t.Fatal(err)
	}
	return tool
}

type fixture struct {
	runner    *runner.Runner
	store     *store.Memory
	computers *fakeComputers
	session   store.Session
}

func newFixture(t *testing.T, backend nacelle.Backend, vm []nacelle.Tool, tune func(*runner.Config)) *fixture {
	t.Helper()
	st := store.NewMemory()
	computers := &fakeComputers{}
	cfg := runner.Config{
		Store:     st,
		Computers: computers,
		Backends:  func(string, string) (nacelle.Backend, error) { return backend, nil },
		VMTools:   func(*sandbox.Target) ([]nacelle.Tool, io.Closer, error) { return vm, noClose{}, nil },
		HostTools: func(int64) ([]nacelle.Tool, error) { return nil, nil },
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		Redactor:  runner.NewRedactor("sk-secret-value"),
	}
	if tune != nil {
		tune(&cfg)
	}
	rn := runner.New(cfg)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		rn.Shutdown(ctx)
	})
	session, err := st.CreateSession(context.Background(), store.Session{ID: store.NewID(), UserID: 1, Channel: store.ChannelWeb})
	if err != nil {
		t.Fatal(err)
	}
	return &fixture{runner: rn, store: st, computers: computers, session: session}
}

func (f *fixture) newSession(t *testing.T, userID int64) store.Session {
	t.Helper()
	session, err := f.store.CreateSession(context.Background(), store.Session{ID: store.NewID(), UserID: userID, Channel: store.ChannelWeb})
	if err != nil {
		t.Fatal(err)
	}
	return session
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func (f *fixture) waitDone(t *testing.T, runID string) store.Run {
	t.Helper()
	var run store.Run
	waitFor(t, "run to end", func() bool {
		run, _ = f.store.GetRun(context.Background(), f.session.UserID, runID)
		return run.Status != "" && run.Status != store.RunRunning && !f.runner.IsLive(runID)
	})
	return run
}

func (f *fixture) eventTypes(t *testing.T, runID string) []string {
	t.Helper()
	events, err := f.store.Events(context.Background(), runID, 0)
	if err != nil {
		t.Fatal(err)
	}
	types := make([]string, len(events))
	for i, event := range events {
		if event.Seq != i+1 {
			t.Fatalf("event %d has seq %d", i, event.Seq)
		}
		types[i] = event.Type
	}
	return types
}

func (f *fixture) waitEvent(t *testing.T, runID, eventType string) store.Event {
	t.Helper()
	var found store.Event
	waitFor(t, eventType, func() bool {
		events, _ := f.store.Events(context.Background(), runID, 0)
		for _, event := range events {
			if event.Type == eventType {
				found = event
				return true
			}
		}
		return false
	})
	return found
}

func TestFullTurnPersistsTheEventSequence(t *testing.T) {
	echo := newTool(t, "echo", func(text string) (string, error) { return "echoed " + text, nil })
	backend := &fakeBackend{calls: []call{{id: "c1", name: "echo", input: `{"text":"hi"}`}}}
	f := newFixture(t, backend, []nacelle.Tool{echo}, nil)

	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "  look around  ")
	if err != nil {
		t.Fatal(err)
	}
	run := f.waitDone(t, runID)

	want := []string{"RUN_STARTED", "THINKING", "TEXT_MESSAGE_CONTENT", "TOOL_CALL_START", "TOOL_CALL_END", "TEXT_MESSAGE_CONTENT", "RUN_FINISHED"}
	got := f.eventTypes(t, runID)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("events\n got %v\nwant %v", got, want)
	}
	if run.Status != store.RunFinished || run.StopReason != "end" || run.CostUSD != 0.002 {
		t.Fatalf("run: %+v", run)
	}

	events, _ := f.store.Events(context.Background(), runID, 0)
	var start, end, finished map[string]any
	_ = json.Unmarshal(events[3].Payload, &start)
	_ = json.Unmarshal(events[4].Payload, &end)
	_ = json.Unmarshal(events[6].Payload, &finished)
	if start["call_id"] != "c1" || start["name"] != "echo" || start["args"].(map[string]any)["text"] != "hi" {
		t.Fatalf("TOOL_CALL_START payload: %v", start)
	}
	if end["call_id"] != "c1" || end["ok"] != true || end["output_preview"] != "echoed hi" {
		t.Fatalf("TOOL_CALL_END payload: %v", end)
	}
	if finished["stop"] != "end" || finished["cost_usd"] != 0.002 {
		t.Fatalf("RUN_FINISHED payload: %v", finished)
	}

	messages, _ := f.store.Messages(context.Background(), f.session.ID)
	roles := ""
	for _, m := range messages {
		role, text, _ := runner.MessageText(m.Payload)
		roles += role + ":" + text + "|"
	}
	if !strings.HasPrefix(roles, "user:look around|assistant:Let me look. ") || !strings.HasSuffix(roles, "All done.|") {
		t.Fatalf("transcript: %s", roles)
	}

	if id, _ := f.computers.lastID.Load().(string); id != "1" || f.computers.touched.Load() == 0 {
		t.Fatalf("computer binding: ensured for %q, touched %d times", id, f.computers.touched.Load())
	}
	session, _ := f.store.GetSession(context.Background(), 1, f.session.ID)
	if session.Title != "look around" {
		t.Fatalf("title %q", session.Title)
	}
}

func TestSecondTurnReplaysTheTranscript(t *testing.T) {
	backend := &fakeBackend{}
	f := newFixture(t, backend, nil, nil)
	for _, text := range []string{"one", "two"} {
		runID, err := f.runner.Start(context.Background(), 1, f.session.ID, text)
		if err != nil {
			t.Fatal(err)
		}
		f.waitDone(t, runID)
	}
	messages, _ := f.store.Messages(context.Background(), f.session.ID)
	if len(messages) != 4 {
		t.Fatalf("want 4 messages, got %d", len(messages))
	}
}

func TestOutwardCallWaitsForApproval(t *testing.T) {
	for _, approve := range []bool{true, false} {
		var sent atomic.Int32
		send := newTool(t, "send_mail", func(string) (string, error) { sent.Add(1); return "sent", nil })
		backend := &fakeBackend{calls: []call{{id: "m1", name: "send_mail", input: `{"text":"to bob"}`}}}
		f := newFixture(t, backend, []nacelle.Tool{send}, func(c *runner.Config) { c.Outward = []string{"send_mail"} })

		runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "mail bob")
		if err != nil {
			t.Fatal(err)
		}
		interrupt := f.waitEvent(t, runID, "INTERRUPT")
		var payload map[string]any
		_ = json.Unmarshal(interrupt.Payload, &payload)
		if payload["call_id"] != "m1" || payload["tool"] != "send_mail" {
			t.Fatalf("INTERRUPT payload: %v", payload)
		}

		time.Sleep(150 * time.Millisecond)
		if sent.Load() != 0 {
			t.Fatal("outward tool ran before approval")
		}
		if run, _ := f.store.GetRun(context.Background(), 1, runID); run.Status != store.RunRunning {
			t.Fatalf("run left running state while waiting: %s", run.Status)
		}

		if err := f.runner.Decide(context.Background(), 2, runID, "m1", approve); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("stranger decision: %v", err)
		}
		if err := f.runner.Decide(context.Background(), 1, runID, "nope", approve); !errors.Is(err, runner.ErrNoPendingApproval) {
			t.Fatalf("unknown call: %v", err)
		}
		if err := f.runner.Decide(context.Background(), 1, runID, "m1", approve); err != nil {
			t.Fatal(err)
		}
		if err := f.runner.Decide(context.Background(), 1, runID, "m1", approve); !errors.Is(err, runner.ErrNoPendingApproval) && !errors.Is(err, runner.ErrRunNotActive) {
			t.Fatalf("second decision: %v", err)
		}
		f.waitDone(t, runID)

		events, _ := f.store.Events(context.Background(), runID, 0)
		var end map[string]any
		for _, event := range events {
			if event.Type == "TOOL_CALL_END" {
				_ = json.Unmarshal(event.Payload, &end)
			}
		}
		if approve {
			if sent.Load() != 1 || end["ok"] != true {
				t.Fatalf("approved call: sent=%d end=%v", sent.Load(), end)
			}
		} else if sent.Load() != 0 || end["ok"] != false {
			t.Fatalf("denied call: sent=%d end=%v", sent.Load(), end)
		}
	}
}

type recordedPush struct {
	mu   sync.Mutex
	seen []push.Notification
}

func (r *recordedPush) Notify(_ context.Context, _ int64, n push.Notification) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seen = append(r.seen, n)
}

func (r *recordedPush) bodies() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []string{}
	for _, n := range r.seen {
		out = append(out, n.Body+"|"+n.Path)
	}
	return out
}

func TestApprovalPushesOnlyWhenNobodyWatches(t *testing.T) {
	for _, watched := range []bool{false, true} {
		sent := &recordedPush{}
		send := newTool(t, "send_mail", func(string) (string, error) { return "sent", nil })
		backend := &fakeBackend{calls: []call{{id: "m1", name: "send_mail", input: `{"text":"secret body"}`}}}
		f := newFixture(t, backend, []nacelle.Tool{send}, func(c *runner.Config) {
			c.Outward = []string{"send_mail"}
			c.Push = sent
		})
		session := f.session

		var stop func()
		runID, err := f.runner.Start(context.Background(), 1, session.ID, "mail bob")
		if err != nil {
			t.Fatal(err)
		}
		if watched {
			_, stop = f.runner.Watch(runID)
		}
		f.waitEvent(t, runID, "INTERRUPT")
		time.Sleep(100 * time.Millisecond)

		got := sent.bodies()
		if watched && len(got) != 0 {
			t.Fatalf("pushed while watched: %v", got)
		}
		if !watched && (len(got) != 1 || got[0] != "Approval needed|sessions/"+session.ID) {
			t.Fatalf("push = %v", got)
		}
		if err := f.runner.Decide(context.Background(), 1, runID, "m1", true); err != nil {
			t.Fatal(err)
		}
		f.waitDone(t, runID)
		if stop != nil {
			stop()
		}
	}
}

func TestVMToolsCarryNoGate(t *testing.T) {
	var ran atomic.Int32
	vm := newTool(t, "bash", func(string) (string, error) { ran.Add(1); return "ok", nil })
	backend := &fakeBackend{calls: []call{{id: "v1", name: "bash", input: `{"text":"ls"}`}}}
	f := newFixture(t, backend, []nacelle.Tool{vm}, func(c *runner.Config) { c.Outward = []string{"send_mail"} })
	runID, _ := f.runner.Start(context.Background(), 1, f.session.ID, "ls")
	f.waitDone(t, runID)
	if ran.Load() != 1 {
		t.Fatal("vm tool did not run")
	}
	for _, kind := range f.eventTypes(t, runID) {
		if kind == "INTERRUPT" {
			t.Fatal("vm tool was gated")
		}
	}
}

func TestOneActiveRunPerSessionAndPerUserCap(t *testing.T) {
	backend := &fakeBackend{block: make(chan struct{})}
	f := newFixture(t, backend, nil, func(c *runner.Config) { c.MaxRunsPerUser = 1 })

	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "first")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.runner.Start(context.Background(), 1, f.session.ID, "again"); !errors.Is(err, runner.ErrTooManyRuns) {
		t.Fatalf("cap: %v", err)
	}
	other := f.newSession(t, 2)
	otherRun, err := f.runner.Start(context.Background(), 2, other.ID, "other user")
	if err != nil {
		t.Fatalf("another user must not share the cap: %v", err)
	}

	if err := f.runner.Stop(context.Background(), 2, runID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger stop: %v", err)
	}
	if err := f.runner.Stop(context.Background(), 1, runID); err != nil {
		t.Fatal(err)
	}
	run := f.waitDone(t, runID)
	if run.Status != store.RunStopped {
		t.Fatalf("stopped run: %+v", run)
	}
	types := f.eventTypes(t, runID)
	if types[len(types)-1] != "RUN_FINISHED" {
		t.Fatalf("events: %v", types)
	}
	events, _ := f.store.Events(context.Background(), runID, 0)
	var finished map[string]any
	_ = json.Unmarshal(events[len(events)-1].Payload, &finished)
	if finished["stop"] != "stopped" {
		t.Fatalf("stop payload: %v", finished)
	}

	_ = f.runner.Stop(context.Background(), 2, otherRun)
	close(backend.block)
}

func TestSessionBusyWhenARunIsActive(t *testing.T) {
	backend := &fakeBackend{block: make(chan struct{})}
	f := newFixture(t, backend, nil, func(c *runner.Config) { c.MaxRunsPerUser = 5 })
	if _, err := f.runner.Start(context.Background(), 1, f.session.ID, "first"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.runner.Start(context.Background(), 1, f.session.ID, "second"); !errors.Is(err, store.ErrSessionBusy) {
		t.Fatalf("busy: %v", err)
	}
	close(backend.block)
}

func TestShutdownMarksLiveRunsInterrupted(t *testing.T) {
	backend := &fakeBackend{block: make(chan struct{})}
	f := newFixture(t, backend, nil, nil)
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "long job")
	if err != nil {
		t.Fatal(err)
	}
	f.waitEvent(t, runID, "TEXT_MESSAGE_CONTENT")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	f.runner.Shutdown(ctx)

	run, _ := f.store.GetRun(context.Background(), 1, runID)
	if run.Status != store.RunInterrupted {
		t.Fatalf("run: %+v", run)
	}
	types := f.eventTypes(t, runID)
	if types[len(types)-1] != "RUN_FINISHED" {
		t.Fatalf("events: %v", types)
	}
	if _, err := f.runner.Start(context.Background(), 1, f.session.ID, "late"); !errors.Is(err, runner.ErrShuttingDown) {
		t.Fatalf("start after shutdown: %v", err)
	}
	messages, _ := f.store.Messages(context.Background(), f.session.ID)
	if len(messages) < 2 {
		t.Fatalf("partial transcript not kept: %d messages", len(messages))
	}
}

func TestRunTimeoutEndsWithAnError(t *testing.T) {
	backend := &fakeBackend{block: make(chan struct{})}
	f := newFixture(t, backend, nil, func(c *runner.Config) { c.RunTimeout = 80 * time.Millisecond })
	runID, _ := f.runner.Start(context.Background(), 1, f.session.ID, "slow")
	run := f.waitDone(t, runID)
	if run.Status != store.RunError || run.StopReason != "timeout" {
		t.Fatalf("run: %+v", run)
	}
	types := f.eventTypes(t, runID)
	if types[len(types)-1] != "RUN_ERROR" {
		t.Fatalf("events: %v", types)
	}
}

func TestProviderKeysNeverLeakIntoEvents(t *testing.T) {
	leak := newTool(t, "bash", func(string) (string, error) { return "KEY=sk-secret-value", nil })
	backend := &fakeBackend{calls: []call{{id: "k1", name: "bash", input: `{"text":"env"}`}}}
	f := newFixture(t, backend, []nacelle.Tool{leak}, nil)
	runID, _ := f.runner.Start(context.Background(), 1, f.session.ID, "env")
	f.waitDone(t, runID)
	events, _ := f.store.Events(context.Background(), runID, 0)
	for _, event := range events {
		if strings.Contains(string(event.Payload), "sk-secret-value") {
			t.Fatalf("secret in %s: %s", event.Type, event.Payload)
		}
	}
	messages, _ := f.store.Messages(context.Background(), f.session.ID)
	for _, m := range messages {
		if strings.Contains(string(m.Payload), "sk-secret-value") {
			t.Fatalf("secret persisted: %s", m.Payload)
		}
	}
}

func TestComputerFailureIsAToolError(t *testing.T) {
	vm := newTool(t, "vm_echo", func(text string) (string, error) { return "vm " + text, nil })
	backend := &fakeBackend{calls: []call{{id: "c1", name: "vm_echo", input: `{"text":"a"}`}}}
	f := newFixture(t, backend, []nacelle.Tool{vm}, func(c *runner.Config) { c.Computers = nil })
	runID, _ := f.runner.Start(context.Background(), 1, f.session.ID, "hello")
	run := f.waitDone(t, runID)
	if run.Status != store.RunFinished {
		t.Fatalf("run: %+v", run)
	}
	events, _ := f.store.Events(context.Background(), runID, 0)
	for _, event := range events {
		if event.Type != "TOOL_CALL_END" {
			continue
		}
		var end map[string]any
		_ = json.Unmarshal(event.Payload, &end)
		if end["ok"] != false || !strings.Contains(end["output_preview"].(string), "your computer is not available") {
			t.Fatalf("TOOL_CALL_END payload: %v", end)
		}
		return
	}
	t.Fatal("no TOOL_CALL_END event")
}

// silentBackend answers with nothing at all, the way a run that ends before the
// model speaks does.
type silentBackend struct{}

func (silentBackend) Name() string                                                { return "silent" }
func (silentBackend) Capabilities() nacelle.Capabilities                          { return nacelle.Capabilities{} }
func (silentBackend) CountTokens(context.Context, nacelle.Request) (int64, error) { return 0, nil }

func (silentBackend) Stream(context.Context, nacelle.Request) iter.Seq2[nacelle.Event, error] {
	return func(yield func(nacelle.Event, error) bool) {
		yield(nacelle.Event{Kind: nacelle.KindDone, Stop: nacelle.StopEnd}, nil)
	}
}

func TestTurnWithoutAnAnswerKeepsTheTranscriptAlternating(t *testing.T) {
	f := newFixture(t, silentBackend{}, nil, nil)
	for _, text := range []string{"first", "second"} {
		runID, err := f.runner.Start(context.Background(), 1, f.session.ID, text)
		if err != nil {
			t.Fatal(err)
		}
		f.waitDone(t, runID)
	}
	messages, _ := f.store.Messages(context.Background(), f.session.ID)
	roles := make([]string, len(messages))
	for i, m := range messages {
		role, text, _ := runner.MessageText(m.Payload)
		if role == "assistant" && text == "" {
			t.Fatalf("empty assistant message at %d", i)
		}
		roles[i] = role
	}
	if strings.Join(roles, ",") != "user,assistant,user,assistant" {
		t.Fatalf("roles = %v", roles)
	}
}

func TestRecoverClosesRunsLeftRunningByACrash(t *testing.T) {
	f := newFixture(t, &fakeBackend{}, nil, nil)
	ctx := context.Background()
	if err := f.store.CreateRun(ctx, store.Run{ID: "stale", SessionID: f.session.ID}); err != nil {
		t.Fatal(err)
	}
	_ = f.store.AppendEvent(ctx, store.Event{RunID: "stale", Seq: 1, Type: "RUN_STARTED", Payload: json.RawMessage(`{}`)})
	if err := f.runner.Recover(ctx); err != nil {
		t.Fatal(err)
	}
	if types := f.eventTypes(t, "stale"); strings.Join(types, ",") != "RUN_STARTED,RUN_FINISHED" {
		t.Fatalf("events: %v", types)
	}
	if run, _ := f.store.GetRun(ctx, 1, "stale"); run.Status != store.RunInterrupted {
		t.Fatalf("run: %+v", run)
	}
}

func TestEmptyMessageIsRefused(t *testing.T) {
	f := newFixture(t, &fakeBackend{}, nil, nil)
	if _, err := f.runner.Start(context.Background(), 1, f.session.ID, "   "); !errors.Is(err, runner.ErrEmptyMessage) {
		t.Fatalf("empty: %v", err)
	}
}

func TestStrangerCannotStartARunOnASession(t *testing.T) {
	f := newFixture(t, &fakeBackend{}, nil, nil)
	if _, err := f.runner.Start(context.Background(), 2, f.session.ID, "hello"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger start: %v", err)
	}
	if active, ok, _ := f.store.ActiveRun(context.Background(), f.session.ID); ok {
		t.Fatalf("a run was created: %+v", active)
	}
}

func TestHostOnlyRunNeverBootsTheComputer(t *testing.T) {
	host := newTool(t, "host_echo", func(text string) (string, error) { return "host " + text, nil })
	vm := newTool(t, "vm_echo", func(text string) (string, error) { return "vm " + text, nil })
	backend := &fakeBackend{calls: []call{{id: "c1", name: "host_echo", input: `{"text":"hi"}`}}}
	f := newFixture(t, backend, []nacelle.Tool{vm}, func(cfg *runner.Config) {
		cfg.HostTools = func(int64) ([]nacelle.Tool, error) { return []nacelle.Tool{host}, nil }
	})

	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "go")
	if err != nil {
		t.Fatal(err)
	}
	if run := f.waitDone(t, runID); run.Status != store.RunFinished {
		t.Fatalf("run: %+v", run)
	}
	if n := f.computers.ensured.Load(); n != 0 {
		t.Fatalf("Ensure called %d times for a host-only run", n)
	}
}

func TestVMToolsBootTheComputerOnce(t *testing.T) {
	vm := newTool(t, "vm_echo", func(text string) (string, error) { return "vm " + text, nil })
	backend := &fakeBackend{calls: []call{
		{id: "c1", name: "vm_echo", input: `{"text":"a"}`},
		{id: "c2", name: "vm_echo", input: `{"text":"b"}`},
	}}
	f := newFixture(t, backend, []nacelle.Tool{vm}, nil)

	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "go")
	if err != nil {
		t.Fatal(err)
	}
	if run := f.waitDone(t, runID); run.Status != store.RunFinished {
		t.Fatalf("run: %+v", run)
	}
	if n := f.computers.ensured.Load(); n != 1 {
		t.Fatalf("Ensure called %d times, want 1", n)
	}
}
