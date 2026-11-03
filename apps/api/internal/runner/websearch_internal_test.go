package runner

import (
	"bytes"
	"context"
	"maps"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/credentials"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/sonar/websearch"
)

type searchProbe struct {
	keys []map[string]string
}

func (p *searchProbe) search(_ context.Context, keys map[string]string, _ string, _ int) ([]websearch.Result, error) {
	p.keys = append(p.keys, maps.Clone(keys))
	return []websearch.Result{{Title: "T", URL: "https://example.com", Snippet: "s"}}, nil
}

func newSearcher(t *testing.T, limit int) (*searcher, *searchProbe, credentials.Store) {
	t.Helper()
	for _, engine := range credentials.SearchEngines() {
		for _, name := range credentials.EnvVars(engine) {
			t.Setenv(name, "")
		}
	}
	creds, err := credentials.NewMemory(bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatal(err)
	}
	probe := &searchProbe{}
	opts := WebOptions{Credentials: creds, Quota: store.NewMemory(), DailyLimit: limit, Search: probe.search, Now: time.Now}
	return &searcher{opts: opts, userID: 1}, probe, creds
}

func TestUserSearchKeyOverridesTheInstanceKey(t *testing.T) {
	s, probe, creds := newSearcher(t, 1)
	t.Setenv("TAVILY_API_KEY", "instance-tavily")
	t.Setenv("EXA_API_KEY", "instance-exa")
	if err := creds.Put(context.Background(), 1, "tavily", "user-tavily"); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := s.run(context.Background(), webSearchInput{Query: "q"}); err != nil {
			t.Fatal(err)
		}
	}
	first, second := probe.keys[0], probe.keys[1]
	if first["tavily"] != "user-tavily" || first["exa"] != "instance-exa" {
		t.Fatalf("first search keys = %v", first)
	}
	if _, ok := second["exa"]; ok || second["tavily"] != "user-tavily" {
		t.Fatalf("second search past the limit kept the instance key: %v", second)
	}
}

func TestOnlyInstanceKeySearchesCount(t *testing.T) {
	s, probe, creds := newSearcher(t, 1)
	if err := creds.Put(context.Background(), 1, "brave", "user-brave"); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		if _, err := s.run(context.Background(), webSearchInput{Query: "q"}); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("EXA_API_KEY", "instance-exa")
	if _, err := s.run(context.Background(), webSearchInput{Query: "q"}); err != nil {
		t.Fatal(err)
	}
	if last := probe.keys[3]; last["exa"] != "instance-exa" {
		t.Fatalf("searches on the user's own key used up the instance quota: %v", last)
	}
	if _, err := s.run(context.Background(), webSearchInput{Query: "q"}); err != nil {
		t.Fatal(err)
	}
	if last := probe.keys[4]; last["exa"] != "" {
		t.Fatalf("instance key used past the daily limit: %v", last)
	}
}
