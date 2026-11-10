package suite

import "testing"

func TestFromEnvRequiresPairs(t *testing.T) {
	t.Setenv("SUITE_AGENDA_URL", "https://agenda.test/")
	if _, err := FromEnv("i", "c", "s"); err == nil {
		t.Fatal("a URL without an audience must stop the boot")
	}
	t.Setenv("SUITE_AGENDA_AUDIENCE", "aud")
	cfg, err := FromEnv("i", "c", "s")
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Enabled(); len(got) != 1 || got[0] != AppAgenda || cfg.Apps[AppAgenda].URL != "https://agenda.test" {
		t.Fatalf("unexpected config %+v", cfg)
	}
}

func TestFromEnvReadsTheCourrierMCPServer(t *testing.T) {
	t.Setenv("SUITE_COURRIER_URL", "https://courrier.test")
	t.Setenv("SUITE_COURRIER_MCP_URL", "http://courrier-api:8080/mcp")
	t.Setenv("SUITE_COURRIER_MCP_RESOURCE", "https://courrier.test/mcp")
	cfg, err := FromEnv("i", "c", "s")
	if err != nil {
		t.Fatal(err)
	}
	server := cfg.MCP[AppCourrier]
	if got := cfg.EnabledMCP(); len(got) != 1 || server.Resource != "https://courrier.test/mcp" || server.Site() != "https://courrier.test" {
		t.Fatalf("unexpected MCP config %+v", cfg.MCP)
	}
	if len(cfg.Enabled()) != 0 {
		t.Fatalf("Courrier has no built-in client any more, got %v", cfg.Enabled())
	}
	t.Setenv("SUITE_COURRIER_MCP_RESOURCE", "")
	if cfg, _ = FromEnv("i", "c", "s"); cfg.MCP[AppCourrier].Resource != "http://courrier-api:8080/mcp" {
		t.Fatalf("the resource must default to the URL, got %+v", cfg.MCP[AppCourrier])
	}
	t.Setenv("SUITE_COURRIER_MCP_URL", "courrier/mcp")
	if _, err := FromEnv("i", "c", "s"); err == nil {
		t.Fatal("a relative MCP URL must stop the boot")
	}
}

func TestFromEnvEmptyDisablesEverything(t *testing.T) {
	cfg, err := FromEnv("i", "c", "s")
	if err != nil || len(cfg.Enabled()) != 0 {
		t.Fatalf("want no app enabled, got %+v, %v", cfg.Enabled(), err)
	}
}
