package suite

import "testing"

func TestFromEnvReadsTheCourrierMCPServer(t *testing.T) {
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
	if err != nil || len(cfg.EnabledMCP()) != 0 {
		t.Fatalf("want no app enabled, got %+v, %v", cfg.EnabledMCP(), err)
	}
}

func TestFromEnvReadsTheAgendaMCPServer(t *testing.T) {
	t.Setenv("SUITE_AGENDA_MCP_URL", "https://agenda.test/mcp")
	cfg, err := FromEnv("i", "c", "s")
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.EnabledMCP(); len(got) != 1 || got[0] != AppAgenda || cfg.MCP[AppAgenda].Resource != "https://agenda.test/mcp" {
		t.Fatalf("Agenda must be reached through its MCP server, got %+v", cfg.MCP)
	}
}

func TestFromEnvReadsTheDocsAndFeuillesMCPServers(t *testing.T) {
	t.Setenv("SUITE_DOCS_MCP_URL", "http://docs-api:8080/mcp")
	t.Setenv("SUITE_DOCS_MCP_RESOURCE", "https://docs.facile.studio/mcp")
	t.Setenv("SUITE_FEUILLES_MCP_URL", "https://feuilles.facile.studio/mcp")
	cfg, err := FromEnv("i", "c", "s")
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.EnabledMCP(); len(got) != 2 || got[0] != AppDocs || got[1] != AppFeuilles {
		t.Fatalf("Docs and Feuilles must be reached through their MCP servers, got %v", got)
	}
	if docs := cfg.MCP[AppDocs]; docs.URL != "http://docs-api:8080/mcp" || docs.Site() != "https://docs.facile.studio" {
		t.Fatalf("unexpected Docs config %+v", docs)
	}
	if feuilles := cfg.MCP[AppFeuilles]; feuilles.Resource != "https://feuilles.facile.studio/mcp" {
		t.Fatalf("the Feuilles resource must default to its URL, got %+v", feuilles)
	}
	if Label(AppDocs) != "Docs" || Label(AppFeuilles) != "Feuilles" {
		t.Fatalf("labels: %q %q", Label(AppDocs), Label(AppFeuilles))
	}
	t.Setenv("SUITE_FEUILLES_MCP_URL", "feuilles/mcp")
	if _, err := FromEnv("i", "c", "s"); err == nil {
		t.Fatal("a relative Feuilles MCP URL must stop the boot")
	}
}

func TestDocsAndFeuillesWordsNameTheirApps(t *testing.T) {
	cases := map[string]string{"documents": AppDocs, "notes": AppDocs, "pages": AppDocs, "tableur": AppFeuilles, "spreadsheets": AppFeuilles, "classeur": AppFeuilles, "workbooks": AppFeuilles}
	for word, app := range cases {
		if got := AppsNamed(word); len(got) != 1 || got[0] != app {
			t.Errorf("AppsNamed(%q) = %v, want [%s]", word, got, app)
		}
	}
}

func TestAppWordsMatchWholeWordsAndPlurals(t *testing.T) {
	for word, want := range map[string]int{"logs": 1, "login": 0, "logo": 0, "docker": 0, "timeline": 0, "times": 1, "tableaux": 1} {
		if got := len(AppsNamed(word)); got != want {
			t.Errorf("AppsNamed(%q) names %d apps, want %d", word, got, want)
		}
	}
}
