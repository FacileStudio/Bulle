package assistant

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	sentWindow = "3d"
	maxSent    = 8
)

// SentItem is one mail the user sent recently, kept as context for a digest:
// it says what they are waiting on, never what to notify.
type SentItem struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	At      string `json:"at"`
}

// RecentSent lists the newest mails the user sent in the last three days,
// across the picked accounts. A failure is returned with whatever was read so
// far; callers treat it as missing context, not as an error to stop on.
func RecentSent(ctx context.Context, d Deps) ([]SentItem, error) {
	src := &MailSource{Suite: d.Suite, Settings: d.Settings}
	accounts, err := d.Suite.MailAccounts(ctx)
	if err != nil {
		return nil, err
	}
	var out []SentItem
	for _, a := range accounts {
		if !src.wantsAccount(a.ID) {
			continue
		}
		got, err := sentOf(ctx, d, a.ID)
		if err != nil {
			return out, fmt.Errorf("sent mail of account %d: %w", a.ID, err)
		}
		out = append(out, got...)
	}
	if len(out) > maxSent {
		out = out[:maxSent]
	}
	return out, nil
}

func sentOf(ctx context.Context, d Deps, account int64) ([]SentItem, error) {
	emails, err := d.Suite.SentMail(ctx, account, sentWindow, maxSent)
	if err != nil {
		return nil, err
	}
	var items []SentItem
	for _, e := range emails {
		var to []string
		for _, r := range e.To {
			to = append(to, firstNonEmpty(r.Name, r.Email))
		}
		at := ""
		if t, err := time.Parse(time.RFC3339, e.Date); err == nil {
			at = t.Format("2006-01-02")
		}
		items = append(items, SentItem{To: strings.Join(to, ", "), Subject: e.Subject, At: at})
	}
	return items, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
