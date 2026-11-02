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

func (c Courrier) Inbox(ctx context.Context, userID, account int64, unread bool, limit int) ([]byte, error) {
	id, err := c.account(ctx, userID, account)
	if err != nil {
		return nil, err
	}
	query := url.Values{"limit": {strconv.Itoa(clampLimit(limit, 10))}}
	if unread {
		query.Set("unread", "true")
	}
	return c.gw.Get(ctx, userID, AppCourrier, fmt.Sprintf(courrierInboxPath, id), query)
}

func (c Courrier) Search(ctx context.Context, userID, account int64, text string, limit int) ([]byte, error) {
	id, err := c.account(ctx, userID, account)
	if err != nil {
		return nil, err
	}
	query := url.Values{"q": {text}, "limit": {strconv.Itoa(clampLimit(limit, 10))}}
	return c.gw.Get(ctx, userID, AppCourrier, fmt.Sprintf(courrierSearchPath, id), query)
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
