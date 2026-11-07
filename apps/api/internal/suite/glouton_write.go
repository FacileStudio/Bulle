package suite

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// GloutonPersonFields are the person fields the agent may set on a create or
// an update. A nil field is left untouched.
type GloutonPersonFields struct {
	CompanyID *string  `json:"company_id,omitempty" jsonschema:"description=Company id this person works for, from glouton_search_contacts"`
	FirstName *string  `json:"first_name,omitempty"`
	LastName  *string  `json:"last_name,omitempty"`
	Email     *string  `json:"email,omitempty"`
	Role      *string  `json:"role,omitempty" jsonschema:"description=Job title or role at the company"`
	Phones    []string `json:"phones,omitempty"`
	Stage     *string  `json:"stage,omitempty" jsonschema:"enum=PROSPECT,enum=CONTACTED,enum=CLIENT,enum=PARTNER,enum=PERSONAL"`
	Notes     *string  `json:"notes,omitempty"`
}

// GloutonCompanyFields are the company fields the agent may set on a create
// or an update. A nil field is left untouched.
type GloutonCompanyFields struct {
	Name    *string `json:"name,omitempty"`
	Domain  *string `json:"domain,omitempty" jsonschema:"description=Bare domain such as example.com"`
	Website *string `json:"website,omitempty"`
	City    *string `json:"city,omitempty"`
	Country *string `json:"country,omitempty"`
	Stage   *string `json:"stage,omitempty" jsonschema:"enum=PROSPECT,enum=CONTACTED,enum=CLIENT,enum=PARTNER,enum=PERSONAL"`
	Notes   *string `json:"notes,omitempty"`
}

// GloutonCompanyPatch is a company update: the company fields plus the
// temperature and score Glouton ranks prospects by.
type GloutonCompanyPatch struct {
	GloutonCompanyFields
	Temperature *string `json:"temperature,omitempty" jsonschema:"enum=HOT,enum=WARM,enum=COLD"`
	Score       *int    `json:"score,omitempty" jsonschema:"description=Lead score from 0 to 100"`
}

// GloutonNewPerson is a person created together with a new company.
type GloutonNewPerson struct {
	FirstName string   `json:"first_name,omitempty"`
	LastName  string   `json:"last_name,omitempty"`
	Email     string   `json:"email,omitempty"`
	Role      string   `json:"role,omitempty"`
	Phones    []string `json:"phones,omitempty"`
}

// GloutonDraft is the content of an outreach email the human sends later
// from Glouton's agent inbox.
type GloutonDraft struct {
	Subject   *string `json:"subject,omitempty"`
	Text      *string `json:"text,omitempty" jsonschema:"description=Plain-text body"`
	HTML      *string `json:"html,omitempty" jsonschema:"description=Optional HTML body"`
	Rationale *string `json:"rationale,omitempty" jsonschema:"description=Why this email, for the human reviewing it"`
}

// GloutonSuggestion is a proposal the human accepts or dismisses in Glouton.
type GloutonSuggestion struct {
	Kind      string         `json:"kind"`
	Title     string         `json:"title"`
	Body      string         `json:"body,omitempty"`
	Rationale string         `json:"rationale,omitempty"`
	CompanyID string         `json:"companyId,omitempty"`
	PersonID  string         `json:"personId,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
}

func (g Glouton) CreateCompany(ctx context.Context, userID int64, fields GloutonCompanyFields, people []GloutonNewPerson) ([]byte, error) {
	if fields.Name == nil || strings.TrimSpace(*fields.Name) == "" {
		return nil, errors.New("a company needs a name")
	}
	body, err := gloutonBody(struct {
		GloutonCompanyFields
		People []GloutonNewPerson `json:"people,omitempty"`
	}{fields, people})
	if err != nil {
		return nil, err
	}
	return g.write(ctx, userID, http.MethodPost, gloutonCompaniesPath, body)
}

func (g Glouton) UpdateCompany(ctx context.Context, userID int64, id string, patch GloutonCompanyPatch) ([]byte, error) {
	path, err := gloutonItem(gloutonCompaniesPath, id)
	if err != nil {
		return nil, err
	}
	body, err := toMap(patch)
	if err != nil {
		return nil, err
	}
	return g.patch(ctx, userID, path, body)
}

func (g Glouton) CreatePerson(ctx context.Context, userID int64, fields GloutonPersonFields) ([]byte, error) {
	if fields.FirstName == nil && fields.LastName == nil && fields.Email == nil {
		return nil, errors.New("a person needs a name or an email")
	}
	body, err := gloutonBody(fields)
	if err != nil {
		return nil, err
	}
	return g.write(ctx, userID, http.MethodPost, gloutonPeoplePath, body)
}

func (g Glouton) UpdatePerson(ctx context.Context, userID int64, id string, fields GloutonPersonFields) ([]byte, error) {
	path, err := gloutonItem(gloutonPeoplePath, id)
	if err != nil {
		return nil, err
	}
	body, err := gloutonBody(fields)
	if err != nil {
		return nil, err
	}
	return g.patch(ctx, userID, path, body)
}

func (g Glouton) SaveDraft(ctx context.Context, userID int64, personID string, draft GloutonDraft) ([]byte, error) {
	if personID == "" || draft.Subject == nil || draft.Text == nil {
		return nil, errors.New("a draft needs a person id, a subject and a text body")
	}
	body, err := toMap(draft)
	if err != nil {
		return nil, err
	}
	body["personId"] = personID
	return g.write(ctx, userID, http.MethodPost, gloutonDraftsPath, body)
}

func (g Glouton) UpdateDraft(ctx context.Context, userID int64, id string, draft GloutonDraft) ([]byte, error) {
	path, err := gloutonItem(gloutonDraftsPath, id)
	if err != nil {
		return nil, err
	}
	body, err := toMap(draft)
	if err != nil {
		return nil, err
	}
	return g.patch(ctx, userID, path, body)
}

func (g Glouton) Suggest(ctx context.Context, userID int64, suggestion GloutonSuggestion) ([]byte, error) {
	if suggestion.Kind == "" || strings.TrimSpace(suggestion.Title) == "" {
		return nil, errors.New("a suggestion needs a kind and a title")
	}
	return g.write(ctx, userID, http.MethodPost, gloutonSuggestionsPath, suggestion)
}
