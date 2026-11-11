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

// PlumeTools are the tools Plume's MCP server exposes.
func PlumeTools() []string {
	return []string{
		"plume_list_documents", "plume_get_document", "plume_list_clients", "plume_create_document",
		"plume_update_document", "plume_add_signer", "plume_save_client", "plume_send_document",
		"plume_remind_signer", "plume_delete_document",
	}
}

// CapsuleTools are the tools Capsule's MCP server exposes.
func CapsuleTools() []string {
	return []string{"capsule_get_status", "capsule_revoke"}
}

// JournalTools are the tools Journal's MCP server exposes.
func JournalTools() []string {
	return []string{
		"journal_search_logs", "journal_read_log", "journal_count_logs", "journal_list_apps",
		"journal_list_saved_queries", "journal_save_query", "journal_list_alerts", "journal_create_alert",
		"journal_set_alert_enabled", "journal_delete_saved_query", "journal_delete_alert",
	}
}

// AntenneTools are the tools Antenne's MCP server exposes.
func AntenneTools() []string {
	return []string{
		"antenne_list_events", "antenne_get_event", "antenne_list_providers", "antenne_list_targets",
		"antenne_pool_status", "antenne_test_provider", "antenne_check_target", "antenne_test_target",
		"antenne_set_enabled", "antenne_replay_event",
	}
}

// VisionTools are the tools Vision's MCP server exposes.
func VisionTools() []string {
	return []string{
		"vision_list_spaces", "vision_list_sites", "vision_get_stats", "vision_list_goals",
		"vision_create_site", "vision_update_site", "vision_unshare_site", "vision_create_goal",
		"vision_update_goal", "vision_delete_site", "vision_share_site", "vision_delete_goal",
	}
}

// EchoTools are the tools Echo's MCP server exposes.
func EchoTools() []string {
	return []string{
		"echo_list_rooms", "echo_list_meetings", "echo_get_meeting", "echo_read_transcript",
		"echo_create_room", "echo_rename_room", "echo_summarize_meeting", "echo_share_recording",
		"echo_delete_room",
	}
}

// ToileTools are the tools Toile's MCP server exposes.
func ToileTools() []string {
	return []string{
		"toile_list_boards", "toile_read_board", "toile_add_shapes", "toile_update_shapes",
		"toile_delete_shapes", "toile_comment",
	}
}
