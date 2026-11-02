package assistant

import (
	"context"
)

func (m *MailSource) fetchFallback(ctx context.Context) ([]Item, error) {
	accounts, err := m.Suite.MailAccounts(ctx)
	if err != nil {
		return nil, err
	}
	var rows []TriageItem
	for _, a := range accounts {
		if !m.wantsAccount(a.ID) {
			continue
		}
		got, err := m.inbox(ctx, a)
		if err != nil {
			return nil, err
		}
		rows = append(rows, got...)
	}
	return m.items(rows), nil
}

func (m *MailSource) inbox(ctx context.Context, a MailAccount) ([]TriageItem, error) {
	emails, err := m.Suite.UnreadInbox(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	for i := range emails {
		e := &emails[i]
		class, reason := ClassifySender(e.FromAddress, m.Settings.DropSenders)
		e.Class, e.Reason = string(class), reason
		e.Account = a.Email
	}
	return emails, nil
}
