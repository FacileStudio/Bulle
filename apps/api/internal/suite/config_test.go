package suite

import (
	"slices"
	"testing"
)

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
	for word, want := range map[string]int{"logs": 1, "login": 0, "logo": 0, "docker": 0, "timetable": 0, "times": 1, "tableaux": 1} {
		if got := len(AppsNamed(word)); got != want {
			t.Errorf("AppsNamed(%q) names %d apps, want %d", word, got, want)
		}
	}
}

func TestDiapoAndClapStayOffUntilTheirURLsAreSet(t *testing.T) {
	if cfg, err := FromEnv("i", "c", "s"); err != nil || cfg.MCP[AppDiapo].Enabled() || cfg.MCP[AppClap].Enabled() {
		t.Fatalf("Diapo and Clap must stay off until their URLs are set: %+v, %v", cfg.MCP, err)
	}
	if lists := MCPToolLists(); len(lists[AppDiapo]) != 8 || len(lists[AppClap]) != 11 {
		t.Fatalf("Diapo serves eight tools and Clap eleven: %v %v", lists[AppDiapo], lists[AppClap])
	}
}

func TestFromEnvReadsTheDiapoAndClapMCPServers(t *testing.T) {
	t.Setenv("SUITE_DIAPO_MCP_URL", "http://diapo-api:8080/mcp")
	t.Setenv("SUITE_DIAPO_MCP_RESOURCE", "https://diapo.facile.studio/mcp")
	t.Setenv("SUITE_CLAP_MCP_URL", "https://clap.facile.studio/mcp")
	cfg, err := FromEnv("i", "c", "s")
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.EnabledMCP(); len(got) != 2 || got[0] != AppDiapo || got[1] != AppClap {
		t.Fatalf("Diapo and Clap must be reached through their MCP servers, got %v", got)
	}
	if diapo := cfg.MCP[AppDiapo]; diapo.URL != "http://diapo-api:8080/mcp" || diapo.Site() != "https://diapo.facile.studio" {
		t.Fatalf("unexpected Diapo config %+v", diapo)
	}
	if clap := cfg.MCP[AppClap]; clap.Resource != "https://clap.facile.studio/mcp" {
		t.Fatalf("the Clap resource must default to its URL, got %+v", clap)
	}
	if Label(AppDiapo) != "Diapo" || Label(AppClap) != "Clap" {
		t.Fatalf("labels: %q %q", Label(AppDiapo), Label(AppClap))
	}
	t.Setenv("SUITE_CLAP_MCP_URL", "clap/mcp")
	if _, err := FromEnv("i", "c", "s"); err == nil {
		t.Fatal("a relative Clap MCP URL must stop the boot")
	}
}

func TestDiapoAndClapWordsNameTheirApps(t *testing.T) {
	cases := map[string][]string{
		"slides": {AppDiapo}, "slide": {AppDiapo}, "deck": {AppDiapo}, "presentation": {AppDiapo}, "diapos": {AppDiapo},
		"diaporama": {AppDiapo}, "diapositives": {AppDiapo}, "keynote": {AppDiapo}, "powerpoint": {AppDiapo}, "pitch": {AppDiapo},
		"montage": {AppClap}, "clips": {AppClap}, "clap": {AppClap}, "films": {AppClap}, "motion": {AppClap},
		"animations": {AppClap}, "render": {AppClap}, "rendu": {AppClap}, "timeline": {AppClap},
		"video": {AppClap, AppEcho}, "videos": {AppClap, AppEcho},
		"visio": {AppEcho}, "transcript": {AppEcho}, "time": {AppSablier}, "project": {AppOpus}, "document": {AppDocs},
	}
	for word, want := range cases {
		if got := AppsNamed(word); !slices.Equal(got, want) {
			t.Errorf("AppsNamed(%q) = %v, want %v", word, got, want)
		}
	}
	for text, want := range map[string]string{"Présentation": AppDiapo, "vidéo": AppClap} {
		words := Words(text)
		if len(words) != 1 || !slices.Contains(AppsNamed(words[0]), want) {
			t.Errorf("%q folds to %v, which must name %s", text, words, want)
		}
	}
}
