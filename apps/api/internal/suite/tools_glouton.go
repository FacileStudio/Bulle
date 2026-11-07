package suite

import (
	"context"
	"fmt"
)

type gloutonContactsInput struct {
	Query       string `json:"query,omitempty" jsonschema:"description=Text to match against names, emails and domains"`
	Stage       string `json:"stage,omitempty" jsonschema:"enum=PROSPECT,enum=CONTACTED,enum=CLIENT,enum=PARTNER,enum=PERSONAL"`
	Temperature string `json:"temperature,omitempty" jsonschema:"enum=HOT,enum=WARM,enum=COLD"`
	Limit       int    `json:"limit,omitempty" jsonschema:"description=Maximum results (default 25, max 100)"`
}

type gloutonCompanyRef struct {
	CompanyID string `json:"company_id" jsonschema:"required,description=Company id as glouton_search_contacts answers it"`
}

type gloutonPersonRef struct {
	PersonID string `json:"person_id" jsonschema:"required,description=Person id as glouton_search_contacts answers it"`
}

type gloutonOutreachInput struct {
	CompanyID string `json:"company_id,omitempty" jsonschema:"description=Company id; pass this or person_id"`
	PersonID  string `json:"person_id,omitempty" jsonschema:"description=Person id; pass this or company_id"`
}

type gloutonSuggestionsInput struct {
	Status string `json:"status,omitempty" jsonschema:"enum=OPEN,enum=ACCEPTED,enum=DISMISSED,description=Defaults to OPEN"`
}

type gloutonSaveDraftInput struct {
	PersonID string `json:"person_id" jsonschema:"required,description=The person the email is for"`
	GloutonDraft
}

type gloutonUpdateDraftInput struct {
	DraftID string `json:"draft_id" jsonschema:"required,description=Draft id, as returned by glouton_save_draft"`
	GloutonDraft
}

type gloutonSuggestInput struct {
	Kind        string `json:"kind" jsonschema:"required,enum=FOLLOW_UP,enum=RESCORE,enum=NEXT_ACTION,enum=MATCH,description=RESCORE proposes a new temperature or score; NEXT_ACTION and FOLLOW_UP describe the step in body"`
	Title       string `json:"title" jsonschema:"required,description=One line the human reads in the inbox"`
	Body        string `json:"body,omitempty" jsonschema:"description=Details of the proposed action"`
	Rationale   string `json:"rationale,omitempty" jsonschema:"description=The evidence behind the suggestion"`
	CompanyID   string `json:"company_id,omitempty"`
	PersonID    string `json:"person_id,omitempty"`
	Temperature string `json:"temperature,omitempty" jsonschema:"enum=HOT,enum=WARM,enum=COLD,description=RESCORE only: the proposed temperature"`
	Score       *int   `json:"score,omitempty" jsonschema:"description=RESCORE only: the proposed score from 0 to 100"`
}

type gloutonCreateCompanyInput struct {
	GloutonCompanyFields
	People []GloutonNewPerson `json:"people,omitempty" jsonschema:"description=People to create at this company"`
}

type gloutonUpdateCompanyInput struct {
	CompanyID string `json:"company_id" jsonschema:"required,description=Company id as glouton_search_contacts answers it"`
	GloutonCompanyPatch
}

type gloutonUpdatePersonInput struct {
	PersonID string `json:"person_id" jsonschema:"required,description=Person id as glouton_search_contacts answers it"`
	GloutonPersonFields
}

func gloutonTools(set *toolSet, g Glouton) {
	gloutonReadTools(set, g)
	gloutonInboxTools(set, g)
	writeTool(set, "glouton_create_company", "Create a company in Glouton, optionally with its people. "+
		"Search glouton_search_contacts first so you do not create a duplicate. name is required; domain is the bare domain such as example.com.",
		func(ctx context.Context, in gloutonCreateCompanyInput) (string, error) {
			return rawResult(g.CreateCompany(ctx, set.userID, in.GloutonCompanyFields, in.People))
		})
	writeTool(set, "glouton_update_company", "Change a Glouton company: its details, stage, temperature or score. Only the fields you pass change. "+
		"company_id comes from glouton_search_contacts.",
		func(ctx context.Context, in gloutonUpdateCompanyInput) (string, error) {
			return rawResult(g.UpdateCompany(ctx, set.userID, in.CompanyID, in.GloutonCompanyPatch))
		})
	writeTool(set, "glouton_create_person", "Create a person in Glouton, linked to a company when you pass company_id. "+
		"Search glouton_search_contacts first so you do not create a duplicate. Pass at least a name or an email.",
		func(ctx context.Context, in GloutonPersonFields) (string, error) {
			return rawResult(g.CreatePerson(ctx, set.userID, in))
		})
	writeTool(set, "glouton_update_person", "Change a Glouton person. Only the fields you pass change. "+
		"person_id comes from glouton_search_contacts or glouton_get_company.",
		func(ctx context.Context, in gloutonUpdatePersonInput) (string, error) {
			return rawResult(g.UpdatePerson(ctx, set.userID, in.PersonID, in.GloutonPersonFields))
		})
}

