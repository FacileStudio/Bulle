package artifacts_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/encre"
	"github.com/FacileStudio/Bulle/apps/api/internal/encre/encretest"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
)

func put(h http.Handler, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	return rec
}

func TestUpdateRefusesBeforeMyceliumSeesIt(t *testing.T) {
	encretest.Install(t)
	if rec := put(router(nil, 1), "/artifacts/a?mycelium_space=common", `{"content":"x"}`); rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("without mycelium: %d", rec.Code)
	}
	mycelium := suite.NewGateway(nil, suite.Config{MCP: map[string]suite.MCPConfig{
		suite.AppMycelium: {URL: "http://mycelium/mcp", Resource: "http://mycelium/mcp"},
	}}, nil)
	h := router(mycelium, 1)
	cases := map[string]struct {
		body string
		want int
		text string
	}{
		"invalid document": {`{"content":"# Bad\n\nINVALID"}`, http.StatusBadRequest, "line 3, column 1: text is empty or holds a forbidden character"},
		"over the cap":     {`{"content":"` + strings.Repeat("a", encre.MaxBytes+1) + `"}`, http.StatusRequestEntityTooLarge, "larger than 1 MiB"},
		"empty content":    {`{"content":"  "}`, http.StatusBadRequest, "content is required"},
		"long title":       {`{"title":"` + strings.Repeat("a", 201) + `","content":"ok"}`, http.StatusBadRequest, "200 characters"},
		"not json":         {`nope`, http.StatusBadRequest, "invalid body"},
	}
	for name, c := range cases {
		rec := put(h, "/artifacts/a?mycelium_space=common", c.body)
		if rec.Code != c.want || !strings.Contains(rec.Body.String(), c.text) {
			t.Errorf("%s: got %d %s, want %d with %q", name, rec.Code, rec.Body.String(), c.want, c.text)
		}
	}
	if rec := put(router(nil, 0), "/artifacts/a", `{"content":"x"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous update = %d", rec.Code)
	}
}

func TestUpdateRefusesNestedFencesAtTheirPlace(t *testing.T) {
	if !encre.Available() {
		t.Skip("encre is not installed")
	}
	mycelium := suite.NewGateway(nil, suite.Config{MCP: map[string]suite.MCPConfig{
		suite.AppMycelium: {URL: "http://mycelium/mcp", Resource: "http://mycelium/mcp"},
	}}, nil)
	body, _ := json.Marshal(map[string]string{"content": encretest.Nested(9)})
	rec := put(router(mycelium, 1), "/artifacts/a?mycelium_space=common", string(body))
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "line 11, column 1: ") || !strings.Contains(rec.Body.String(), "nest") {
		t.Fatalf("nine nested fences must be refused with their place: %d %s", rec.Code, rec.Body.String())
	}
}
