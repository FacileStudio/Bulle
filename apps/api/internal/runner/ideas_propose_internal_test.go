package runner

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func proposeFixture(t *testing.T) (*Runner, *store.Memory, func() *ideaRun) {
	t.Helper()
	st := store.NewMemory()
	space, err := st.EnsureDefaultSpace(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	r := New(Config{Store: st})
	fresh := func() *ideaRun {
		return &ideaRun{batch: store.IdeaBatch{ID: store.NewID(), UserID: 1, SpaceID: space.ID}, space: space}
	}
	return r, st, fresh
}

func validProposal(fingerprint string) proposeInput {
	return proposeInput{
		Kind: "mail", Title: "Reply to Ann", Why: "She asked about the quote yesterday",
		Prompt: "Reply to Ann about the quote", Fingerprint: fingerprint,
		Sources: []store.IdeaSource{{Label: "Email from Ann", Ref: "courrier:5/email:812"}},
	}
}

func TestProposeValidatesKindAndLimits(t *testing.T) {
	r, _, fresh := proposeFixture(t)
	run := fresh()
	bad := map[string]func(*proposeInput){
		"kind":    func(in *proposeInput) { in.Kind = "tweet" },
		"title":   func(in *proposeInput) { in.Title = strings.Repeat("a", store.IdeaTitleLimit+1) },
		"why":     func(in *proposeInput) { in.Why = "" },
		"prompt":  func(in *proposeInput) { in.Prompt = strings.Repeat("a", store.IdeaPromptLimit+1) },
		"draft":   func(in *proposeInput) { in.Draft = strings.Repeat("a", store.IdeaDraftLimit+1) },
		"sources": func(in *proposeInput) { in.Sources = make([]store.IdeaSource, store.IdeaSourcesLimit+1) },
	}
	for name, spoil := range bad {
		in := validProposal("fp-" + name)
		spoil(&in)
		if _, err := r.propose(context.Background(), run, in); err == nil {
			t.Errorf("a bad %s was accepted", name)
		}
	}
	if run.count != 0 {
		t.Fatalf("refused ideas were counted: %d", run.count)
	}
}

func TestProposeRefusesTheEighthIdea(t *testing.T) {
	r, st, fresh := proposeFixture(t)
	run := fresh()
	for i := range store.IdeaBatchLimit {
		if _, err := r.propose(context.Background(), run, validProposal(fmt.Sprintf("fp-%d", i))); err != nil {
			t.Fatalf("idea %d: %v", i+1, err)
		}
	}
	if _, err := r.propose(context.Background(), run, validProposal("fp-extra")); err == nil {
		t.Fatal("the eighth idea was accepted")
	}
	active, _ := st.ListActiveIdeas(context.Background(), 1, run.space.ID, "", r.now())
	if len(active) != store.IdeaBatchLimit {
		t.Fatalf("%d ideas stored", len(active))
	}
}

func TestProposeRefusesActiveAndDismissedFingerprints(t *testing.T) {
	r, st, fresh := proposeFixture(t)
	ctx := context.Background()
	first := fresh()
	if _, err := r.propose(ctx, first, validProposal("opus:task:44/due")); err != nil {
		t.Fatal(err)
	}
	if _, err := r.propose(ctx, first, validProposal("courrier:5/email:812/reply")); err != nil {
		t.Fatal(err)
	}
	active, _ := st.ListActiveIdeas(ctx, 1, first.space.ID, "", r.now())
	for _, idea := range active {
		if idea.Fingerprint == "courrier:5/email:812/reply" {
			idea.Status = store.IdeaDismissed
			if _, err := st.UpdateIdea(ctx, idea); err != nil {
				t.Fatal(err)
			}
		}
	}
	next := fresh()
	for _, fingerprint := range []string{"opus:task:44/due", "courrier:5/email:812/reply"} {
		_, err := r.propose(ctx, next, validProposal(fingerprint))
		if err == nil || !strings.Contains(err.Error(), "already suggested") {
			t.Errorf("fingerprint %s: %v", fingerprint, err)
		}
	}
	if _, err := r.propose(ctx, next, validProposal("agenda:event:9/prep")); err != nil {
		t.Fatalf("a new fingerprint was refused: %v", err)
	}
}
