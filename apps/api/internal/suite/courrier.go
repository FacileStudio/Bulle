package suite

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

const (
	courrierAccountsPath = "/api/accounts/"
	courrierInboxPath    = "/api/accounts/%d/mail/folders/inbox/emails"
	courrierSearchPath   = "/api/accounts/%d/mail/search"
)

type CourrierAccount struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	IsDefault bool   `json:"is_default"`
}

type Courrier struct{ gw *Gateway }

func NewCourrier(gw *Gateway) Courrier { return Courrier{gw: gw} }

func (c Courrier) Accounts(ctx context.Context, userID int64) ([]CourrierAccount, error) {
	raw, err := c.gw.Get(ctx, userID, AppCourrier, courrierAccountsPath, nil)
	if err != nil {
		return nil, err
	}
	var page struct {
		Accounts []CourrierAccount `json:"accounts"`
	}
	if err := decodeInto(raw, &page); err != nil {
		return nil, fmt.Errorf("courrier: reading accounts: %w", err)
	}
	return page.Accounts, nil
}

// Inbox lists one account's inbox, or every account's merged by date when
// account is 0.
func (c Courrier) Inbox(ctx context.Context, userID, account int64, unread bool, limit int) ([]byte, error) {
	query := url.Values{"limit": {strconv.Itoa(clampLimit(limit, 10))}}
	if unread {
		query.Set("unread", "true")
	}
	return c.acrossAccounts(ctx, userID, account, mailQuery{path: courrierInboxPath, values: query, limit: clampLimit(limit, 10)})
}

// Search finds emails in one account, or in every account when account is 0,
// so a thread in a secondary mailbox is never missed.
func (c Courrier) Search(ctx context.Context, userID, account int64, text string, limit int) ([]byte, error) {
	query := url.Values{"q": {text}, "limit": {strconv.Itoa(clampLimit(limit, 10))}}
	return c.acrossAccounts(ctx, userID, account, mailQuery{path: courrierSearchPath, values: query, limit: clampLimit(limit, 10)})
}

type mailQuery struct {
	path   string
	values url.Values
	limit  int
}

func (c Courrier) acrossAccounts(ctx context.Context, userID, account int64, q mailQuery) ([]byte, error) {
	if account > 0 {
		return c.gw.Get(ctx, userID, AppCourrier, fmt.Sprintf(q.path, account), q.values)
	}
	accounts, err := c.Accounts(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return nil, errors.New("the user has no mail account configured in Courrier")
	}
	pages := make([][]byte, 0, len(accounts))
	for _, acc := range accounts {
		raw, err := c.gw.Get(ctx, userID, AppCourrier, fmt.Sprintf(q.path, acc.ID), q.values)
		if err != nil {
			return nil, err
		}
		pages = append(pages, raw)
	}
	return mergeMailPages(accounts, pages, q.limit)
}

func (c Courrier) account(ctx context.Context, userID, requested int64) (int64, error) {
	if requested > 0 {
		return requested, nil
	}
	accounts, err := c.Accounts(ctx, userID)
	if err != nil {
		return 0, err
	}
	for _, account := range accounts {
		if account.IsDefault {
			return account.ID, nil
		}
	}
	if len(accounts) == 0 {
		return 0, errors.New("the user has no mail account configured in Courrier")
	}
	return accounts[0].ID, nil
}

func clampLimit(limit, fallback int) int {
	if limit <= 0 {
		return fallback
	}
	return min(limit, 50)
}
