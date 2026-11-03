package suite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/FacileStudio/nacelle"
)

func runTool(t *testing.T, tools []nacelle.Tool, name, input string) (string, error) {
	t.Helper()
	for _, tool := range tools {
		if tool.Name() == name {
			return tool.Run(context.Background(), json.RawMessage(input))
		}
	}
	t.Fatalf("tool %s is not registered", name)
	return "", nil
}

func TestToolsRegisterOnlyConfiguredApps(t *testing.T) {
	w := newWorld(t)
	w.cfg.Apps[AppOpus] = AppConfig{}
	w.cfg.Apps[AppSablier] = AppConfig{}
	tools, note, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{})
	if err != nil || len(tools) != 15 {
		t.Fatalf("want 15 tools, got %d, %v", len(tools), err)
	}
	for _, name := range []string{"courrier_inbox", "courrier_search", "courrier_accounts", "courrier_send", "agenda_events", "agenda_calendars", "agenda_delete_event"} {
		if !strings.Contains(note, name) {
			t.Errorf("prompt must list %s: %q", name, note)
		}
	}
	if strings.Contains(note, "opus_tasks") || strings.Contains(note, "sablier_status") {
		t.Errorf("prompt lists a tool that is not registered: %q", note)
	}
}

func TestEveryMutatingToolNeedsApproval(t *testing.T) {
	w := newWorld(t)
	tools, _, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{})
	if err != nil {
		t.Fatal(err)
	}
	gated := map[string]bool{}
	for _, name := range WriteTools() {
		gated[name] = true
	}
	for _, tool := range tools {
		readOnly, ok := tool.(nacelle.ReadOnlyTool)
		mutates := !ok || !readOnly.IsReadOnly()
		if mutates != gated[tool.Name()] {
			t.Errorf("%s: mutates=%t but approval gated=%t", tool.Name(), mutates, gated[tool.Name()])
		}
		delete(gated, tool.Name())
	}
	for name := range gated {
		t.Errorf("%s is gated but never registered", name)
	}
}

func TestToolsNoneConfigured(t *testing.T) {
	w := newWorld(t)
	for _, name := range appNames {
		w.cfg.Apps[name] = AppConfig{}
	}
	tools, note, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{})
	if err != nil || len(tools) != 0 || note != "" {
		t.Fatalf("want nothing, got %d tools, %q, %v", len(tools), note, err)
	}
}

func assertOnlyOwnData(t *testing.T, tools []nacelle.Tool, user int64, name, marker string) {
	t.Helper()
	out, err := runTool(t, tools, name, `{}`)
	want, other := fmt.Sprintf("%s %d", marker, user), fmt.Sprintf("%s %d", marker, 3-user)
	if err != nil || !strings.Contains(out, want) || strings.Contains(out, other) {
		t.Errorf("%s as user %d: %q, %v", name, user, out, err)
	}
}

func TestToolsNeverCrossUsers(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.addUser(2, future)
	gw := NewGateway(w.tokens(), w.cfg, nil)
	checks := map[string]string{
		"courrier_inbox": "mail of", "agenda_events": "event of", "opus_tasks": "task of", "sablier_status": "timer of",
	}
	var wg sync.WaitGroup
	for _, user := range []int64{1, 2} {
		tools, _, err := Tools(gw, user, Defaults{})
		if err != nil {
			t.Fatal(err)
		}
		for name, marker := range checks {
			wg.Go(func() { assertOnlyOwnData(t, tools, user, name, marker) })
		}
	}
	wg.Wait()
}

func TestToolOutputIsProjected(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	tools, _, _ := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{})
	accounts, _ := runTool(t, tools, "courrier_accounts", `{}`)
	events, _ := runTool(t, tools, "agenda_events", `{"from":"2026-11-02"}`)
	tasks, _ := runTool(t, tools, "opus_tasks", `{"project":"p"}`)
	if strings.Contains(accounts, "imap_host") || strings.Contains(events, "etag") || strings.Contains(tasks, "done") {
		t.Fatalf("hidden fields leaked:\n%s\n%s\n%s", accounts, events, tasks)
	}
	mine, _ := runTool(t, tools, "opus_tasks", `{}`)
	done, _ := runTool(t, tools, "opus_tasks", `{"status":"done"}`)
	if strings.Contains(mine, "done thing") || !strings.Contains(done, "done thing") || strings.Contains(done, "task of") {
		t.Fatalf("my-tasks default must hide done work unless a status is asked: %s | %s", mine, done)
	}
	if !strings.Contains(tasks, `"project":"P"`) || !strings.Contains(tasks, `"assignee":"A"`) {
		t.Fatalf("task projection lost fields: %s", tasks)
	}
}

func TestToolReportsReconnect(t *testing.T) {
	w := newWorld(t)
	tools, _, _ := Tools(NewGateway(w.tokens(), w.cfg, nil), 99, Defaults{})
	_, err := runTool(t, tools, "courrier_inbox", `{}`)
	if err == nil || !strings.Contains(err.Error(), "sign in again") {
		t.Fatalf("want a reconnect message, got %v", err)
	}
}

func TestClipTruncates(t *testing.T) {
	long := strings.Repeat("a", maxToolOutput+10)
	if out := clip(long); !strings.Contains(out, "truncated") || len(out) > maxToolOutput+100 {
		t.Fatalf("clip did not truncate: %d", len(out))
	}
}
