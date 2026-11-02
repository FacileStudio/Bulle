package assistant

import (
	"context"
	"errors"
	"time"
)

// ErrTriageUnsupported means the mail server has no triage endpoint, so the
// mail source falls back to per-account inboxes and local sender rules.
var ErrTriageUnsupported = errors.New("assistant: courrier triage unsupported")

// Suite is the read side of the Facile apps the assistant watches, for the
// user it runs as.
type Suite interface {
	MailReader
	CalendarReader
	TaskReader
}

// MailReader reads courrier.
type MailReader interface {
	MailAccounts(ctx context.Context) ([]MailAccount, error)
	Triage(ctx context.Context, q TriageQuery) (Triage, error)
	UnreadInbox(ctx context.Context, account int64) ([]TriageItem, error)
	SentMail(ctx context.Context, account int64, since string, limit int) ([]SentMail, error)
}

// CalendarReader reads agenda.
type CalendarReader interface {
	Calendars(ctx context.Context) ([]Calendar, error)
	Events(ctx context.Context, calendar int, from, to time.Time) ([]CalendarEvent, error)
}

// TaskReader reads opus.
type TaskReader interface {
	Workspaces(ctx context.Context) ([]Workspace, error)
	Tasks(ctx context.Context, workspace string) ([]Task, error)
}

// MailAccount is one courrier mailbox.
type MailAccount struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

// TriageQuery pages through unread triaged mail. Since is a duration such as
// "90m"; empty means no lower bound.
type TriageQuery struct {
	Page  int
	Limit int
	Since string
}

// SentMail is one mail the user sent.
type SentMail struct {
	Subject string      `json:"subject"`
	Date    string      `json:"date"`
	To      []Recipient `json:"to_addresses"`
}

// Recipient is one address of a sent mail.
type Recipient struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Calendar is one agenda calendar.
type Calendar struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CalendarEvent is one agenda event.
type CalendarEvent struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	StartAt time.Time `json:"start_at"`
	Status  string    `json:"status"`
}

// Workspace is one opus workspace.
type Workspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Task is one opus task.
type Task struct {
	ID      string     `json:"id"`
	Title   string     `json:"title"`
	Status  string     `json:"status"`
	Done    bool       `json:"done"`
	DueDate *time.Time `json:"dueDate"`
}