func gloutonInboxTools(set *toolSet, g Glouton) {
	inbox := fmt.Sprintf(" Nothing is ever sent: the human reviews and sends it from Glouton's agent inbox at %s.", g.InboxURL())
	selfTool(set, "glouton_save_draft", "Save an outreach email draft for a Glouton contact."+inbox+" Tell the user where to find it. "+
		"person_id comes from glouton_search_contacts; subject and text are required.",
		func(ctx context.Context, in gloutonSaveDraftInput) (string, error) {
			return rawResult(g.SaveDraft(ctx, set.userID, in.PersonID, in.GloutonDraft))
		})
	selfTool(set, "glouton_update_draft", "Change a Glouton draft that is still unsent. Only the fields you pass change."+inbox,
		func(ctx context.Context, in gloutonUpdateDraftInput) (string, error) {
			return rawResult(g.UpdateDraft(ctx, set.userID, in.DraftID, in.GloutonDraft))
		})
	selfTool(set, "glouton_suggest", "Propose a follow-up, a rescore or a next action on a Glouton company or person. "+
		"The human accepts or dismisses it in Glouton's agent inbox at "+g.InboxURL()+". "+
		"Read glouton_list_suggestions first so you do not repeat an open one.",
		func(ctx context.Context, in gloutonSuggestInput) (string, error) {
			return rawResult(g.Suggest(ctx, set.userID, in.suggestion()))
		})
}

func gloutonReadTools(set *toolSet, g Glouton) {
	readTool(set, "glouton_search_contacts", "Search the user's prospects and contacts in Glouton, companies and people together, with their stage and temperature. "+
		"Use it to find a company_id or person_id, or to pick prospects by temperature. query matches names, emails and domains.",
		func(ctx context.Context, in gloutonContactsInput) (string, error) {
			return rawResult(g.Contacts(ctx, set.userID, GloutonContactQuery(in)))
		})
	readTool(set, "glouton_get_company", "Show one Glouton company with its people, recent activity, open suggestions and website audit. "+
		"company_id comes from glouton_search_contacts.",
		func(ctx context.Context, in gloutonCompanyRef) (string, error) {
			return rawResult(g.Company(ctx, set.userID, in.CompanyID))
		})
	readTool(set, "glouton_get_person", "Show one Glouton person with their company, activity and outreach emails. "+
		"person_id comes from glouton_search_contacts.",
		func(ctx context.Context, in gloutonPersonRef) (string, error) {
			return rawResult(g.Person(ctx, set.userID, in.PersonID))
		})
	readTool(set, "glouton_list_outreach", "List the outreach emails of a Glouton company or person: status, subject, sent and replied dates, and whether the agent drafted them. "+
		"Pass company_id or person_id.",
		func(ctx context.Context, in gloutonOutreachInput) (string, error) {
			return rawResult(g.Outreach(ctx, set.userID, in.CompanyID, in.PersonID))
		})
	readTool(set, "glouton_list_suggestions", "List the suggestions in Glouton's agent inbox, open ones by default, so you do not repeat one. "+
		"status filters by OPEN, ACCEPTED or DISMISSED.",
		func(ctx context.Context, in gloutonSuggestionsInput) (string, error) {
			return rawResult(g.Suggestions(ctx, set.userID, in.Status))
		})
}

func (in gloutonSuggestInput) suggestion() GloutonSuggestion {
	out := GloutonSuggestion{Kind: in.Kind, Title: in.Title, Body: in.Body, Rationale: in.Rationale, CompanyID: in.CompanyID, PersonID: in.PersonID}
	payload := map[string]any{}
	if in.Temperature != "" {
		payload["temperature"] = in.Temperature
	}
	if in.Score != nil {
		payload["score"] = *in.Score
	}
	if len(payload) > 0 {
		out.Payload = payload
	}
	return out
}
