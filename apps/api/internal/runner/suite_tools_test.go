package runner

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/FacileStudio/nacelle"
)

type mockExecutor struct {
	fn func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func (m mockExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if m.fn != nil {
		return m.fn(ctx, name, args...)
	}
	return nil, nil
}

func findSuiteTool(tools []nacelle.Tool, name string) nacelle.Tool {
	for _, t := range tools {
		if t.Name() == name {
			return t
		}
	}
	return nil
}

func TestCourrierAccountsTool(t *testing.T) {
	var capturedArgs []string
	mock := mockExecutor{fn: func(_ context.Context, _ string, args ...string) ([]byte, error) {
		capturedArgs = args
		return []byte(`[{"id": 2, "email": "test@example.com"}]`), nil
	}}
	tool := findSuiteTool(buildSuiteTools(mock), "courrier_accounts")
	ctx := context.Background()

	out, err := tool.Run(ctx, json.RawMessage(`{}`))
	if err != nil || out != `[{"id": 2, "email": "test@example.com"}]` {
		t.Fatalf("unexpected result: %q, %v", out, err)
	}
	want := []string{"accounts", "--json"}
	if !reflect.DeepEqual(capturedArgs, want) {
		t.Fatalf("got args %v, want %v", capturedArgs, want)
	}
}

func TestCourrierInboxTool(t *testing.T) {
	var capturedArgs []string
	mock := mockExecutor{fn: func(_ context.Context, _ string, args ...string) ([]byte, error) {
		capturedArgs = args
		return []byte(`[{"id": 1}]`), nil
	}}
	tool := findSuiteTool(buildSuiteTools(mock), "courrier_inbox")
	ctx := context.Background()

	out, err := tool.Run(ctx, json.RawMessage(`{}`))
	if err != nil || out != `[{"id": 1}]` {
		t.Fatalf("unexpected result: %q, %v", out, err)
	}
	wantDefault := []string{"inbox", "--limit", "10", "--json"}
	if !reflect.DeepEqual(capturedArgs, wantDefault) {
		t.Fatalf("got args %v, want %v", capturedArgs, wantDefault)
	}

	if _, err := tool.Run(ctx, json.RawMessage(`{"account": 2, "unread": true, "limit": 25}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantCustom := []string{"inbox", "--unread", "--account", "2", "--limit", "25", "--json"}
	if !reflect.DeepEqual(capturedArgs, wantCustom) {
		t.Fatalf("got args %v, want %v", capturedArgs, wantCustom)
	}

	fallbackMock := mockExecutor{fn: func(_ context.Context, _ string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "inbox" {
			return []byte("several mail accounts"), errors.New("exit 1")
		}
		if len(args) > 0 && args[0] == "accounts" {
			return []byte(`[{"id": 2, "is_default": true}]`), nil
		}
		capturedArgs = args
		return []byte(`[{"id": 42}]`), nil
	}}
	fbTool := findSuiteTool(buildSuiteTools(fallbackMock), "courrier_inbox")
	fbOut, err := fbTool.Run(ctx, json.RawMessage(`{}`))
	if err != nil || fbOut != `[{"id": 42}]` {
		t.Fatalf("expected fallback output, got %q, %v", fbOut, err)
	}
	wantFallback := []string{"--account", "2", "inbox", "--limit", "10", "--json"}
	if !reflect.DeepEqual(capturedArgs, wantFallback) {
		t.Fatalf("got fallback args %v, want %v", capturedArgs, wantFallback)
	}

	errMock := mockExecutor{fn: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return []byte("account invalid"), errors.New("exit 1")
	}}
	errTool := findSuiteTool(buildSuiteTools(errMock), "courrier_inbox")
	errOut, err := errTool.Run(ctx, json.RawMessage(`{}`))
	if err != nil || !strings.Contains(errOut, "courrier error: account invalid") {
		t.Fatalf("expected graceful error message, got %q, %v", errOut, err)
	}
}

func TestCourrierSearchTool(t *testing.T) {
	var capturedArgs []string
	mock := mockExecutor{fn: func(_ context.Context, _ string, args ...string) ([]byte, error) {
		capturedArgs = args
		return []byte(`{"results": []}`), nil
	}}
	tool := findSuiteTool(buildSuiteTools(mock), "courrier_search")
	ctx := context.Background()

	out, err := tool.Run(ctx, json.RawMessage(`{"query": "invoice"}`))
	if err != nil || out != `{"results": []}` {
		t.Fatalf("unexpected result: %q, %v", out, err)
	}
	wantDefault := []string{"search", "invoice", "--json"}
	if !reflect.DeepEqual(capturedArgs, wantDefault) {
		t.Fatalf("got args %v, want %v", capturedArgs, wantDefault)
	}

	if _, err := tool.Run(ctx, json.RawMessage(`{"query": "tax", "account": 3, "limit": 5}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantCustom := []string{"search", "tax", "--account", "3", "--limit", "5", "--json"}
	if !reflect.DeepEqual(capturedArgs, wantCustom) {
		t.Fatalf("got args %v, want %v", capturedArgs, wantCustom)
	}

	fallbackMock := mockExecutor{fn: func(_ context.Context, _ string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "search" {
			return []byte("several mail accounts"), errors.New("exit 1")
		}
		if len(args) > 0 && args[0] == "accounts" {
			return []byte(`[{"id": 7, "is_default": true}]`), nil
		}
		capturedArgs = args
		return []byte(`{"results": [1]}`), nil
	}}
	fbTool := findSuiteTool(buildSuiteTools(fallbackMock), "courrier_search")
	fbOut, err := fbTool.Run(ctx, json.RawMessage(`{"query": "hello"}`))
	if err != nil || fbOut != `{"results": [1]}` {
		t.Fatalf("expected fallback output, got %q, %v", fbOut, err)
	}
	wantFallback := []string{"--account", "7", "search", "hello", "--json"}
	if !reflect.DeepEqual(capturedArgs, wantFallback) {
		t.Fatalf("got fallback args %v, want %v", capturedArgs, wantFallback)
	}

	errMock := mockExecutor{fn: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return nil, errors.New("command not found")
	}}
	errTool := findSuiteTool(buildSuiteTools(errMock), "courrier_search")
	errOut, err := errTool.Run(ctx, json.RawMessage(`{"query": "test"}`))
	if err != nil || !strings.Contains(errOut, "courrier error: command not found") {
		t.Fatalf("expected graceful error message, got %q, %v", errOut, err)
	}
}

func TestAgendaCalendarsTool(t *testing.T) {
	var capturedArgs []string
	mock := mockExecutor{fn: func(_ context.Context, _ string, args ...string) ([]byte, error) {
		capturedArgs = args
		return []byte(`[{"id": 1, "name": "Personal"}]`), nil
	}}
	tool := findSuiteTool(buildSuiteTools(mock), "agenda_calendars")
	ctx := context.Background()

	out, err := tool.Run(ctx, json.RawMessage(`{}`))
	if err != nil || out != `[{"id": 1, "name": "Personal"}]` {
		t.Fatalf("unexpected result: %q, %v", out, err)
	}
	want := []string{"calendars", "list", "--json"}
	if !reflect.DeepEqual(capturedArgs, want) {
		t.Fatalf("got args %v, want %v", capturedArgs, want)
	}
}

func TestAgendaEventsTool(t *testing.T) {
	var capturedArgs []string
	mock := mockExecutor{fn: func(_ context.Context, _ string, args ...string) ([]byte, error) {
		capturedArgs = args
		return []byte(`[]`), nil
	}}
	tool := findSuiteTool(buildSuiteTools(mock), "agenda_events")
	ctx := context.Background()

	out, err := tool.Run(ctx, json.RawMessage(`{}`))
	if err != nil || out != `[]` {
		t.Fatalf("unexpected result: %q, %v", out, err)
	}
	wantDefault := []string{"events", "list", "--from", "today", "--json"}
	if !reflect.DeepEqual(capturedArgs, wantDefault) {
		t.Fatalf("got args %v, want %v", capturedArgs, wantDefault)
	}

	if _, err := tool.Run(ctx, json.RawMessage(`{"calendar": 3, "from": "2026-11-01", "to": "2026-11-05"}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantCustom := []string{"events", "list", "--calendar", "3", "--from", "2026-11-01", "--to", "2026-11-05", "--json"}
	if !reflect.DeepEqual(capturedArgs, wantCustom) {
		t.Fatalf("got args %v, want %v", capturedArgs, wantCustom)
	}

	autoMock := mockExecutor{fn: func(_ context.Context, _ string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "calendars" {
			return []byte(`[{"id": 5, "is_personal": true}]`), nil
		}
		capturedArgs = args
		return []byte(`[{"id": 10}]`), nil
	}}
	autoTool := findSuiteTool(buildSuiteTools(autoMock), "agenda_events")
	autoOut, err := autoTool.Run(ctx, json.RawMessage(`{}`))
	if err != nil || autoOut != `[{"id": 10}]` {
		t.Fatalf("expected auto calendar output, got %q, %v", autoOut, err)
	}
	wantAuto := []string{"events", "list", "--calendar", "5", "--from", "today", "--json"}
	if !reflect.DeepEqual(capturedArgs, wantAuto) {
		t.Fatalf("got auto calendar args %v, want %v", capturedArgs, wantAuto)
	}

	errMock := mockExecutor{fn: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return []byte("calendar not found"), errors.New("exit 1")
	}}
	errTool := findSuiteTool(buildSuiteTools(errMock), "agenda_events")
	errOut, err := errTool.Run(ctx, json.RawMessage(`{}`))
	if err != nil || !strings.Contains(errOut, "agenda error: calendar not found") {
		t.Fatalf("expected graceful error, got %q, %v", errOut, err)
	}
}

func TestOpusTasksTool(t *testing.T) {
	var capturedArgs []string
	mock := mockExecutor{fn: func(_ context.Context, _ string, args ...string) ([]byte, error) {
		capturedArgs = args
		return []byte(`[{"id": "1"}]`), nil
	}}
	tool := findSuiteTool(buildSuiteTools(mock), "opus_tasks")
	ctx := context.Background()

	out, err := tool.Run(ctx, json.RawMessage(`{}`))
	if err != nil || out != `[{"id": "1"}]` {
		t.Fatalf("unexpected result: %q, %v", out, err)
	}
	wantDefault := []string{"task", "list", "--json"}
	if !reflect.DeepEqual(capturedArgs, wantDefault) {
		t.Fatalf("got args %v, want %v", capturedArgs, wantDefault)
	}

	if _, err := tool.Run(ctx, json.RawMessage(`{"workspace": "facile", "project": "Bulle", "status": "in-progress", "limit": 20}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantFull := []string{"-w", "facile", "task", "list", "--project", "Bulle", "--status", "in-progress", "--limit", "20", "--json"}
	if !reflect.DeepEqual(capturedArgs, wantFull) {
		t.Fatalf("got args %v, want %v", capturedArgs, wantFull)
	}

	errMock := mockExecutor{fn: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return []byte("unauthorized"), errors.New("exit 2")
	}}
	errTool := findSuiteTool(buildSuiteTools(errMock), "opus_tasks")
	errOut, err := errTool.Run(ctx, json.RawMessage(`{}`))
	if err != nil || !strings.Contains(errOut, "opus error: unauthorized") {
		t.Fatalf("expected graceful error, got %q, %v", errOut, err)
	}
}

func TestSablierStatusTool(t *testing.T) {
	var capturedName string
	var capturedArgs []string
	mock := mockExecutor{fn: func(_ context.Context, name string, args ...string) ([]byte, error) {
		capturedName = name
		capturedArgs = args
		return []byte(`{"running": false}`), nil
	}}
	tool := findSuiteTool(buildSuiteTools(mock), "sablier_status")
	ctx := context.Background()

	out, err := tool.Run(ctx, json.RawMessage(`{}`))
	if err != nil || out != `{"running": false}` {
		t.Fatalf("unexpected result: %q, %v", out, err)
	}
	if capturedName != "sablier" || !reflect.DeepEqual(capturedArgs, []string{"status", "--json"}) {
		t.Fatalf("got %s %v, want sablier [status --json]", capturedName, capturedArgs)
	}

	fallbackMock := mockExecutor{fn: func(_ context.Context, _ string, args ...string) ([]byte, error) {
		if len(args) == 2 && args[1] == "--json" {
			return []byte("unexpected argument '--json'"), errors.New("exit 2")
		}
		capturedArgs = args
		return []byte("No timer running"), nil
	}}
	fbTool := findSuiteTool(buildSuiteTools(fallbackMock), "sablier_status")
	fbOut, err := fbTool.Run(ctx, json.RawMessage(`{}`))
	if err != nil || fbOut != "No timer running" {
		t.Fatalf("unexpected fallback result: %q, %v", fbOut, err)
	}
	if !reflect.DeepEqual(capturedArgs, []string{"status"}) {
		t.Fatalf("got fallback args %v, want [status]", capturedArgs)
	}

	errMock := mockExecutor{fn: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
		return nil, errors.New("daemon offline")
	}}
	errTool := findSuiteTool(buildSuiteTools(errMock), "sablier_status")
	errOut, err := errTool.Run(ctx, json.RawMessage(`{}`))
	if err != nil || !strings.Contains(errOut, "sablier error: daemon offline") {
		t.Fatalf("expected graceful error, got %q, %v", errOut, err)
	}
}

func TestBuildSuiteTools(t *testing.T) {
	tools := buildSuiteTools(mockExecutor{})
	if len(tools) != 7 {
		t.Fatalf("expected 7 tools, got %d", len(tools))
	}
	expected := []string{"courrier_inbox", "courrier_search", "courrier_accounts", "agenda_events", "agenda_calendars", "opus_tasks", "sablier_status"}
	for _, name := range expected {
		tTool := findSuiteTool(tools, name)
		if tTool == nil {
			t.Fatalf("missing tool: %s", name)
		}
		ro, ok := tTool.(nacelle.ReadOnlyTool)
		if !ok || !ro.IsReadOnly() {
			t.Fatalf("tool %s should be read only", name)
		}
	}
}
