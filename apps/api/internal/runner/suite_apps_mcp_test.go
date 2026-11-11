package runner_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// suiteApps are the MCP servers Bulle reaches with no built-in client before
// them: one read each, and one action its server confirms.
var suiteApps = []suiteApp{
	{suite.AppPlume, "plume_list_documents", "plume_delete_document", suite.PlumeTools()},
	{suite.AppCapsule, "capsule_get_status", "capsule_revoke", suite.CapsuleTools()},
	{suite.AppJournal, "journal_search_logs", "journal_delete_alert", suite.JournalTools()},
	{suite.AppAntenne, "antenne_list_events", "antenne_replay_event", suite.AntenneTools()},
	{suite.AppVision, "vision_list_sites", "vision_delete_site", suite.VisionTools()},
	{suite.AppEcho, "echo_list_rooms", "echo_delete_room", suite.EchoTools()},
}

// fakeAppMCP is one suite app's server: its tools as it names them, every
// one answering at once except confirmed, which asks through MRTR first.
type fakeAppMCP struct {
	*fakeSuiteMCP

	mu       sync.Mutex
	ran      []string
	declined int
}

type appOut struct {
	Status string `json:"status"`
	Items  []any  `json:"items,omitempty"`
}

func newFakeAppMCP(t *testing.T, app, confirmed string, tools []string) *fakeAppMCP {
	t.Helper()
	f := &fakeAppMCP{}
	f.fakeSuiteMCP = newFakeSuiteMCP(t, app, func(server *sdk.Server) {
		for _, name := range tools {
			sensitive := name == confirmed
			sdk.AddTool(server, &sdk.Tool{Name: name, Annotations: hints(!sensitive, sensitive, true)}, f.handler(name, sensitive))
		}
	})
	return f
}

func (f *fakeAppMCP) handler(name string, sensitive bool) sdk.ToolHandlerFor[map[string]any, appOut] {
	return func(_ context.Context, req *sdk.CallToolRequest, _ map[string]any) (*sdk.CallToolResult, appOut, error) {
		if sensitive && req.Params.RequestState == "" {
			question := &sdk.ElicitParams{Mode: "form", Message: "Delete it for good?", RequestedSchema: confirmForm}
			return &sdk.CallToolResult{InputRequests: sdk.InputRequestMap{"confirm": question}, RequestState: "1.x.0.sig"}, appOut{}, nil
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		if sensitive {
			answer, _ := req.Params.InputResponses["confirm"].(*sdk.ElicitResult)
			if answer == nil || answer.Action != "accept" || answer.Content["confirm"] == false {
				f.declined++
				return nil, appOut{Status: "declined"}, nil
			}
		}
		f.ran = append(f.ran, name)
		return nil, appOut{Status: "ok", Items: []any{map[string]any{"id": 9, "name": injected}}}, nil
	}
}

func (f *fakeAppMCP) outcome() (ran []string, declined int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.ran...), f.declined
}

type suiteApp = struct {
	app, read, confirmed string
	tools                []string
}

func TestSuiteAppsReadAtOnceAndConfirmOnTheCard(t *testing.T) {
	for _, tc := range suiteApps {
		t.Run(tc.app, func(t *testing.T) { checkSuiteApp(t, tc) })
	}
}

