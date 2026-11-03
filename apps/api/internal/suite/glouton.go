package suite

import (
	"context"
	"errors"
	"net/url"
	"strconv"
)

const (
	gloutonContactsPath    = "/api/agent/contacts"
	gloutonProbePath       = gloutonContactsPath + "?limit=1"
	gloutonCompaniesPath   = "/api/agent/companies"
	gloutonPeoplePath      = "/api/agent/people"
	gloutonOutreachPath    = "/api/agent/outreach"
	gloutonDraftsPath      = "/api/agent/drafts"
	gloutonSuggestionsPath = "/api/agent/suggestions"
	gloutonInboxPath       = "/app/inbox"
	gloutonMaxLimit        = 100
)

// GloutonContactQuery filters the unified company and person list.
type GloutonContactQuery struct {
	Query       string
	Stage       string
	Temperature string
	Limit       int
}

// Glouton calls Glouton's agent API as the user.
type Glouton struct{ gw *Gateway }

// NewGlouton returns a Glouton client that calls through gw.
func NewGlouton(gw *Gateway) Glouton { return Glouton{gw: gw} }

// InboxURL is where the human reviews and sends the agent's drafts.
func (g Glouton) InboxURL() string {
	return g.gw.Config().Apps[AppGlouton].URL + gloutonInboxPath
}

func (g Glouton) Contacts(ctx context.Context, userID int64, q GloutonContactQuery) ([]byte, error) {
	query := url.Values{"limit": {strconv.Itoa(gloutonLimit(q.Limit))}}
	for key, value := range map[string]string{"q": q.Query, "stage": q.Stage, "temperature": q.Temperature} {
		if value != "" {
			query.Set(key, value)
		}
	}
	return g.read(ctx, userID, gloutonContactsPath, query)
}

func (g Glouton) Company(ctx context.Context, userID int64, id string) ([]byte, error) {
	path, err := gloutonItem(gloutonCompaniesPath, id)
	if err != nil {
		return nil, err
	}
	return g.read(ctx, userID, path, nil)
}

func (g Glouton) Person(ctx context.Context, userID int64, id string) ([]byte, error) {
	path, err := gloutonItem(gloutonPeoplePath, id)
	if err != nil {
		return nil, err
	}
	return g.read(ctx, userID, path, nil)
}

func (g Glouton) Outreach(ctx context.Context, userID int64, companyID, personID string) ([]byte, error) {
	if companyID == "" && personID == "" {
		return nil, errors.New("pass a company id or a person id")
	}
	query := url.Values{}
	if companyID != "" {
		query.Set("companyId", companyID)
	}
	if personID != "" {
		query.Set("personId", personID)
	}
	return g.read(ctx, userID, gloutonOutreachPath, query)
}

func (g Glouton) Suggestions(ctx context.Context, userID int64, status string) ([]byte, error) {
	if status == "" {
		status = "OPEN"
	}
	return g.read(ctx, userID, gloutonSuggestionsPath, url.Values{"status": {status}})
}
