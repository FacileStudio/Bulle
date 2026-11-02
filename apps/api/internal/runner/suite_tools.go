package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	return cmd.CombinedOutput()
}

type courrierInboxInput struct {
	Account int  `json:"account,omitempty"`
	Unread  bool `json:"unread,omitempty"`
	Limit   int  `json:"limit,omitempty"`
}

type courrierSearchInput struct {
	Query   string `json:"query" jsonschema:"required"`
	Account int    `json:"account,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

type courrierAccountsInput struct{}

type agendaEventsInput struct {
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Calendar int    `json:"calendar,omitempty"`
}

type agendaCalendarsInput struct{}

type opusTasksInput struct {
	Workspace string `json:"workspace,omitempty"`
	Project   string `json:"project,omitempty"`
	Status    string `json:"status,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type sablierStatusInput struct{}

type courrierAccountItem struct {
	ID        int  `json:"id"`
	IsDefault bool `json:"is_default"`
}

type agendaCalendarItem struct {
	ID         int  `json:"id"`
	IsPersonal bool `json:"is_personal"`
}

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

func getDefaultCourrierAccount(ctx context.Context, exec CommandExecutor) int {
	out, err := exec.Run(ctx, "courrier", "accounts", "--json")
	if err != nil {
		return 0
	}
	var accounts []courrierAccountItem
	if err := json.Unmarshal(out, &accounts); err != nil || len(accounts) == 0 {
		return 0
	}
	for _, a := range accounts {
		if a.IsDefault {
			return a.ID
		}
	}
	return accounts[0].ID
}

func getDefaultAgendaCalendar(ctx context.Context, exec CommandExecutor) int {
	out, err := exec.Run(ctx, "agenda", "calendars", "list", "--json")
	if err != nil {
		return 0
	}
	var calendars []agendaCalendarItem
	if err := json.Unmarshal(out, &calendars); err != nil || len(calendars) == 0 {
		return 0
	}
	for _, c := range calendars {
		if c.IsPersonal {
			return c.ID
		}
	}
	return calendars[0].ID
}

func runCourrierAccounts(ctx context.Context, exec CommandExecutor) (string, error) {
	return executeCommand(ctx, exec, "courrier", "accounts", "--json")
}

func runCourrierInbox(ctx context.Context, exec CommandExecutor, in courrierInboxInput) (string, error) {
	account := in.Account
	args := []string{"inbox"}
	if in.Unread {
		args = append(args, "--unread")
	}
	if account > 0 {
		args = append(args, "--account", strconv.Itoa(account))
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}
	args = append(args, "--limit", strconv.Itoa(limit), "--json")
	out, err := exec.Run(ctx, "courrier", args...)
	if err != nil && account <= 0 && strings.Contains(string(out), "several mail accounts") {
		if defID := getDefaultCourrierAccount(ctx, exec); defID > 0 {
			fallbackArgs := []string{"--account", strconv.Itoa(defID), "inbox"}
			if in.Unread {
				fallbackArgs = append(fallbackArgs, "--unread")
			}
			fallbackArgs = append(fallbackArgs, "--limit", strconv.Itoa(limit), "--json")
			return executeCommand(ctx, exec, "courrier", fallbackArgs...)
		}
	}
	return formatOutput(out, err, "courrier"), nil
}

func runCourrierSearch(ctx context.Context, exec CommandExecutor, in courrierSearchInput) (string, error) {
	account := in.Account
	args := []string{"search", in.Query}
	if account > 0 {
		args = append(args, "--account", strconv.Itoa(account))
	}
	if in.Limit > 0 {
		args = append(args, "--limit", strconv.Itoa(in.Limit))
	}
	args = append(args, "--json")
	out, err := exec.Run(ctx, "courrier", args...)
	if err != nil && account <= 0 && strings.Contains(string(out), "several mail accounts") {
		if defID := getDefaultCourrierAccount(ctx, exec); defID > 0 {
			fallbackArgs := []string{"--account", strconv.Itoa(defID), "search", in.Query}
			if in.Limit > 0 {
				fallbackArgs = append(fallbackArgs, "--limit", strconv.Itoa(in.Limit))
			}
			fallbackArgs = append(fallbackArgs, "--json")
			return executeCommand(ctx, exec, "courrier", fallbackArgs...)
		}
	}
	return formatOutput(out, err, "courrier"), nil
}

func runAgendaCalendars(ctx context.Context, exec CommandExecutor) (string, error) {
	return executeCommand(ctx, exec, "agenda", "calendars", "list", "--json")
}

func runAgendaEvents(ctx context.Context, exec CommandExecutor, in agendaEventsInput) (string, error) {
	calID := in.Calendar
	if calID <= 0 {
		calID = getDefaultAgendaCalendar(ctx, exec)
	}
	args := []string{"events", "list"}
	if calID > 0 {
		args = append(args, "--calendar", strconv.Itoa(calID))
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
	if in.Project != "" {
		args = append(args, "--project", in.Project)
	}
	if in.Status != "" {
		args = append(args, "--status", in.Status)
	}
	if in.Limit > 0 {
		args = append(args, "--limit", strconv.Itoa(in.Limit))
	}
	args = append(args, "--json")
	return executeCommand(ctx, exec, "opus", args...)
}

func ensureSablierConfig() {
	token := os.Getenv("SABLIER_TOKEN")
	if token == "" {
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	cfgPath := filepath.Join(home, ".sablier.yml")
	serverURL := os.Getenv("SABLIER_SERVER_URL")
	if serverURL == "" {
		serverURL = "https://sablier.facile.studio/api"
	}
	content := fmt.Sprintf("server_url: %s\ntoken: '%s'\n", serverURL, token)
	_ = os.WriteFile(cfgPath, []byte(content), 0600)
}

func runSablierStatus(ctx context.Context, exec CommandExecutor) (string, error) {
	ensureSablierConfig()
	out, err := exec.Run(ctx, "sablier", "status", "--json")
	if err != nil && strings.Contains(string(out), "unexpected argument '--json'") {
		return executeCommand(ctx, exec, "sablier", "status")
	}
	return formatOutput(out, err, "sablier"), nil
}

func buildSuiteTools(executor CommandExecutor) []nacelle.Tool {
	opts := nacelle.ToolOptions{ReadOnly: true}
	t1, e1 := nacelle.NewToolWithOptions("courrier_inbox", "List inbox conversations from Courrier.", func(ctx context.Context, in courrierInboxInput) (string, error) {
		return runCourrierInbox(ctx, executor, in)
	}, opts)
	t2, e2 := nacelle.NewToolWithOptions("courrier_search", "Search emails in Courrier.", func(ctx context.Context, in courrierSearchInput) (string, error) {
		return runCourrierSearch(ctx, executor, in)
	}, opts)
	t3, e3 := nacelle.NewToolWithOptions("courrier_accounts", "List configured mail accounts in Courrier.", func(ctx context.Context, _ courrierAccountsInput) (string, error) {
		return runCourrierAccounts(ctx, executor)
	}, opts)
	t4, e4 := nacelle.NewToolWithOptions("agenda_events", "List calendar events from Agenda.", func(ctx context.Context, in agendaEventsInput) (string, error) {
		return runAgendaEvents(ctx, executor, in)
	}, opts)
	t5, e5 := nacelle.NewToolWithOptions("agenda_calendars", "List all calendars from Agenda.", func(ctx context.Context, _ agendaCalendarsInput) (string, error) {
		return runAgendaCalendars(ctx, executor)
	}, opts)
	t6, e6 := nacelle.NewToolWithOptions("opus_tasks", "List tasks from Opus project management.", func(ctx context.Context, in opusTasksInput) (string, error) {
		return runOpusTasks(ctx, executor, in)
	}, opts)
	t7, e7 := nacelle.NewToolWithOptions("sablier_status", "Show currently running timer status in Sablier.", func(ctx context.Context, _ sablierStatusInput) (string, error) {
		return runSablierStatus(ctx, executor)
	}, opts)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil || e6 != nil || e7 != nil {
		return nil
	}
	return []nacelle.Tool{t1, t2, t3, t4, t5, t6, t7}
}
