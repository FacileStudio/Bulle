package store_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestMemoryIdeas(t *testing.T) {
	exerciseIdeas(t, store.NewMemory(), 1, 2)
}

func TestPostgresIdeas(t *testing.T) {
	if testDB == nil {
		t.Skip("no TEST_DATABASE_URL")
	}
	if err := testDB.Exec("TRUNCATE users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{1, 2} {
		err := testDB.Exec("INSERT INTO users (id, email) VALUES (?, ?)", id, fmt.Sprintf("i%d@facile.studio", id)).Error
		if err != nil {
			t.Fatal(err)
		}
	}
	exerciseIdeas(t, store.NewPostgres(testDB), 1, 2)
}

func exerciseIdeas(t *testing.T, s store.Store, owner, stranger int64) {
	t.Helper()
	ctx := context.Background()
	space, err := s.EnsureDefaultSpace(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.EnsureDefaultSpace(ctx, stranger); err != nil {
		t.Fatal(err)
	}
	hidden, err := s.CreateSession(ctx, store.Session{ID: store.NewID(), UserID: owner, SpaceID: space.ID, Channel: store.ChannelIdeas})
	if err != nil {
		t.Fatal(err)
	}
	listed, _ := s.ListSessions(ctx, owner, "")
	for _, session := range listed {
		if session.ID == hidden.ID {
			t.Fatal("an ideas session shows in the session list")
		}
	}
	exerciseIdeaCards(t, s, space.ID, owner, stranger)
	exerciseIdeaBatches(t, s, space.ID, owner, stranger)
	exerciseIdeaSettings(t, s, space.ID, owner, stranger)
	exerciseGoals(t, s, space.ID, owner, stranger)
	exerciseIdeaApps(t, s, owner)
}

func exerciseIdeaApps(t *testing.T, s store.Store, owner int64) {
	t.Helper()
	ctx := context.Background()
	space, err := s.CreateSpace(ctx, store.Space{UserID: owner, Name: "Apps"})
	if err != nil {
		t.Fatal(err)
	}
	all, err := s.CreateIdeaBatch(ctx, store.IdeaBatch{UserID: owner, SpaceID: space.ID, Trigger: store.BatchManual})
	if err != nil {
		t.Fatal(err)
	}
	glouton, err := s.CreateIdeaBatch(ctx, store.IdeaBatch{UserID: owner, SpaceID: space.ID, App: "glouton", Trigger: store.BatchManual})
	if err != nil {
		t.Fatalf("a glouton batch is refused while Bulle's own one runs: %v", err)
	}
	if _, err := s.CreateIdeaBatch(ctx, store.IdeaBatch{UserID: owner, SpaceID: space.ID, App: "glouton", Trigger: store.BatchManual}); !errors.Is(err, store.ErrIdeaBatchRunning) {
		t.Fatalf("a second running glouton batch: %v", err)
	}
	if got, running, _ := s.RunningIdeaBatch(ctx, owner, space.ID, "glouton"); !running || got.ID != glouton.ID || got.App != "glouton" {
		t.Fatalf("running glouton batch = %+v %v", got, running)
	}
	if n, _ := s.CountIdeaBatches(ctx, owner, "", "glouton", store.BatchManual, time.Now().Add(-time.Hour)); n != 1 {
		t.Fatalf("glouton batches across spaces = %d", n)
	}
	mail, _ := s.CreateIdea(ctx, newIdea(owner, space.ID, all.ID, "courrier:5/email:1/reply"))
	lead, _ := s.CreateIdea(ctx, store.Idea{
		UserID: owner, SpaceID: space.ID, BatchID: glouton.ID, App: "glouton", Kind: "follow_up", Title: "Follow up Acme",
		Why: "Replied", Prompt: "Follow up", Fingerprint: "glouton:company:7/follow-up", ExpiresAt: time.Now().Add(time.Hour),
	})
	if mail.App != "courrier" || lead.App != "glouton" {
		t.Fatalf("apps = %q %q", mail.App, lead.App)
	}
	if got, _ := s.ListActiveIdeas(ctx, owner, space.ID, "glouton", time.Now()); len(got) != 1 || got[0].ID != lead.ID {
		t.Fatalf("glouton ideas = %+v", got)
	}
	if got, _ := s.ListActiveIdeas(ctx, owner, space.ID, "", time.Now()); len(got) != 2 {
		t.Fatalf("every app's ideas = %d", len(got))
	}
	if err := s.ExpireIdeas(ctx, owner, space.ID, "", "next"); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.GetIdea(ctx, owner, lead.ID); got.Status != store.IdeaNew {
		t.Fatalf("Bulle's own batch expired a glouton batch's idea: %s", got.Status)
	}
	if got, _ := s.GetIdea(ctx, owner, mail.ID); got.Status != store.IdeaExpired {
		t.Fatalf("Bulle's own batch kept its older idea: %s", got.Status)
	}
	if err := s.ExpireIdeas(ctx, owner, space.ID, "glouton", "next"); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.GetIdea(ctx, owner, lead.ID); got.Status != store.IdeaExpired {
		t.Fatalf("a glouton batch kept the older glouton idea: %s", got.Status)
	}
}

func newIdea(owner int64, spaceID, batchID, fingerprint string) store.Idea {
	return store.Idea{
		UserID: owner, SpaceID: spaceID, BatchID: batchID, Kind: "mail", Title: "Reply to Ann",
		Why: "She asked yesterday", Prompt: "Reply to Ann", Fingerprint: fingerprint,
		Sources:   []store.IdeaSource{{Label: "Email from Ann", Ref: "courrier:5/email:812"}},
		ExpiresAt: time.Now().Add(store.IdeaLifetime),
	}
}

func exerciseIdeaCards(t *testing.T, s store.Store, spaceID string, owner, stranger int64) {
	t.Helper()
	ctx := context.Background()
	if _, err := s.CreateIdea(ctx, newIdea(stranger, spaceID, "b0", "x")); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("a stranger created an idea in the owner's space: %v", err)
	}
	old, err := s.CreateIdea(ctx, newIdea(owner, spaceID, "b1", "fp-old"))
	if err != nil {
		t.Fatal(err)
	}
	saved, _ := s.CreateIdea(ctx, newIdea(owner, spaceID, "b1", "fp-saved"))
	saved.Status = store.IdeaSaved
	if _, err := s.UpdateIdea(ctx, saved); err != nil {
		t.Fatal(err)
	}
	fresh, _ := s.CreateIdea(ctx, newIdea(owner, spaceID, "b2", "fp-new"))
	if got, err := s.GetIdea(ctx, owner, old.ID); err != nil || got.Sources[0].Ref != "courrier:5/email:812" {
		t.Fatalf("get idea: %+v %v", got, err)
	}
	if _, err := s.GetIdea(ctx, stranger, old.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("a stranger read the owner's idea: %v", err)
	}
	hijack := old
	hijack.UserID, hijack.Status = stranger, store.IdeaDismissed
	if _, err := s.UpdateIdea(ctx, hijack); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("a stranger updated the owner's idea: %v", err)
	}
	if err := s.ExpireIdeas(ctx, owner, spaceID, "", "b2"); err != nil {
		t.Fatal(err)
	}
	active, _ := s.ListActiveIdeas(ctx, owner, spaceID, "", time.Now())
	ids := []string{}
	for _, idea := range active {
		ids = append(ids, idea.ID)
	}
	if len(ids) != 2 || !slices.Contains(ids, saved.ID) || !slices.Contains(ids, fresh.ID) {
		t.Fatalf("expiring batch b1 must keep the saved idea and the b2 one: %v", ids)
	}
	if other, _ := s.ListActiveIdeas(ctx, stranger, spaceID, "", time.Now()); len(other) != 0 {
		t.Fatalf("a stranger listed the owner's ideas: %v", other)
	}
	fresh.Status, fresh.Reason = store.IdeaDismissed, "already done"
	if _, err := s.UpdateIdea(ctx, fresh); err != nil {
		t.Fatal(err)
	}
	fps, _ := s.SuppressedFingerprints(ctx, owner, spaceID, time.Now())
	if !slices.Equal(fps, []string{"fp-new", "fp-saved"}) {
		t.Fatalf("suppressed fingerprints = %v", fps)
	}
	feedback, _ := s.IdeaFeedback(ctx, owner, spaceID, "", 30)
	if len(feedback) != 1 || feedback[0].Reason != "already done" {
		t.Fatalf("feedback = %+v", feedback)
	}
	later, _ := s.CreateIdea(ctx, newIdea(owner, spaceID, "b3", "fp-stale"))
	if err := s.ExpireStaleIdeas(ctx, time.Now().Add(store.IdeaLifetime+time.Hour)); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.GetIdea(ctx, owner, later.ID); got.Status != store.IdeaExpired {
		t.Fatalf("an idea past expires_at is %s", got.Status)
	}
}

