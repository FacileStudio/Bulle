package suite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	if !strings.Contains(note, "callout") || len(tools) == 0 {
		t.Fatalf("the artifact note is missing: %q", note)
	}
	w.cfg.Apps[AppMycelium] = AppConfig{}
	_, note, _ = Tools(NewGateway(w.tokens(), w.cfg, nil), 1, Defaults{}, Runtime{})
	for _, name := range artifactToolNames {
		if strings.Contains(note, name) {
			t.Fatalf("%s leaks without Mycelium: %q", name, note)
		}
	}
}

func TestArtifactCreateSignsAsBulleInTheResolvedSpace(t *testing.T) {
	w := newWorld(t)
	tools := teamTools(t, w)
	out, err := runTool(t, tools, "artifact_create", `{"title":"Plan","content":"# Plan\n\nStep one."}`)
	want := `{"artifact":{"id":"2026-11-06-art-1","title":"Plan","space_id":"bulle-work","mycelium_space":"sp-team"}}`
	if err != nil || out != want {
		t.Fatalf("create: %q %v", out, err)
	}
	if body := w.fakes.arts.bodies[0]; body.Machine != ArtifactMachine || body.Title != "Plan" {
		t.Fatalf("create body: %+v", body)
	}
	if path := w.fakes.arts.paths[0]; path != "/api/artifacts?space_id=sp-team" {
		t.Fatalf("create went to %q", path)
	}
	out, err = runTool(t, tools, "artifact_update", `{"artifact_id":"2026-11-06-art-1","content":"Step two."}`)
	if err != nil || out != want {
		t.Fatalf("update: %q %v", out, err)
	}
	read, err := runTool(t, tools, "artifact_read", `{"artifact_id":"2026-11-06-art-1"}`)
	var detail ArtifactDetail
	if err != nil || json.Unmarshal([]byte(read), &detail) != nil || detail.Content != "Step two." {
		t.Fatalf("read must strip the frontmatter: %q %v", read, err)
	}
	if listed, _ := runTool(t, tools, "artifact_list", `{"query":"nothing"}`); listed != "[]" {
		t.Fatalf("filtered list: %q", listed)
	}
}

func ownTools(t *testing.T, w *world, user int64) []nacelle.Tool {
	t.Helper()
	w.addUser(user, future)
	personal := fmt.Sprint("bulle-", user)
	w.fakes.own(personal, fmt.Sprint(user))
	rt := Runtime{CurrentSpace: "home", PersonalSpace: func(context.Context) (string, error) { return personal, nil }}
	tools, _, err := Tools(NewGateway(w.tokens(), w.cfg, nil), user, Defaults{}, rt)
	if err != nil {
		t.Fatal(err)
	}
	return tools
}

func TestPersonalArtifactsGoWhereTheUsersCLISeesThem(t *testing.T) {
	w := newWorld(t)
	w.fakes.own("", "1")
	admin := ownTools(t, w, 1)
	member := ownTools(t, w, 2)
	if _, err := runTool(t, admin, "artifact_create", `{"title":"Mine","content":"x"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := runTool(t, member, "artifact_create", `{"title":"Theirs","content":"x"}`); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(w.fakes.arts.paths, " "); got != "/api/artifacts? /api/artifacts?space_id=bulle-2" {
		t.Fatalf("an admin writes to the common tree, anyone else to their personal space: %s", got)
	}
}

func TestListAllArtifactsKeepsEachUserToTheirOwnTrees(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	w.addUser(2, future)
	w.fakes.own("", "1")
	w.fakes.own("team", "1")
	w.fakes.own("other", "2")
	m := NewMycelium(NewGateway(w.tokens(), w.cfg, nil))
	for _, space := range []string{CommonTree, "team"} {
		if _, err := m.CreateArtifact(context.Background(), 1, space, ArtifactWrite{Title: space, Content: "x"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.CreateArtifact(context.Background(), 2, "other", ArtifactWrite{Title: "other", Content: "x"}); err != nil {
		t.Fatal(err)
	}
	items, err := m.ListAllArtifacts(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	trees := map[string]string{}
	for _, item := range items {
		trees[item.Title] = item.SpaceID
	}
	if len(trees) != 2 || trees[CommonTree] != CommonTree || trees["team"] != "team" {
		t.Fatalf("user 1 sees their common tree and spaces, never user 2's: %v", trees)
	}
}

func TestArtifactToolsRefuseBeforeMycelium(t *testing.T) {
	w := newWorld(t)
	tools := teamTools(t, w)
	inputs := map[string]string{
		"artifact_create": `{"title":"","content":"x"}`,
		"artifact_update": `{"artifact_id":"..","content":"x"}`,
		"artifact_read":   `{"artifact_id":" "}`,
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
	_, getErr := m.GetArtifact(context.Background(), 1, "sp", "a/b c?d")
	var refused *RefusedError
	if !errors.As(getErr, &refused) || refused.Status != http.StatusNotFound {
		t.Fatalf("a missing artifact must keep its 404: %v", getErr)
	}
	if err := m.DeleteArtifact(context.Background(), 1, "sp", "../x"); err == nil {
		t.Fatal("deleting a missing artifact succeeded")
	}
	got := strings.Join(w.fakes.arts.paths, " ")
	if got != "/api/artifacts/a%2Fb%20c%3Fd?space_id=sp /api/artifacts/..%2Fx?space_id=sp" {
		t.Fatalf("ids were not escaped as one segment: %s", got)
	}
}

func TestArtifactResultFitsThePreview(t *testing.T) {
	saved := ArtifactDetail{ArtifactSummary: ArtifactSummary{ID: strings.Repeat("a", 230), Title: strings.Repeat(" <", 400)}}
	out, err := artifactResult(saved, strings.Repeat("s", 36), strings.Repeat("m", 36), nil)
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
