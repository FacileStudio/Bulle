package suite

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ScopeFor names the target one write tool call aims at, resolved through the
// run's Defaults when the call omits its workspace, account or calendar. The
// scope keys a remembered approval to that target and nothing wider. An error
// means the input could not be read; an empty scope with no error means no
// target was named. A caller must ask every time for both and must never
// treat either as grantable.
func ScopeFor(name string, input []byte, d Defaults) (string, map[string]any, error) {
	switch {
	case strings.HasPrefix(name, "opus_"):
		return scopeOpus(name, input, d)
	case strings.HasPrefix(name, "agenda_"):
		return scopeAgenda(name, input, d)
	case strings.HasPrefix(name, "courrier_update"), strings.HasPrefix(name, "courrier_move"):
		return scopeCourrierBatch(name, input, d)
	case strings.HasPrefix(name, "courrier_"):
		return scopeCourrier(name, input, d)
	case strings.HasPrefix(name, "glouton_"):
		return scopeGlouton(name, input, d)
	case strings.HasPrefix(name, "sablier_"), strings.HasPrefix(name, "files_"):
		return scopeSablierFiles(name, input, d)
	}
	return "", nil, fmt.Errorf("suite: no scope rule for %q", name)
}

func scopeDecode[In any](name string, input []byte, d Defaults, fill func(In, Defaults) (string, map[string]any)) (string, map[string]any, error) {
	var in In
	if err := json.Unmarshal(input, &in); err != nil {
		return "", nil, fmt.Errorf("suite: %s: reading arguments: %w", name, err)
	}
	scope, summary := fill(in, d)
	return scope, summary, nil
}

func scopeValue[T string | int64](kind string, value T) string {
	var zero T
	if value == zero {
		return ""
	}
	return kind + ":" + fmt.Sprint(value)
}

func scopeJoin(base, part string) string {
	if base == "" || part == "" {
		return base
	}
	return base + "/" + part
}

func scopeOpus(name string, input []byte, d Defaults) (string, map[string]any, error) {
	switch name {
	case "opus_create_task":
		return scopeDecode(name, input, d, func(in opusCreateInput, d Defaults) (string, map[string]any) {
			workspace := d.workspace(in.Workspace)
			return scopeJoin(scopeValue("workspace", workspace), scopeValue("project", in.Project)), map[string]any{"resource": "task", "action": "create", "workspace": workspace, "project": in.Project, "title": in.Title}
		})
	case "opus_update_task":
		return scopeDecode(name, input, d, func(in opusUpdateInput, _ Defaults) (string, map[string]any) {
			return scopeValue("task", in.ID), map[string]any{"resource": "task", "action": "update", "task": in.ID}
		})
	case "opus_delete_task":
		return scopeDecode(name, input, d, func(in opusTaskInput, _ Defaults) (string, map[string]any) {
			return scopeValue("task", in.ID), map[string]any{"resource": "task", "action": "delete", "task": in.ID}
		})
	case "opus_create_project":
		return scopeDecode(name, input, d, func(in opusCreateProjectInput, _ Defaults) (string, map[string]any) {
			return scopeValue("workspace", in.Workspace), map[string]any{"resource": "project", "action": "create", "workspace": in.Workspace, "name": in.Name}
		})
	case "opus_update_project":
		return scopeDecode(name, input, d, func(in opusUpdateProjectInput, _ Defaults) (string, map[string]any) {
			return scopeValue("project", in.ID), map[string]any{"resource": "project", "action": "update", "project": in.ID}
		})
	case "opus_delete_project":
		return scopeDecode(name, input, d, func(in opusProjectInput, _ Defaults) (string, map[string]any) {
			return scopeValue("project", in.ID), map[string]any{"resource": "project", "action": "delete", "project": in.ID}
		})
	}
	return "", nil, fmt.Errorf("suite: no scope rule for %q", name)
}

func scopeAgenda(name string, input []byte, d Defaults) (string, map[string]any, error) {
	switch name {
	case "agenda_create_event":
		return scopeDecode(name, input, d, func(in agendaCreateInput, d Defaults) (string, map[string]any) {
			calendar := d.calendar(in.Calendar)
			return scopeValue("calendar", calendar), map[string]any{"resource": "event", "action": "create", "calendar": calendar, "title": in.Title}
		})
	case "agenda_update_event":
		return scopeDecode(name, input, d, func(in agendaUpdateInput, _ Defaults) (string, map[string]any) {
			return scopeValue("event", in.ID), map[string]any{"resource": "event", "action": "update", "event": in.ID}
		})
	case "agenda_delete_event":
		return scopeDecode(name, input, d, func(in agendaEventInput, _ Defaults) (string, map[string]any) {
			return scopeValue("event", in.ID), map[string]any{"resource": "event", "action": "delete", "event": in.ID}
		})
	}
	return "", nil, fmt.Errorf("suite: no scope rule for %q", name)
}

