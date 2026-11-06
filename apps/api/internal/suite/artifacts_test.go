package suite

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/FacileStudio/nacelle"
)

var artifactToolNames = []string{"artifact_create", "artifact_update", "artifact_list", "artifact_read"}

func teamTools(t *testing.T, w *world) []nacelle.Tool {
	t.Helper()
	w.addUser(1, future)
	w.fakes.own("sp-team", "1")
	rt := Runtime{CurrentSpace: "bulle-work"}
	tools, _, err := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{MyceliumSpace: "sp-team"}, rt)
	if err != nil {
		t.Fatal(err)
	}
	return tools
}

func TestArtifactToolsNeedMycelium(t *testing.T) {
	w := newWorld(t)
	tools, note, _ := Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, Runtime{})
	for _, name := range artifactToolNames {
		if !strings.Contains(note, name) {
			t.Errorf("%s is missing with Mycelium on", name)
		}
	}
	if !strings.Contains(note, "chart:bar") || len(tools) == 0 {
		t.Fatalf("the artifact note is missing: %q", note)
	}
	w.cfg.Apps[AppMycelium] = AppConfig{}
	_, note, _ = Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, Runtime{})
	if strings.Contains(note, "artifact") {
		t.Fatalf("artifact tools leak without Mycelium: %q", note)
	}
}

func TestArtifactCreateSignsAsBulleInTheResolvedSpace(t *testing.T) {
	w := newWorld(t)
	tools := teamTools(t, w)
	out, err := runTool(t, tools, "artifact_create", `{"title":"Plan","content":"# Plan\n\nStep one."}`)
	want := `{"artifact":{"id":"2026-11-06-art-1","title":"Plan","space_id":"bulle-work"}}`
	if err != nil || out != want {
		t.Fatalf("create: %q %v", out, err)
	}
	if body := w.fakes.arts.bodies[0]; body.Machine != ArtifactMachine || body.Title != "Plan" {
		t.Fatalf("create body: %+v", body)
	}
	if path := w.fakes.arts.paths[0]; path != "/api/artifacts?space_id=sp-team" {
		t.Fatalf("create went to %q", path)
	}
	out, err = runTool(t, tools, "artifact_update", `{"id":"2026-11-06-art-1","content":"Step two."}`)
	if err != nil || out != want {
		t.Fatalf("update: %q %v", out, err)
	}
	read, err := runTool(t, tools, "artifact_read", `{"id":"2026-11-06-art-1"}`)
	var detail ArtifactDetail
	if err != nil || json.Unmarshal([]byte(read), &detail) != nil || detail.Content != "Step two." {
		t.Fatalf("read must strip the frontmatter: %q %v", read, err)
	}
	if listed, _ := runTool(t, tools, "artifact_list", `{"query":"nothing"}`); listed != "[]" {
		t.Fatalf("filtered list: %q", listed)
	}
}

func TestArtifactToolsRefuseBeforeMycelium(t *testing.T) {
	w := newWorld(t)
	tools := teamTools(t, w)
	inputs := map[string]string{
		"artifact_create": `{"title":"","content":"x"}`,
		"artifact_update": `{"id":"..","content":"x"}`,
		"artifact_read":   `{"id":" "}`,
	}
	for name, input := range inputs {
		if _, err := runTool(t, tools, name, input); err == nil {
			t.Errorf("%s accepted %s", name, input)
		}
	}
	if _, err := runTool(t, tools, "artifact_create", `{"title":"`+strings.Repeat("t", 201)+`","content":"x"}`); err == nil {
		t.Error("a 201 character title was accepted")
	}
	if len(w.fakes.arts.paths) != 0 {
		t.Fatalf("a refused call reached Mycelium: %v", w.fakes.arts.paths)
	}
}

func TestArtifactIDsArePathEscaped(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.fakes.own("sp", "1")
	m := NewMycelium(NewGateway(w.tokens(), w.cfg, nil))
	_, _ = m.GetArtifact(context.Background(), 1, "sp", "a/b c?d")
	_ = m.DeleteArtifact(context.Background(), 1, "sp", "../x")
	got := strings.Join(w.fakes.arts.paths, " ")
	if got != "/api/artifacts/a%2Fb%20c%3Fd?space_id=sp /api/artifacts/..%2Fx?space_id=sp" {
		t.Fatalf("ids were not escaped as one segment: %s", got)
	}
}

func TestArtifactResultFitsThePreview(t *testing.T) {
	saved := ArtifactDetail{ArtifactSummary: ArtifactSummary{ID: strings.Repeat("a", 230), Title: strings.Repeat(" <", 400)}}
	out, err := artifactResult(saved, strings.Repeat("s", 36), nil)
	var parsed map[string]artifactRef
	if err != nil || len(out) >= 2000 || json.Unmarshal([]byte(out), &parsed) != nil {
		t.Fatalf("%d bytes, %v", len(out), err)
	}
	if got := []rune(parsed["artifact"].Title); len(got) != maxArtifactTitle {
		t.Fatalf("title kept %d runes", len(got))
	}
}

func TestStripFrontmatter(t *testing.T) {
	cases := map[string]string{
		"---\ntitle: \"x\"\n---\n\n# Body\n": "# Body\n",
		"---\r\ntitle: x\r\n---\r\nBody":     "Body",
		"# No frontmatter":                   "# No frontmatter",
		"---\nunclosed":                      "---\nunclosed",
		"---\ntitle: x\n---":                 "",
	}
	for in, want := range cases {
		if got := stripFrontmatter(in); got != want {
			t.Errorf("stripFrontmatter(%q) = %q, want %q", in, got, want)
		}
	}
}
