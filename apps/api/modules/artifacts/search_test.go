package artifacts_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite/myceliumtest"
	"github.com/FacileStudio/Bulle/apps/api/modules/artifacts"
)

func searchWorld(t *testing.T) (*suite.Gateway, *myceliumtest.Server) {
	t.Helper()
	fake := myceliumtest.New(t)
	cfg := suite.Config{Issuer: fake.Issuer(), ClientID: "bulle", ClientSecret: "s3cret",
		MCP: map[string]suite.MCPConfig{suite.AppMycelium: {URL: fake.Resource(), Resource: fake.Resource()}}}
	gw := suite.NewGateway(suite.NewTokens(cfg, fake, nil), cfg)
	t.Cleanup(func() { _ = gw.MCP().Close() })
	return gw, fake
}

func search(t *testing.T, h http.Handler, query, cursor string) (int, artifacts.SearchResponse) {
	t.Helper()
	path := "/artifacts/search?q=" + url.QueryEscape(query)
	if cursor != "" {
		path += "&cursor=" + url.QueryEscape(cursor)
	}
	rec := call(h, http.MethodGet, path)
	var out artifacts.SearchResponse
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("%s: %v", rec.Body, err)
		}
	}
	return rec.Code, out
}

func TestSearchFindsTitlesThenContentInTheCallersTrees(t *testing.T) {
	gw, fake := searchWorld(t)
	fake.MakeAdmin("1")
	team := fake.AddSpace("1", "Team")
	fake.AddArtifact(team, "Budget", "# Budget\n\nWe bought OOLONG for the office.")
	fake.AddArtifact("", "Oolong guide", "# Steeping\n\nThree minutes.")
	fake.AddArtifact(fake.AddSpace("2", "Theirs"), "Oolong private", "secret")

	code, out := search(t, router(gw, 1), "  oolong ", "")
	if code != http.StatusOK || !out.Available || len(out.Artifacts) != 2 {
		t.Fatalf("search = %d %+v", code, out)
	}
	title, body := out.Artifacts[0], out.Artifacts[1]
	if title.Title != "Oolong guide" || title.Match != "title" || title.MyceliumSpace != suite.CommonTree {
		t.Errorf("title hit = %+v", title)
	}
	if body.Title != "Budget" || body.Match != "content" || body.MyceliumSpace != team || body.SpaceName != "Team" || !strings.Contains(body.Snippet, "OOLONG") {
		t.Errorf("content hit = %+v", body)
	}
	if _, out := search(t, router(gw, 1), "oolong office", ""); len(out.Artifacts) != 1 || out.Artifacts[0].Title != "Budget" {
		t.Errorf("every word must match: %+v", out.Artifacts)
	}
	if _, out := search(t, router(gw, 1), "nothing here", ""); out.Artifacts == nil || len(out.Artifacts) != 0 || !out.Available {
		t.Errorf("no match must answer an empty list: %+v", out)
	}
}

func TestSearchPagesWithTheCursor(t *testing.T) {
	gw, fake := searchWorld(t)
	team := fake.AddSpace("1", "Team")
	for i := range myceliumtest.SearchPage + 5 {
		fake.AddArtifact(team, fmt.Sprint("Note ", i), "tea")
	}
	h := router(gw, 1)
	_, first := search(t, h, "tea", "")
	if len(first.Artifacts) != myceliumtest.SearchPage || first.NextCursor == "" {
		t.Fatalf("first page = %d, cursor %q", len(first.Artifacts), first.NextCursor)
	}
	_, second := search(t, h, "tea", first.NextCursor)
	if len(second.Artifacts) != 5 || second.NextCursor != "" {
		t.Fatalf("second page = %d, cursor %q", len(second.Artifacts), second.NextCursor)
	}
	if code, _ := search(t, h, "tea", "bogus"); code != http.StatusBadRequest {
		t.Errorf("bad cursor = %d", code)
	}
	calls := fake.Calls()
	if last := calls[len(calls)-1]; last.Tool != "mycelium_search_artifacts" || last.Args["query"] != "tea" {
		t.Errorf("last call = %+v", last)
	}
}

func TestSearchRefusesBeforeMycelium(t *testing.T) {
	gw, fake := searchWorld(t)
	h := router(gw, 1)
	for name, q := range map[string]string{"empty": "   ", "too long": strings.Repeat("a", 201)} {
		if code, _ := search(t, h, q, ""); code != http.StatusBadRequest {
			t.Errorf("%s: %d", name, code)
		}
	}
	if len(fake.Calls()) != 0 {
		t.Fatalf("Mycelium saw a refused query: %+v", fake.Calls())
	}
	if code, _ := search(t, router(gw, 0), "tea", ""); code != http.StatusUnauthorized {
		t.Errorf("anonymous = %d", code)
	}
	rec := call(router(nil, 1), http.MethodGet, "/artifacts/search?q=tea")
	if rec.Code != http.StatusOK || rec.Body.String() != "{\"artifacts\":[],\"available\":false}\n" {
		t.Errorf("without mycelium: %d %s", rec.Code, rec.Body)
	}
}
