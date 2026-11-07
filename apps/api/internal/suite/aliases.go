package suite

import (
	"context"
	"encoding/json"

	"github.com/FacileStudio/nacelle"
)

// legacyToolNames maps each tool name an older Bulle gave the model to the
// name it has now, so a stored transcript or a pending approval that still
// carries the old name replays.
var legacyToolNames = map[string]string{
	"opus_tasks":            "opus_list_tasks",
	"opus_task":             "opus_get_task",
	"opus_projects":         "opus_list_projects",
	"opus_project":          "opus_get_project",
	"agenda_events":         "agenda_list_events",
	"agenda_event":          "agenda_get_event",
	"agenda_calendars":      "agenda_list_calendars",
	"courrier_inbox":        "courrier_list_inbox",
	"courrier_read":         "courrier_get_email",
	"courrier_accounts":     "courrier_list_accounts",
	"glouton_contacts":      "glouton_search_contacts",
	"glouton_company":       "glouton_get_company",
	"glouton_person":        "glouton_get_person",
	"glouton_outreach":      "glouton_list_outreach",
	"glouton_suggestions":   "glouton_list_suggestions",
	"files_list":            "nuage_list",
	"files_search":          "nuage_search",
	"files_read":            "nuage_read",
	"files_save":            "nuage_save",
	"create_background_job": "jobs_create",
	"list_background_jobs":  "jobs_list",
	"update_background_job": "jobs_update",
	"delete_background_job": "jobs_delete",
}

var (
	taskKeys   = map[string]string{"id": "task_id", "startDate": "start_date", "dueDate": "due_date"}
	projectKey = map[string]string{"id": "project_id"}
	eventKey   = map[string]string{"id": "event_id"}
	draftKey   = map[string]string{"id": "draft_id"}
	emailsKey  = map[string]string{"ids": "email_ids"}
	personKeys = map[string]string{"id": "person_id", "companyId": "company_id", "firstName": "first_name", "lastName": "last_name"}
	refKeys    = map[string]string{"companyId": "company_id", "personId": "person_id"}
	jobKey     = map[string]string{"id": "job_id"}
)

// legacyArgs maps, per current tool name, each argument name an older Bulle
// used to the one the schema advertises now. Nested objects are renamed too.
var legacyArgs = map[string]map[string]string{
	"opus_get_task":           taskKeys,
	"opus_create_task":        taskKeys,
	"opus_update_task":        taskKeys,
	"opus_delete_task":        taskKeys,
	"opus_get_project":        projectKey,
	"opus_update_project":     projectKey,
	"opus_delete_project":     projectKey,
	"agenda_get_event":        eventKey,
	"agenda_update_event":     eventKey,
	"agenda_delete_event":     eventKey,
	"courrier_get_email":      {"id": "email_id"},
	"courrier_save_draft":     draftKey,
	"courrier_delete_draft":   draftKey,
	"courrier_update":         emailsKey,
	"courrier_mark":           emailsKey,
	"courrier_move":           emailsKey,
	"glouton_search_contacts": {"q": "query"},
	"glouton_get_company":     {"id": "company_id"},
	"glouton_get_person":      {"id": "person_id"},
	"glouton_list_outreach":   refKeys,
	"glouton_save_draft":      refKeys,
	"glouton_update_draft":    draftKey,
	"glouton_suggest":         refKeys,
	"glouton_create_company":  personKeys,
	"glouton_update_company":  {"id": "company_id"},
	"glouton_create_person":   personKeys,
	"glouton_update_person":   personKeys,
	"nuage_read":              {"id": "file_id"},
	"artifact_read":           {"id": "artifact_id"},
	"artifact_update":         {"id": "artifact_id"},
	"jobs_update":             jobKey,
	"jobs_delete":             jobKey,
}

// CanonicalTool answers the current name of a tool, given a name an older
// Bulle used or the current one.
func CanonicalTool(name string) string {
	if current, ok := legacyToolNames[name]; ok {
		return current
	}
	return name
}

// CanonicalInput renames the old argument names of a call to the current
// tool name's arguments. An argument already given its new name wins, and an
// input that is not a JSON object comes back unchanged.
func CanonicalInput(name string, input json.RawMessage) json.RawMessage {
	keys := legacyArgs[CanonicalTool(name)]
	if keys == nil || len(input) == 0 {
		return input
	}
	var decoded any
	if err := json.Unmarshal(input, &decoded); err != nil {
		return input
	}
	if _, ok := decoded.(map[string]any); !ok {
		return input
	}
	out, err := json.Marshal(renameKeys(decoded, keys))
	if err != nil {
		return input
	}
	return out
}

func renameKeys(value any, keys map[string]string) any {
	switch v := value.(type) {
	case []any:
		for i, item := range v {
			v[i] = renameKeys(item, keys)
		}
		return v
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			out[renamedKey(v, key, keys)] = renameKeys(item, keys)
		}
		return out
	}
	return value
}

func renamedKey(object map[string]any, key string, keys map[string]string) string {
	current, ok := keys[key]
	if _, taken := object[current]; !ok || taken {
		return key
	}
	return current
}

// Aliased wraps a tool so its old argument names still decode, or answers it
// unchanged when it never had any.
func Aliased(tool nacelle.Tool) nacelle.Tool {
	if tool == nil || legacyArgs[tool.Spec().Name] == nil {
		return tool
	}
	name := tool.Spec().Name
	return nacelle.Wrap(tool, nil, func(next nacelle.RunFunc) nacelle.RunFunc {
		return func(ctx context.Context, call nacelle.Call, emit func(string)) (nacelle.Result, error) {
			call.Input = CanonicalInput(name, call.Input)
			return next(ctx, call, emit)
		}
	})
}
