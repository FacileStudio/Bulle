package catalog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func swapVendor(t *testing.T, url string) {
	t.Helper()
	original, ok := vendors["openrouter"]
	if !ok {
		t.Fatal("no openrouter vendor")
	}
	vendors["openrouter"] = vendor{
		url:  url,
		auth: original.auth,
		parse: func(raw []byte) ([]Model, error) {
			return []Model{{ID: "b/model", Name: "B"}, {ID: "a/model", Name: "A"}}, nil
		},
	}
	t.Cleanup(func() { vendors["openrouter"] = original })
}

func TestCatalogueListsOnceAndSorts(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(srv.Close)
	swapVendor(t, srv.URL)

	c := New(nil, []string{"fallback/model"})
	first := c.List(context.Background(), 7, "openrouter")
	if len(first) != 2 || first[0].ID != "a/model" || first[1].ID != "b/model" {
		t.Fatalf("models = %+v, want the two sorted ids", first)
	}
	c.List(context.Background(), 7, "openrouter")
	if hits.Load() != 1 {
		t.Errorf("vendor hit %d times, want the ten-minute cache to answer once", hits.Load())
	}
	if ids := c.IDs(context.Background(), 7, "openrouter"); len(ids) != 2 || ids[0] != "a/model" {
		t.Errorf("ids = %v", ids)
	}
}

func TestCatalogueFallsBackWhenTheVendorIsUnreachable(t *testing.T) {
	swapVendor(t, "http://127.0.0.1:1/models")

	c := New(nil, []string{"fallback/model"})
	list := c.List(context.Background(), 7, "openrouter")
	if len(list) != 1 || list[0].ID != "fallback/model" {
		t.Fatalf("list = %+v, want the operator's fallback", list)
	}
	if got := c.List(context.Background(), 7, "nope"); len(got) != 1 || got[0].ID != "fallback/model" {
		t.Errorf("unknown backend = %+v, want the fallback", got)
	}
}
