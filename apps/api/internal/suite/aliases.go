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
	"glouton_contacts":      "glouton_search_contacts",
	"glouton_company":       "glouton_get_company",
	"glouton_person":        "glouton_get_person",
	"glouton_outreach":      "glouton_list_outreach",
	"glouton_suggestions":   "glouton_list_suggestions",
	"create_background_job": "jobs_create",
	"list_background_jobs":  "jobs_list",
	"update_background_job": "jobs_update",
	"delete_background_job": "jobs_delete",
}

// mcpLegacy maps every built-in tool Bulle had before its app moved to MCP
// to the MCP tool that does its job, so a stored transcript renders and
// taints under the current name. Nothing runs them any more.
var mcpLegacy = map[string]string{
	"agenda_events":       "agenda_list_events",
	"agenda_calendars":    "agenda_list_calendars",
	"agenda_event":        "agenda_read_event",
	"agenda_get_event":    "agenda_read_event",
	"courrier_inbox":      "courrier_search_mail",
	"courrier_list_inbox": "courrier_search_mail",
	"courrier_search":     "courrier_search_mail",
	"courrier_read":       "courrier_read_thread",
	"courrier_get_email":  "courrier_read_thread",
	"courrier_accounts":   "courrier_list_accounts",
	"courrier_save_draft": "courrier_draft",
	"courrier_mark":       "courrier_organize",
	"courrier_update":     "courrier_organize",
	"courrier_move":       "courrier_organize",
	"files_list":          "nuage_list_folder",
	"files_search":        "nuage_search",
	"files_read":          "nuage_read_file",
	"files_save":          "nuage_save_file",
	"nuage_list":          "nuage_list_folder",
	"nuage_read":          "nuage_read_file",
	"nuage_save":          "nuage_save_file",
	"opus_tasks":          "opus_search_tasks",
	"opus_list_tasks":     "opus_search_tasks",
	"opus_task":           "opus_read_task",
	"opus_get_task":       "opus_read_task",
	"opus_projects":       "opus_list_projects",
	"opus_project":        "opus_list_projects",
	"opus_get_project":    "opus_list_projects",
	"web_search":          "sonar_search",
	"web_fetch":           "sonar_fetch",
}

var (
	taskKeys   = map[string]string{"id": "task_id", "startDate": "start", "dueDate": "due", "start_date": "start", "due_date": "due", "project": "project_id"}
	projectKey = map[string]string{"id": "project_id"}
	eventKey   = map[string]string{"id": "event_id"}
	eventKeys  = map[string]string{"id": "event_id", "calendar": "calendar_id", "start_at": "start", "end_at": "end", "is_all_day": "all_day"}
	draftKey   = map[string]string{"id": "draft_id"}
	mailKeys   = map[string]string{"account": "account_id", "reply_to": "reply_to_message_id", "id": "draft_id"}
	personKeys = map[string]string{"id": "person_id", "companyId": "company_id", "firstName": "first_name", "lastName": "last_name"}
	refKeys    = map[string]string{"companyId": "company_id", "personId": "person_id"}
	jobKey     = map[string]string{"id": "job_id"}
)

// legacyArgs maps, per current tool name, each argument name an older Bulle
// used to the one the schema advertises now. Nested objects are renamed too.
var legacyArgs = map[string]map[string]string{
	"opus_read_task":          taskKeys,
	"opus_create_task":        taskKeys,
	"opus_update_task":        taskKeys,
	"opus_delete_task":        taskKeys,
	"opus_update_project":     projectKey,
	"opus_delete_project":     projectKey,
	"agenda_read_event":       eventKey,
	"agenda_create_event":     eventKeys,
	"agenda_update_event":     eventKeys,
	"agenda_delete_event":     eventKey,
	"courrier_search_mail":    {"account": "account_id", "unread": "unread_only"},
	"courrier_read_thread":    {"account": "account_id", "email_id": "message_id", "id": "message_id"},
	"courrier_draft":          mailKeys,
	"courrier_send":           mailKeys,
	"courrier_delete_draft":   mailKeys,
	"courrier_organize":       {"account": "account_id", "email_ids": "message_ids", "ids": "message_ids"},
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
	"nuage_list_folder":       {"folder": "folder_id"},
	"nuage_read_file":         {"id": "file_id"},
	"nuage_save_file":         {"folder": "folder_id"},
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
	if current, ok := mcpLegacy[name]; ok {
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
