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