func exerciseIdeaBatches(t *testing.T, s store.Store, spaceID string, owner, stranger int64) {
	t.Helper()
	ctx := context.Background()
	if _, err := s.CreateIdeaBatch(ctx, store.IdeaBatch{UserID: stranger, SpaceID: spaceID, Trigger: store.BatchManual}); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("a stranger started a batch in the owner's space: %v", err)
	}
	batch, err := s.CreateIdeaBatch(ctx, store.IdeaBatch{UserID: owner, SpaceID: spaceID, Trigger: store.BatchManual})
	if err != nil || batch.Status != store.BatchRunning {
		t.Fatalf("create batch: %+v %v", batch, err)
	}
	if _, err := s.CreateIdeaBatch(ctx, store.IdeaBatch{UserID: owner, SpaceID: spaceID, Trigger: store.BatchDaily}); !errors.Is(err, store.ErrIdeaBatchRunning) {
		t.Fatalf("a second running batch: %v", err)
	}
	if _, running, _ := s.RunningIdeaBatch(ctx, owner, spaceID, ""); !running {
		t.Fatal("the batch is not reported running")
	}
	if _, running, _ := s.RunningIdeaBatch(ctx, stranger, spaceID, ""); running {
		t.Fatal("a stranger sees the owner's running batch")
	}
	now := time.Now().UTC()
	batch.Status, batch.RunID, batch.FinishedAt = store.BatchFinished, "run-1", &now
	if err := s.UpdateIdeaBatch(ctx, batch); err != nil {
		t.Fatal(err)
	}
	if last, ok, _ := s.LastIdeaBatch(ctx, owner, spaceID, ""); !ok || last.ID != batch.ID || last.RunID != "run-1" {
		t.Fatalf("last batch: %+v %v", last, ok)
	}
	if n, _ := s.CountIdeaBatches(ctx, owner, spaceID, "", store.BatchManual, now.Add(-time.Hour)); n != 1 {
		t.Fatalf("manual batches = %d", n)
	}
	stuck, _ := s.CreateIdeaBatch(ctx, store.IdeaBatch{UserID: owner, SpaceID: spaceID, Trigger: store.BatchDaily})
	if err := s.InterruptIdeaBatches(ctx); err != nil {
		t.Fatal(err)
	}
	if _, running, _ := s.RunningIdeaBatch(ctx, owner, spaceID, ""); running {
		t.Fatalf("batch %s still running after the boot sweep", stuck.ID)
	}
}

