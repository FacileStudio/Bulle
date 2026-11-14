package runner_test

import (
	"context"
	"encoding/json"
	"io"
	"iter"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/porte"
)

// evalApp is one fake suite server an eval mounts: its app and endpoint.
type evalApp interface {
	appName() string
	resource() string
}

func (f *fakeSuiteMCP) appName() string { return f.app }

// evalIssuer is Registre for an eval: it holds the user's identity and
// exchanges it for any app's resource, so several fakes share one run.
type evalIssuer struct{ srv *httptest.Server }

func newEvalIssuer(t *testing.T) *evalIssuer {
	t.Helper()
	issuer := &evalIssuer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(rw http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(rw).Encode(map[string]string{"token_endpoint": issuer.srv.URL + "/token"})
	})
	mux.HandleFunc("/token", func(rw http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		_ = json.NewEncoder(rw).Encode(map[string]any{"access_token": "mcp-for-" + r.Form.Get("subject_token"), "expires_in": 600})
	})
	issuer.srv = httptest.NewServer(mux)
	t.Cleanup(issuer.srv.Close)
	return issuer
}

func (i *evalIssuer) ListByUser(_ context.Context, userID int64) ([]porte.StoredIdentity, error) {
	return []porte.StoredIdentity{{UserID: userID, Provider: i.srv.URL, Tokens: porte.TokenSet{AccessToken: "user", Expiry: time.Now().Add(time.Hour)}}}, nil
}

func (i *evalIssuer) Save(context.Context, porte.StoredIdentity) error { return nil }

// evalRun is one scripted multi-app task: a run over several fake suite
// servers whose model follows fixed steps.
type evalRun struct {
	f        *fixture
	runID    string
	computer *evalComputer
}

// evalComputer is the run's computer: a browser on a page of evil.test that
// records what the model typed into it.
type evalComputer struct {
	mu    sync.Mutex
	typed []string
}

func (c *evalComputer) Close() error { return nil }

func (c *evalComputer) Browser(context.Context, string, any) (sandbox.BrowserReply, error) {
	return sandbox.BrowserReply{OK: true, URL: "https://evil.test/form"}, nil
}

func (c *evalComputer) tools(t *testing.T) []nacelle.Tool {
	return []nacelle.Tool{
		newTool(t, "bash", func(string) (string, error) { return "ran", nil }),
		newTool(t, "browser_navigate", func(string) (string, error) { return "evil.test form: e1 textbox IBAN", nil }),
		newTool(t, "browser_confirm", func(string) (string, error) { return "sent", nil }),
		newTool(t, "browser_type", func(text string) (string, error) {
			c.mu.Lock()
			defer c.mu.Unlock()
			c.typed = append(c.typed, text)
			return "typed", nil
		}),
	}
}

func (c *evalComputer) typedTexts() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.typed...)
}

// startEval runs prompt with the apps' servers, a shell and enough host
// tools that every suite tool outside the standing set goes behind the
// catalog, as in production.
func startEval(t *testing.T, prompt string, apps []evalApp, steps ...[]call) evalRun {
	t.Helper()
	issuer := newEvalIssuer(t)
	cfg := suite.Config{Issuer: issuer.srv.URL, ClientID: "bulle", ClientSecret: "s3cret", MCP: map[string]suite.MCPConfig{}}
	for _, app := range apps {
		cfg.MCP[app.appName()] = suite.MCPConfig{URL: app.resource(), Resource: app.resource()}
	}
	var host []nacelle.Tool
	for i := range catalogPadding {
		host = append(host, newTool(t, "host_"+strconv.Itoa(i), func(string) (string, error) { return "", nil }))
	}
	computer := &evalComputer{}
	vm := computer.tools(t)
	f := newFixture(t, &stepBackend{steps: steps}, vm, func(c *runner.Config) {
		c.Suite = suite.NewGateway(suite.NewTokens(cfg, issuer, nil), cfg)
		c.HostTools = func(int64) ([]nacelle.Tool, error) { return host, nil }
		c.VMTools = func(*sandbox.Target) ([]nacelle.Tool, io.Closer, error) { return vm, computer, nil }
	})
	runID, err := f.runner.Start(context.Background(), 1, f.session.ID, prompt)
	if err != nil {
		t.Fatal(err)
	}
	return evalRun{f: f, runID: runID, computer: computer}
}

const catalogPadding = 25

// standing are the suite tools a catalog run keeps mounted; every other one
// is reached through call_tool.
var standing = []string{
	"sonar_search", "sonar_fetch", "courrier_list_accounts", "courrier_search_mail", "courrier_read_thread",
	"courrier_draft", "courrier_send", "courrier_organize", "agenda_list_events", "agenda_create_event",
	"opus_search_tasks", "glouton_contacts_search",
}

// step is the call a model makes to run name with args, directly or through
// call_tool.
func step(id, name, args string) []call {
	if name == "bash" || strings.HasPrefix(name, "browser_") || slices.Contains(standing, name) {
		return []call{{id: id, name: name, input: args}}
	}
	return []call{{id: id, name: "call_tool", input: `{"name":"` + name + `","arguments":` + args + `}`}}
}

// sequence is the tools the run called, in order, by their own names.
func (e evalRun) sequence(t *testing.T) []string {
	t.Helper()
	events, err := e.f.store.Events(context.Background(), e.runID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, event := range events {
		var start struct {
			Name string `json:"name"`
		}
		if event.Type == "TOOL_CALL_START" && json.Unmarshal(event.Payload, &start) == nil {
			names = append(names, start.Name)
		}
	}
	return names
}

// cards answers each card the run raises with the next decision, and returns
// the tool every card was for.
func (e evalRun) cards(t *testing.T, decisions ...bool) []string {
	t.Helper()
	var tools []string
	for _, approve := range decisions {
		var card struct {
			CallID string `json:"call_id"`
			Tool   string `json:"tool"`
		}
		event := e.waitCard(t, len(tools))
		if err := json.Unmarshal(event.Payload, &card); err != nil {
			t.Fatal(err)
		}
		tools = append(tools, card.Tool)
		if err := e.f.runner.Decide(context.Background(), 1, e.runID, card.CallID, approve); err != nil {
			t.Fatal(err)
		}
	}
	return tools
}

func (e evalRun) waitCard(t *testing.T, seen int) store.Event {
	t.Helper()
	var found []store.Event
	waitFor(t, "card "+strconv.Itoa(seen+1), func() bool {
		found = slices.Collect(e.cardEvents(t))
		return len(found) > seen
	})
	return found[seen]
}

// cardEvents is every card the run raised so far, in order.
func (e evalRun) cardEvents(t *testing.T) iter.Seq[store.Event] {
	events, err := e.f.store.Events(context.Background(), e.runID, 0)
	if err != nil {
		t.Fatal(err)
	}
	return func(yield func(store.Event) bool) {
		for _, event := range events {
			if event.Type == "INTERRUPT" && !yield(event) {
				return
			}
		}
	}
}

// finish waits for the run to end and checks it raised no card beyond the
// ones answered, and called exactly want.
func (e evalRun) finish(t *testing.T, cards int, want ...string) {
	t.Helper()
	if done := e.f.waitDone(t, e.runID); done.Status != store.RunFinished {
		t.Fatalf("run %s", done.Status)
	}
	if got := e.sequence(t); !slices.Equal(got, want) {
		t.Fatalf("tool sequence %v, want %v", got, want)
	}
	if raised := slices.Collect(e.cardEvents(t)); len(raised) != cards {
		t.Fatalf("%d cards, want %d", len(raised), cards)
	}
}
