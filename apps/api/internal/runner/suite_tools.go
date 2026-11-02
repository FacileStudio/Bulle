package runner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/FacileStudio/nacelle"
)

// CommandExecutor executes system commands.
type CommandExecutor interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type defaultCommandExecutor struct{}

func (defaultCommandExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return exitErr.Stderr, err
		}
		return out, err
	}
	return out, nil
}

type courrierInboxInput struct {
	Account int  `json:"account,omitempty"`
	Unread  bool `json:"unread,omitempty"`
	Limit   int  `json:"limit,omitempty"`
}

type courrierSearchInput struct {
	Query string `json:"query" jsonschema:"required"`
	Limit int    `json:"limit,omitempty"`
}

type agendaEventsInput struct {
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Calendar int    `json:"calendar,omitempty"`
}

type opusTasksInput struct {
	Workspace string `json:"workspace,omitempty"`
	Status    string `json:"status,omitempty"`
}

type sablierStatusInput struct{}

func formatOutput(out []byte, err error, name string) string {
	if err == nil {
		return string(out)
	}
	msg := strings.TrimSpace(string(out))
	if msg != "" {
		return fmt.Sprintf("%s error: %s", name, msg)
	}
	return fmt.Sprintf("%s error: %v", name, err)
}

func executeCommand(ctx context.Context, exec CommandExecutor, name string, args ...string) (string, error) {
	out, err := exec.Run(ctx, name, args...)
	return formatOutput(out, err, name), nil
}

func runCourrierInbox(ctx context.Context, exec CommandExecutor, in courrierInboxInput) (string, error) {
	args := []string{"inbox"}
	if in.Unread {
		args = append(args, "--unread")
	}
	if in.Account > 0 {
		args = append(args, "--account", strconv.Itoa(in.Account))
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}
	args = append(args, "--limit", strconv.Itoa(limit), "--json")
	return executeCommand(ctx, exec, "courrier", args...)
}

func runCourrierSearch(ctx context.Context, exec CommandExecutor, in courrierSearchInput) (string, error) {
	args := []string{"search", in.Query}
	if in.Limit > 0 {
		args = append(args, "--limit", strconv.Itoa(in.Limit))
	}
	args = append(args, "--json")
	return executeCommand(ctx, exec, "courrier", args...)
}

func runAgendaEvents(ctx context.Context, exec CommandExecutor, in agendaEventsInput) (string, error) {
	args := []string{"events", "list"}
	if in.Calendar > 0 {
		args = append(args, "--calendar", strconv.Itoa(in.Calendar))
	}
	from := in.From
	if from == "" {
		from = "today"
	}
	args = append(args, "--from", from)
	if in.To != "" {
		args = append(args, "--to", in.To)
	}
	args = append(args, "--json")
	return executeCommand(ctx, exec, "agenda", args...)
}

func runOpusTasks(ctx context.Context, exec CommandExecutor, in opusTasksInput) (string, error) {
	var args []string
	if in.Workspace != "" {
		args = append(args, "-w", in.Workspace)
	}
	args = append(args, "task", "list")
	if in.Status != "" {
		args = append(args, "--status", in.Status)
	}
	args = append(args, "--json")
	return executeCommand(ctx, exec, "opus", args...)
}

func buildSuiteTools(executor CommandExecutor) []nacelle.Tool {
	opts := nacelle.ToolOptions{ReadOnly: true}
	t1, e1 := nacelle.NewToolWithOptions("courrier_inbox", "List inbox conversations from Courrier.", func(ctx context.Context, in courrierInboxInput) (string, error) {
		return runCourrierInbox(ctx, executor, in)
	}, opts)
	t2, e2 := nacelle.NewToolWithOptions("courrier_search", "Search emails in Courrier.", func(ctx context.Context, in courrierSearchInput) (string, error) {
		return runCourrierSearch(ctx, executor, in)
	}, opts)
	t3, e3 := nacelle.NewToolWithOptions("agenda_events", "List calendar events from Agenda.", func(ctx context.Context, in agendaEventsInput) (string, error) {
		return runAgendaEvents(ctx, executor, in)
	}, opts)
	t4, e4 := nacelle.NewToolWithOptions("opus_tasks", "List tasks from Opus project management.", func(ctx context.Context, in opusTasksInput) (string, error) {
		return runOpusTasks(ctx, executor, in)
	}, opts)
	t5, e5 := nacelle.NewToolWithOptions("sablier_status", "Show currently running timer status in Sablier.", func(ctx context.Context, _ sablierStatusInput) (string, error) {
		return executeCommand(ctx, executor, "sablier", "status", "--json")
	}, opts)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil {
		return nil
	}
	return []nacelle.Tool{t1, t2, t3, t4, t5}
}