func exerciseIdeaSettings(t *testing.T, s store.Store, spaceID string, owner, stranger int64) {
	t.Helper()
	ctx := context.Background()
	if _, err := s.GetIdeaSettings(ctx, owner, spaceID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("settings before any save: %v", err)
	}
	theirs := store.DefaultIdeaSettings(stranger, spaceID)
	if _, err := s.SaveIdeaSettings(ctx, theirs); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("a stranger saved settings on the owner's space: %v", err)
	}
	mine := store.DefaultIdeaSettings(owner, spaceID)
	mine.Timezone, mine.Push = "Europe/Paris", true
	if _, err := s.SaveIdeaSettings(ctx, mine); err != nil {
		t.Fatal(err)
	}
	if marked, _ := s.MarkIdeaBatchDate(ctx, owner, spaceID, "2026-11-06"); !marked {
		t.Fatal("the first mark of a day must succeed")
	}
	if marked, _ := s.MarkIdeaBatchDate(ctx, owner, spaceID, "2026-11-06"); marked {
		t.Fatal("the same day was marked twice")
	}
	mine.Steering = "focus on clients"
	if _, err := s.SaveIdeaSettings(ctx, mine); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetIdeaSettings(ctx, owner, spaceID)
	if err != nil || got.Timezone != "Europe/Paris" || got.LastBatchDate != "2026-11-06" || got.Steering != "focus on clients" {
		t.Fatalf("settings: %+v %v", got, err)
	}
	if _, err := s.GetIdeaSettings(ctx, stranger, spaceID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("a stranger read the owner's settings: %v", err)
	}
	if enabled, _ := s.ListEnabledIdeaSettings(ctx); len(enabled) != 1 {
		t.Fatalf("enabled settings = %+v", enabled)
	}
}

func exerciseGoals(t *testing.T, s store.Store, spaceID string, owner, stranger int64) {
	t.Helper()
	ctx := context.Background()
	if _, err := s.CreateGoal(ctx, store.Goal{UserID: stranger, SpaceID: spaceID, Title: "x", Category: "other"}); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("a stranger created a goal in the owner's space: %v", err)
	}
	first, _ := s.CreateGoal(ctx, store.Goal{UserID: owner, SpaceID: spaceID, Title: "Run 10k", Category: "health"})
	second, _ := s.CreateGoal(ctx, store.Goal{UserID: owner, SpaceID: spaceID, Title: "Save", Category: "finance"})
	first.Done = true
	if _, err := s.UpdateGoal(ctx, first); err != nil {
		t.Fatal(err)
	}
	goals, _ := s.ListGoals(ctx, owner, spaceID)
	if len(goals) != 2 || goals[0].ID != second.ID || !goals[1].Done {
		t.Fatalf("open goals must come first: %+v", goals)
	}
	stolen := second
	stolen.UserID = stranger
	if _, err := s.UpdateGoal(ctx, stolen); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("a stranger updated the owner's goal: %v", err)
	}
	if _, err := s.GetGoal(ctx, stranger, second.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("a stranger read the owner's goal: %v", err)
	}
	if got, err := s.GetGoal(ctx, owner, second.ID); err != nil || got.Title != "Save" {
		t.Fatalf("get goal: %+v %v", got, err)
	}
	if err := s.DeleteGoal(ctx, stranger, second.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("a stranger deleted the owner's goal: %v", err)
	}
	if err := s.DeleteGoal(ctx, owner, second.ID); err != nil {
		t.Fatal(err)
	}
}
