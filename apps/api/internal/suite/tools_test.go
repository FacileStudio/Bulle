package suite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/FacileStudio/Bulle/core/nacelletest"
	"github.com/FacileStudio/nacelle"
)

func runTool(t *testing.T, tools []nacelle.Tool, name, input string) (string, error) {
	t.Helper()
	for _, tool := range tools {
		if tool.Spec().Name == name {
			return nacelletest.Run(context.Background(), tool, json.RawMessage(input))
		}
	}
	t.Fatalf("tool %s is not registered", name)
	return "", nil
}

func TestToolsRegisterOnlyConfiguredApps(t *testing.T) {
	w := newWorld(t)
	w.cfg.Apps[AppOpus] = AppConfig{}
	w.cfg.Apps[AppMycelium] = AppConfig{}
	w.cfg.Apps[AppGlouton] = AppConfig{}
	tools, note, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, Runtime{})
	if err != nil || len(tools) != 4 {
		t.Fatalf("want 4 tools, got %d, %v", len(tools), err)
	}
	for _, name := range []string{"sablier_status", "sablier_projects", "sablier_stop"} {
		if !strings.Contains(note, name) {
			t.Errorf("prompt must list %s: %q", name, note)
		}
	}
	if strings.Contains(note, "opus_list_tasks") || strings.Contains(note, "agenda_list_events") {
		t.Errorf("prompt lists a tool that is not registered: %q", note)
	}
}

func TestEveryMutatingToolNeedsApproval(t *testing.T) {
	w := newWorld(t)
	tools, _, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, Runtime{})
	if err != nil {
		t.Fatal(err)
	}
	gated := map[string]bool{}
	for _, name := range WriteTools() {
		gated[name] = true
	}
	self := map[string]bool{}
	for _, name := range append(SelfWrites(), InboxTools()...) {
		self[name] = true
	}
	for _, tool := range tools {
		mutates := !tool.Spec().ReadOnly
		if mutates && self[tool.Spec().Name] {
			delete(self, tool.Spec().Name)
			continue
		}
		if mutates != gated[tool.Spec().Name] {
			t.Errorf("%s: mutates=%t but approval gated=%t", tool.Spec().Name, mutates, gated[tool.Spec().Name])
		}
		delete(gated, tool.Spec().Name)
	}
	for name := range gated {
		t.Errorf("%s is gated but never registered", name)
	}
	for name := range self {
		t.Errorf("%s is a self write but never registered", name)
	}
}

func TestGloutonInboxToolsSkipApprovalOnly(t *testing.T) {
	gated := strings.Join(WriteTools(), ",")
	for _, name := range InboxTools() {
		if strings.Contains(gated, name) {
			t.Errorf("%s files into Glouton's inbox and must not raise an approval card", name)
		}
	}
	for _, name := range []string{"glouton_create_company", "glouton_update_company", "glouton_create_person", "glouton_update_person"} {
		if !strings.Contains(gated, name) {
			t.Errorf("%s changes contacts directly and must stay approval gated", name)
		}
	}
}

func TestToolsNoneConfigured(t *testing.T) {
	w := newWorld(t)
	for _, name := range appNames {
		w.cfg.Apps[name] = AppConfig{}
	}
	tools, note, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, Runtime{})
	if err != nil || len(tools) != 0 || note != NoMemoryNote {
		t.Fatalf("want no tools and only the no-memory note, got %d tools, %q, %v", len(tools), note, err)
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
		"opus_list_tasks": "task of", "sablier_status": "timer of",
	}
	var wg sync.WaitGroup
	for _, user := range []int64{1, 2} {
		tools, _, err := Tools(gw, user, Defaults{}, Runtime{})
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
	tools, _, _ := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, Runtime{})
	tasks, _ := runTool(t, tools, "opus_list_tasks", `{"project":"p"}`)
	if strings.Contains(tasks, "done") {
		t.Fatalf("hidden fields leaked:\n%s", tasks)
	}
	mine, _ := runTool(t, tools, "opus_list_tasks", `{}`)
	done, _ := runTool(t, tools, "opus_list_tasks", `{"status":"done"}`)
	if strings.Contains(mine, "done thing") || !strings.Contains(done, "done thing") || strings.Contains(done, "task of") {
		t.Fatalf("my-tasks default must hide done work unless a status is asked: %s | %s", mine, done)
	}
	if !strings.Contains(tasks, `"project":"P"`) || !strings.Contains(tasks, `"assignee":"A"`) {
		t.Fatalf("task projection lost fields: %s", tasks)
	}
}

func TestToolReportsReconnect(t *testing.T) {
	w := newWorld(t)
	tools, _, _ := Tools(NewGateway(w.tokens(), w.cfg, nil), 99, Defaults{}, Runtime{})
	_, err := runTool(t, tools, "sablier_status", `{}`)
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