func scopeCourrier(name string, input []byte, d Defaults) (string, map[string]any, error) {
	switch name {
	case "courrier_send":
		return scopeDecode(name, input, d, func(in courrierSendInput, d Defaults) (string, map[string]any) {
			account := d.account(in.Account)
			return scopeValue("account", account), map[string]any{"resource": "email", "action": "send", "account": account, "subject": in.Subject, "to": in.To}
		})
	case "courrier_save_draft":
		return scopeDecode(name, input, d, func(in courrierDraftInput, d Defaults) (string, map[string]any) {
			account := d.account(in.Account)
			scope := scopeValue("account", account)
			draft := scopeValue("draft", in.ID)
			if scope != "" && draft != "" {
				scope += "/" + draft
			}
			return scope, map[string]any{"resource": "draft", "action": "save", "account": account, "draft": in.ID, "subject": in.Subject}
		})
	case "courrier_delete_draft":
		return scopeDecode(name, input, d, func(in courrierEmailInput, d Defaults) (string, map[string]any) {
			account := d.account(in.Account)
			scope := scopeValue("account", account)
			draft := scopeValue("draft", in.ID)
			if scope != "" && draft != "" {
				scope += "/" + draft
			}
			return scope, map[string]any{"resource": "draft", "action": "delete", "account": account, "draft": in.ID}
		})
	}
	return "", nil, fmt.Errorf("suite: no scope rule for %q", name)
}

func scopeCourrierBatch(name string, input []byte, d Defaults) (string, map[string]any, error) {
	switch name {
	case "courrier_update":
		return scopeDecode(name, input, d, func(in courrierUpdateInput, d Defaults) (string, map[string]any) {
			account := d.account(in.Account)
			scope := scopeValue("account", account)
			if scope != "" && in.Action != "" {
				scope += "/action:" + in.Action
			}
			return scope, map[string]any{"resource": "email", "action": in.Action, "account": account, "emails": len(in.IDs)}
		})
	case "courrier_move":
		return scopeDecode(name, input, d, func(in courrierMoveInput, d Defaults) (string, map[string]any) {
			account := d.account(in.Account)
			folder := in.Folder
			if in.FolderID != 0 {
				folder = strconv.FormatInt(in.FolderID, 10)
			}
			scope := scopeValue("account", account)
			if scope != "" && folder != "" {
				scope += "/move:" + folder
			}
			return scope, map[string]any{"resource": "email", "action": "move", "account": account, "folder": folder, "emails": len(in.IDs)}
		})
	}
	return "", nil, fmt.Errorf("suite: no scope rule for %q", name)
}

func scopeGlouton(name string, input []byte, d Defaults) (string, map[string]any, error) {
	switch name {
	case "glouton_create_company":
		return scopeDecode(name, input, d, func(in gloutonCreateCompanyInput, _ Defaults) (string, map[string]any) {
			return "companies", map[string]any{"resource": "company", "action": "create", "name": in.Name}
		})
	case "glouton_update_company":
		return scopeDecode(name, input, d, func(in gloutonUpdateCompanyInput, _ Defaults) (string, map[string]any) {
			return scopeValue("company", in.ID), map[string]any{"resource": "company", "action": "update", "company": in.ID}
		})
	case "glouton_create_person":
		return scopeDecode(name, input, d, func(in GloutonPersonFields, _ Defaults) (string, map[string]any) {
			company := ""
			if in.CompanyID != nil {
				company = *in.CompanyID
			}
			scope := "people"
			if company != "" {
				scope = "company:" + company
			}
			return scope, map[string]any{"resource": "person", "action": "create", "company": company}
		})
	case "glouton_update_person":
		return scopeDecode(name, input, d, func(in gloutonUpdatePersonInput, _ Defaults) (string, map[string]any) {
			return scopeValue("person", in.ID), map[string]any{"resource": "person", "action": "update", "person": in.ID}
		})
	}
	return "", nil, fmt.Errorf("suite: no scope rule for %q", name)
}

func scopeSablierFiles(name string, input []byte, d Defaults) (string, map[string]any, error) {
	switch name {
	case "sablier_start":
		return scopeDecode(name, input, d, func(in sablierStartInput, _ Defaults) (string, map[string]any) {
			scope := ""
			if in.Project != 0 && in.Task != 0 {
				scope = scopeValue("project", in.Project) + "/" + scopeValue("task", in.Task)
			}
			return scope, map[string]any{"resource": "timer", "action": "start", "project": in.Project, "task": in.Task}
		})
	case "sablier_stop":
		return scopeDecode(name, input, d, func(_ sablierStatusInput, _ Defaults) (string, map[string]any) {
			return "timer", map[string]any{"resource": "timer", "action": "stop"}
		})
	case "files_save":
		return scopeDecode(name, input, d, func(in filesSaveInput, _ Defaults) (string, map[string]any) {
			return "folder:" + strconv.FormatInt(in.Folder, 10), map[string]any{"resource": "file", "action": "save", "folder": in.Folder, "name": in.Name}
		})
	}
	return "", nil, fmt.Errorf("suite: no scope rule for %q", name)
}