func checkSuiteApp(t *testing.T, tc suiteApp) {
	server := newFakeAppMCP(t, tc.app, tc.confirmed, tc.tools)
	backend := &stepBackend{steps: [][]call{
		{{id: "r1", name: tc.read, input: `{}`}},
		{{id: "x1", name: tc.confirmed, input: `{"id":9}`}},
	}}
	f := newFixture(t, backend, nil, server.tune)
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "do it")
	if err != nil {
		t.Fatal(err)
	}
	if card := interruptCard(t, f, runID); card.Tool != tc.confirmed || card.Args["kind"] != "mcp" || card.Args["server"] != suite.Label(tc.app) {
		t.Fatalf("the card must show %s's own question: %+v", tc.app, card)
	}
	if err := f.runner.Decide(context.Background(), 1, runID, "x1", false); err != nil {
		t.Fatal(err)
	}
	if done := f.waitDone(t, runID); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	if ran, declined := server.outcome(); !slices.Equal(ran, []string{tc.read}) || declined != 1 {
		t.Fatalf("the read runs at once and a denied card reaches the server as a decline: ran=%v declined=%d", ran, declined)
	}
	if missing := slices.DeleteFunc(slices.Clone(tc.tools), func(name string) bool { return slices.Contains(backend.mounted(), name) }); len(missing) > 0 {
		t.Errorf("not mounted under their own names: %v", missing)
	}
	if exchanges := server.exchanged(); len(exchanges) == 0 || exchanges[0] != server.resource()+" "+tc.app+".read "+tc.app+".write" {
		t.Fatalf("the token must be exchanged for the resource: %v", exchanges)
	}
}

type mcpCard struct {
	Tool string         `json:"tool"`
	Args map[string]any `json:"args"`
}

func interruptCard(t *testing.T, f *fixture, runID string) mcpCard {
	t.Helper()
	var card mcpCard
	if err := json.Unmarshal(f.waitEvent(t, runID, "INTERRUPT").Payload, &card); err != nil {
		t.Fatal(err)
	}
	return card
}

func TestAnAppRegistreRefusesCostsTheRunNothingAndLogsNothing(t *testing.T) {
	journal := newFakeAppMCP(t, suite.AppJournal, "journal_delete_alert", suite.JournalTools())
	journal.deny = true
	var logs bytes.Buffer
	backend := &stepBackend{}
	f := newFixture(t, backend, nil, func(c *runner.Config) {
		journal.tune(c)
		c.Logger = slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelWarn}))
	})
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, "any errors?")
	if err != nil {
		t.Fatal(err)
	}
	if done := f.waitDone(t, runID); done.Status != store.RunFinished {
		t.Fatalf("a refused app must not cost the run: %s", done.Status)
	}
	if slices.ContainsFunc(backend.mounted(), func(name string) bool { return strings.HasPrefix(name, "journal_") }) {
		t.Fatalf("a person Registre refuses gets no Journal tools: %v", backend.mounted())
	}
	if logs.Len() != 0 {
		t.Fatalf("a refused app must log nothing: %s", logs.String())
	}
}

func TestAnAppRegistreRefusesIsUnavailableAndAskedOnce(t *testing.T) {
	journal := newFakeAppMCP(t, suite.AppJournal, "journal_delete_alert", suite.JournalTools())
	journal.deny = true
	var cfg runner.Config
	journal.tune(&cfg)
	for range 3 {
		statuses := cfg.Suite.Statuses(context.Background(), 1)
		if len(statuses) != 1 || statuses[0].State != suite.StateUnavailable {
			t.Fatalf("a refused app shows as unavailable: %+v", statuses)
		}
	}
	if asked := len(journal.exchanged()); asked != 1 {
		t.Fatalf("Bulle asks Registre once, then remembers the refusal: %d exchanges", asked)
	}
}

func TestAServerThatForbidsThePersonIsUnavailable(t *testing.T) {
	vision := newFakeAppMCP(t, suite.AppVision, "vision_delete_site", suite.VisionTools())
	vision.forbid = true
	var cfg runner.Config
	vision.tune(&cfg)
	statuses := cfg.Suite.Statuses(context.Background(), 1)
	if len(statuses) != 1 || statuses[0].State != suite.StateUnavailable {
		t.Fatalf("a server keeping the person out shows as unavailable: %+v", statuses)
	}
	tools, err := cfg.Suite.MCP().Tools(context.Background(), 1)
	if err != nil || len(tools) != 0 {
		t.Fatalf("a forbidden app is left out silently: %d tools, %v", len(tools), err)
	}
}
