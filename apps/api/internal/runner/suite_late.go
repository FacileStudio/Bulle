package runner

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/nacelle/mcp/client"
)

// relistEvery is the shortest pause between two tries at listing the apps
// that did not answer when the run started.
const relistEvery = 3 * time.Second

// absentApps are the suite apps a run could not list, each with why, and how
// to list the ones that did not answer again.
type absentApps struct {
	apps   map[string]*suite.AppError
	relist func(apps []string) suite.Listing
}

func downApps(apps map[string]*suite.AppError) []string {
	return suite.Listing{Absent: apps}.Down()
}

// suiteCatalog is nacelle's catalog over the run's suite tools, with Bulle's
// own search, and apps that did not answer tried again on the next use, so a
// Registre or app redeploy costs the run seconds rather than its tools.
type suiteCatalog struct {
	specs [3]nacelle.Spec

	mu     sync.Mutex
	all    []nacelle.Tool
	inner  []nacelle.Tool
	absent map[string]*suite.AppError
	relist func(apps []string) suite.Listing
	tried  time.Time
}

func newSuiteCatalog(tools []nacelle.Tool, absent absentApps) *suiteCatalog {
	all := slices.Clone(tools)
	inner := client.Catalog(all, 0)
	c := &suiteCatalog{all: all, inner: inner, absent: maps.Clone(absent.apps), relist: absent.relist}
	if c.absent == nil {
		c.absent = map[string]*suite.AppError{}
	}
	for i, tool := range inner {
		c.specs[i] = tool.Spec()
	}
	return c
}

func (c *suiteCatalog) tools() []nacelle.Tool {
	return []nacelle.Tool{catalogSearch{c}, catalogDetails{c}, catalogRoute{c}}
}

// refresh lists again the apps that did not answer, at most every
// relistEvery, and adds the tools of those that do now.
func (c *suiteCatalog) refresh() {
	c.mu.Lock()
	defer c.mu.Unlock()
	down := downApps(c.absent)
	if len(down) == 0 || c.relist == nil || time.Since(c.tried) < relistEvery {
		return
	}
	listing := c.relist(down)
	c.tried = time.Now()
	for _, app := range down {
		delete(c.absent, app)
		if absent, still := listing.Absent[app]; still {
			c.absent[app] = absent
		}
	}
	if len(listing.Tools) > 0 {
		c.all = append(c.all, listing.Tools...)
		c.inner = client.Catalog(c.all, 0)
	}
}

func (c *suiteCatalog) current(i int) nacelle.Tool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.inner[i]
}

// search answers search_tools: the tools that share the most words with the
// query, an app's own tools first when the query names the app in English or
// French, and a line for each named app that is out of reach and why.
func (c *suiteCatalog) search(query string) string {
	c.mu.Lock()
	tools, absent := slices.Clone(c.all), maps.Clone(c.absent)
	c.mu.Unlock()
	words := suite.Words(query)
	named := map[string]bool{}
	for _, word := range words {
		for _, app := range suite.AppsNamed(word) {
			named[app] = true
		}
	}
	var body strings.Builder
	for _, spec := range ranked(tools, words, named) {
		fmt.Fprintf(&body, "%s — %s\n", spec.Name, spec.Description)
	}
	for _, app := range slices.Sorted(maps.Keys(absent)) {
		if len(words) == 0 || named[app] {
			fmt.Fprintln(&body, absentLine(absent[app]))
		}
	}
	if body.Len() == 0 {
		return fmt.Sprintf("No tool matches %q. Fewer or broader words find more.", query)
	}
	return body.String()
}

type hit struct {
	spec  nacelle.Spec
	score int
}

// ranked is the tools matching words, best first: one point per word of three
// letters or more found in the name or description, two when the query names
// the tool's app. Hits under half the best score are dropped.
func ranked(tools []nacelle.Tool, words []string, named map[string]bool) []nacelle.Spec {
	var hits []hit
	best := 0
	for _, tool := range tools {
		spec := tool.Spec()
		score := matches(spec, words, named)
		if score == 0 {
			continue
		}
		hits = append(hits, hit{spec: spec, score: score})
		best = max(best, score)
	}
	slices.SortFunc(hits, func(a, b hit) int {
		return cmp.Or(cmp.Compare(b.score, a.score), cmp.Compare(a.spec.Name, b.spec.Name))
	})
	var specs []nacelle.Spec
	for _, h := range hits {
		if h.score*2 >= best {
			specs = append(specs, h.spec)
		}
	}
	return specs
}

func matches(spec nacelle.Spec, words []string, named map[string]bool) int {
	if len(words) == 0 {
		return 1
	}
	text := " " + strings.Join(suite.Words(spec.Name+" "+spec.Description), " ")
	score := 0
	for _, word := range words {
		if len(word) >= 3 && strings.Contains(text, " "+word) {
			score++
		}
	}
	if app, _, _ := strings.Cut(spec.Name, "_"); named[app] {
		score += 2
	}
	return score
}

type catalogSearch struct{ c *suiteCatalog }

func (t catalogSearch) Spec() nacelle.Spec { return t.c.specs[0] }

func (t catalogSearch) Run(_ context.Context, call nacelle.Call, _ func(string)) (nacelle.Result, error) {
	var in struct {
		Query string `json:"query"`
	}
	if len(call.Input) > 0 {
		if err := json.Unmarshal(call.Input, &in); err != nil {
			return nacelle.Result{}, fmt.Errorf("reading query: %w", err)
		}
	}
	t.c.refresh()
	return nacelle.Result{Text: t.c.search(in.Query)}, nil
}

type catalogDetails struct{ c *suiteCatalog }

func (t catalogDetails) Spec() nacelle.Spec { return t.c.specs[1] }

func (t catalogDetails) Run(ctx context.Context, call nacelle.Call, emit func(string)) (nacelle.Result, error) {
	t.c.refresh()
	return t.c.current(1).Run(ctx, call, emit)
}

// catalogRoute is call_tool: a nacelle.Router, so the loop runs the named
// tool itself and hooks, approvals and taint see that tool.
type catalogRoute struct{ c *suiteCatalog }

func (t catalogRoute) Spec() nacelle.Spec { return t.c.specs[2] }

func (t catalogRoute) Route(input json.RawMessage) (nacelle.Tool, json.RawMessage, error) {
	t.c.refresh()
	router, ok := t.c.current(2).(nacelle.Router)
	if !ok {
		return nil, nil, errors.New("call_tool cannot route")
	}
	return router.Route(input)
}

func (t catalogRoute) Run(ctx context.Context, call nacelle.Call, emit func(string)) (nacelle.Result, error) {
	t.c.refresh()
	return t.c.current(2).Run(ctx, call, emit)
}
