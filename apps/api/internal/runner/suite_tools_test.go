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

	if _, err := tool.Run(ctx, json.RawMessage(`{"query": "tax", "limit": 5}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantLimit := []string{"search", "tax", "--limit", "5", "--json"}
	if !reflect.DeepEqual(capturedArgs, wantLimit) {
		t.Fatalf("got args %v, want %v", capturedArgs, wantLimit)
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

	if _, err := tool.Run(ctx, json.RawMessage(`{"workspace": "facile", "status": "in-progress"}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantWorkspace := []string{"-w", "facile", "task", "list", "--status", "in-progress", "--json"}
	if !reflect.DeepEqual(capturedArgs, wantWorkspace) {
		t.Fatalf("got args %v, want %v", capturedArgs, wantWorkspace)
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
	if len(tools) != 5 {
		t.Fatalf("expected 5 tools, got %d", len(tools))
	}
	expected := []string{"courrier_inbox", "courrier_search", "agenda_events", "opus_tasks", "sablier_status"}
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
