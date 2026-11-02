package assistant

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
)

// ExecSuite reads the suite through the courrier, agenda and opus CLIs.
type ExecSuite struct {
	Run Runner
}

// MailAccounts runs `courrier accounts`.
func (e ExecSuite) MailAccounts(ctx context.Context) ([]MailAccount, error) {
	return execJSON[[]MailAccount](ctx, e.Run, "courrier", "accounts", "--json")
}

// Triage runs `courrier triage` for one page.
func (e ExecSuite) Triage(ctx context.Context, q TriageQuery) (Triage, error) {
	args := []string{"courrier", "triage", "--all", "--unread", "--limit", strconv.Itoa(q.Limit), "--page", strconv.Itoa(q.Page)}
	if q.Since != "" {
		args = append(args, "--since", q.Since)
	}
	var doc Triage
	err := runJSON(ctx, e.Run, 0, &doc, append(args, "--json")...)
	var ce *CmdError
	if errors.As(err, &ce) && strings.Contains(ce.Stderr, triageUnsupported) {
		return doc, errors.Join(ErrTriageUnsupported, err)
	}
	return doc, err
}

// UnreadInbox runs `courrier inbox --unread` for one account.
func (e ExecSuite) UnreadInbox(ctx context.Context, account int64) ([]TriageItem, error) {
	var doc struct {
		Emails []TriageItem `json:"emails"`
	}
	id := strconv.FormatInt(account, 10)
	err := runJSON(ctx, e.Run, 0, &doc, "courrier", "inbox", "--account", id, "--unread", "--json")
	return doc.Emails, err
}

// SentMail runs `courrier list sent` for one account.
func (e ExecSuite) SentMail(ctx context.Context, account int64, since string, limit int) ([]SentMail, error) {
	var doc struct {
		Emails []SentMail `json:"emails"`
	}
	id := strconv.FormatInt(account, 10)
	args := []string{"courrier", "list", "sent", "--account", id, "--since", since, "--limit", strconv.Itoa(limit), "--json"}
	err := runJSON(ctx, e.Run, 0, &doc, args...)
	return doc.Emails, err
}

// Calendars runs `agenda calendars list`.
func (e ExecSuite) Calendars(ctx context.Context) ([]Calendar, error) {
	return execJSON[[]Calendar](ctx, e.Run, "agenda", "calendars", "list", "--json")
}

// Events runs `agenda events list` for one calendar and a day range.
func (e ExecSuite) Events(ctx context.Context, calendar int, from, to time.Time) ([]CalendarEvent, error) {
	args := []string{
		"agenda", "events", "list", "--calendar", strconv.Itoa(calendar),
		"--from", from.Format("2006-01-02"), "--to", to.Format("2006-01-02"), "--json",
	}
	return execJSON[[]CalendarEvent](ctx, e.Run, args...)
}

// Workspaces runs `opus workspace list`.
func (e ExecSuite) Workspaces(ctx context.Context) ([]Workspace, error) {
	return execJSON[[]Workspace](ctx, e.Run, "opus", "workspace", "list", "--json")
}

// Tasks runs `opus task list` in one workspace.
func (e ExecSuite) Tasks(ctx context.Context, workspace string) ([]Task, error) {
	return execJSON[[]Task](ctx, e.Run, "opus", "-w", workspace, "task", "list", "--json")
}
