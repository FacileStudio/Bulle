package suite

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite/myceliumtest"
	"github.com/FacileStudio/nacelle"
)

var artifactToolNames = []string{"artifact_create", "artifact_update", "artifact_list", "artifact_read"}

func teamTools(t *testing.T) ([]nacelle.Tool, *myceliumtest.Server, string) {
	t.Helper()
	gw, fake := myceliumWorld(t)
	team := fake.AddSpace("1", "Team")
	tools, _, err := Tools(gw, 1, Defaults{MyceliumSpace: team}, Runtime{CurrentSpace: "bulle-work"})
	if err != nil {
		t.Fatal(err)
	}
	return tools, fake, team
}

func TestArtifactToolsNeedMycelium(t *testing.T) {
	gw, _ := myceliumWorld(t)
	tools, note, _ := Tools(gw, 1, Defaults{}, Runtime{})
	for _, name := range artifactToolNames {
		if !strings.Contains(note, name) {
			t.Errorf("%s is missing with Mycelium on", name)
		}
	}
	if !strings.Contains(note, "callout") || len(tools) == 0 {
		t.Fatalf("the artifact note is missing: %q", note)
	}
	w := newWorld(t)
	_, note, _ = Tools(NewGateway(w.tokens(), w.cfg), 1, Defaults{}, Runtime{})
	for _, name := range artifactToolNames {
		if strings.Contains(note, name) {
			t.Fatalf("%s leaks without Mycelium: %q", name, note)
		}
	}
}

func TestArtifactCreateLandsInTheResolvedSpace(t *testing.T) {
	tools, fake, team := teamTools(t)
	out, err := runTool(t, tools, "artifact_create", `{"title":"Plan","content":"# Plan\n\nStep one."}`)
	var ref map[string]artifactRef
	if err != nil || json.Unmarshal([]byte(out), &ref) != nil || ref["artifact"].SpaceID != "bulle-work" || ref["artifact"].MyceliumSpace != team {
		t.Fatalf("create: %q %v", out, err)
	}
	id := ref["artifact"].ID
	if got := fake.Artifacts(team); len(got) != 1 || got[0].Title != "Plan" {
		t.Fatalf("create landed elsewhere: %+v", got)
	}
	if _, err := runTool(t, tools, "artifact_update", `{"artifact_id":"`+id+`","content":"Step two."}`); err != nil {
		t.Fatalf("update: %v", err)
	}
	read, err := runTool(t, tools, "artifact_read", `{"artifact_id":"`+id+`"}`)
	var detail ArtifactDetail
	if err != nil || json.Unmarshal([]byte(read), &detail) != nil || detail.Content != "Step two." || detail.Title != "Plan" {
		t.Fatalf("read must strip the frontmatter: %q %v", read, err)
	}
	if listed, _ := runTool(t, tools, "artifact_list", `{"query":"nothing"}`); listed != "[]" {
		t.Fatalf("filtered list: %q", listed)
	}
}

func TestPersonalArtifactsGoWhereTheUsersCLISeesThem(t *testing.T) {
	gw, fake := myceliumWorld(t)
	fake.MakeAdmin("1")
	personal := map[int64]string{1: fake.AddSpace("1", "Bulle"), 2: fake.AddSpace("2", "Bulle")}
	for user, title := range map[int64]string{1: "Mine", 2: "Theirs"} {
		rt := Runtime{CurrentSpace: "home", PersonalSpace: func(context.Context) (string, error) { return personal[user], nil }}
		tools, _, err := Tools(gw, user, Defaults{}, rt)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := runTool(t, tools, "artifact_create", `{"title":"`+title+`","content":"x"}`); err != nil {
			t.Fatal(err)
		}
	}
	if common, theirs := fake.Artifacts(""), fake.Artifacts(personal[2]); len(common) != 1 || common[0].Title != "Mine" || len(theirs) != 1 {
		t.Fatalf("an admin writes to the common tree, anyone else to their personal space: %+v %+v", common, theirs)
	}
}

func TestListAllArtifactsKeepsEachUserToTheirOwnTrees(t *testing.T) {
	gw, fake := myceliumWorld(t)
	fake.MakeAdmin("1")
	team, other := fake.AddSpace("1", "team"), fake.AddSpace("2", "other")
	m := NewMycelium(gw)
	for _, space := range []string{CommonTree, team} {
		if _, err := m.CreateArtifact(context.Background(), 1, space, ArtifactWrite{Title: space, Content: "x"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.CreateArtifact(context.Background(), 2, other, ArtifactWrite{Title: "other", Content: "x"}); err != nil {
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
	if len(trees) != 2 || trees[CommonTree] != CommonTree || trees[team] != team {
		t.Fatalf("user 1 sees their common tree and spaces, never user 2's: %v", trees)
	}
}

func TestArtifactToolsRefuseBeforeMycelium(t *testing.T) {
	tools, fake, _ := teamTools(t)
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
	for _, call := range fake.Calls() {
		if strings.Contains(call.Tool, "artifact") {
			t.Fatalf("a refused call reached Mycelium: %+v", call)
		}
	}
}

func TestMissingArtifactsKeepTheir404(t *testing.T) {
	gw, fake := myceliumWorld(t)
	space := fake.AddSpace("1", "sp")
	m := NewMycelium(gw)
	_, err := m.GetArtifact(context.Background(), 1, space, "a/b c?d")
	var refused *RefusedError
	if !errors.As(err, &refused) || refused.Status != http.StatusNotFound {
		t.Fatalf("a missing artifact must keep its 404: %v", err)
	}
	if err := m.DeleteArtifact(WithConfirmed(context.Background()), 1, space, "x"); err == nil {
		t.Fatal("deleting a missing artifact succeeded")
	}
}

func TestArtifactDeleteRunsOnlyConfirmed(t *testing.T) {
	gw, fake := myceliumWorld(t)
	space := fake.AddSpace("1", "sp")
	m := NewMycelium(gw)
	saved, err := m.CreateArtifact(context.Background(), 1, space, ArtifactWrite{Title: "Plan", Content: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.DeleteArtifact(context.Background(), 1, space, saved.ID); !errors.Is(err, ErrDeclined) || len(fake.Artifacts(space)) != 1 {
		t.Fatalf("an unanswered delete must keep the artifact: %v", err)
	}
	if err := m.DeleteArtifact(WithConfirmed(context.Background()), 1, space, saved.ID); err != nil || len(fake.Artifacts(space)) != 0 {
		t.Fatalf("a confirmed delete must run: %v", err)
	}
}

func TestArtifactResultFitsThePreview(t *testing.T) {
	saved := ArtifactDetail{ArtifactSummary: ArtifactSummary{ID: strings.Repeat("a", 230), Title: strings.Repeat(" <", 400)}}
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
