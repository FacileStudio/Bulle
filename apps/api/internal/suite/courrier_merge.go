package suite

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type mailAccountRef struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type mergedMail struct {
	Accounts []mailAccountRef `json:"accounts"`
	Emails   []map[string]any `json:"emails"`
}

// mergeMailPages joins one email page per account into a single list, newest
// first, each email tagged with the address of the mailbox it sits in.
func mergeMailPages(accounts []CourrierAccount, pages [][]byte, limit int) ([]byte, error) {
	out := mergedMail{Emails: []map[string]any{}}
	for i, acc := range accounts {
		out.Accounts = append(out.Accounts, mailAccountRef{ID: acc.ID, Email: acc.Email})
		var page struct {
			Emails []map[string]any `json:"emails"`
		}
		if err := decodeInto(pages[i], &page); err != nil {
			return nil, fmt.Errorf("courrier: reading account %d: %w", acc.ID, err)
		}
		for _, email := range page.Emails {
			email["account_id"] = acc.ID
			email["account_email"] = acc.Email
			out.Emails = append(out.Emails, email)
		}
	}
	sort.SliceStable(out.Emails, func(a, b int) bool {
		return mailDate(out.Emails[a]).After(mailDate(out.Emails[b]))
	})
	if len(out.Emails) > limit {
		out.Emails = out.Emails[:limit]
	}
	return json.Marshal(out)
}

func mailDate(email map[string]any) time.Time {
	text, _ := email["date"].(string)
	when, _ := time.Parse(time.RFC3339, text)
	return when
}
