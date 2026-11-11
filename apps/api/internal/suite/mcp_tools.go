package suite

// CourrierTools are the tools Courrier's MCP server exposes.
func CourrierTools() []string {
	return []string{
		"courrier_list_accounts", "courrier_search_mail", "courrier_read_thread",
		"courrier_draft", "courrier_send", "courrier_organize", "courrier_delete_draft",
	}
}

// AgendaTools are the tools Agenda's MCP server exposes.
func AgendaTools() []string {
	return []string{
		"agenda_list_calendars", "agenda_list_events", "agenda_read_event", "agenda_find_free_slots",
		"agenda_create_event", "agenda_update_event", "agenda_delete_event",
	}
}

// SonarTools are the tools sonar's MCP server exposes.
func SonarTools() []string {
	return []string{"sonar_search", "sonar_fetch"}
}

// NuageTools are the tools Nuage's MCP server exposes.
func NuageTools() []string {
	return []string{
		"nuage_list_folder", "nuage_search", "nuage_read_file", "nuage_list_shares",
		"nuage_save_file", "nuage_create_folder", "nuage_move", "nuage_revoke_share",
		"nuage_delete", "nuage_create_share",
	}
}

// OpusTools are the tools Opus's MCP server exposes.
func OpusTools() []string {
	return []string{
		"opus_list_workspaces", "opus_list_projects", "opus_search_tasks", "opus_read_task",
		"opus_create_task", "opus_update_task", "opus_comment_task", "opus_create_project", "opus_update_project",
		"opus_delete_task", "opus_delete_project",
	}
}

// SablierTools are the tools Sablier's MCP server exposes.
func SablierTools() []string {
	return []string{
		"sablier_timer_status", "sablier_list_entries", "sablier_list_projects", "sablier_weekly_summary",
		"sablier_start_timer", "sablier_stop_timer", "sablier_add_entry", "sablier_update_entry", "sablier_delete_entry",
	}
}

// GloutonTools are the tools Glouton's MCP server exposes.
func GloutonTools() []string {
	return append([]string{
		"glouton_contacts_search", "glouton_company_get", "glouton_person_get", "glouton_outreach_list", "glouton_suggestions_list",
	}, GloutonWrites()...)
}

// GloutonWrites are Glouton's tools that change its data. A run view lists
// what they did as the run's outputs, so Glouton can show what an idea made.
func GloutonWrites() []string {
	return []string{
		"glouton_company_create", "glouton_company_update", "glouton_person_create", "glouton_person_update",
		"glouton_draft_save", "glouton_draft_update", "glouton_suggestion_create",
		"glouton_company_delete", "glouton_person_delete",
	}
}
