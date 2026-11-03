package store_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestMemorySpaces(t *testing.T) {
	exerciseSpaces(t, store.NewMemory(), 1, 2)
}

func TestPostgresSpaces(t *testing.T) {
	if testDB == nil {
		t.Skip("no TEST_DATABASE_URL")
	}
	if err := testDB.Exec("TRUNCATE users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{1, 2} {
		err := testDB.Exec("INSERT INTO users (id, email) VALUES (?, ?)", id, fmt.Sprintf("s%d@facile.studio", id)).Error
		if err != nil {
			t.Fatal(err)
		}
	}
	exerciseSpaces(t, store.NewPostgres(testDB), 1, 2)
}

func exerciseSpaces(t *testing.T, s store.Store, owner, stranger int64) {
	t.Helper()
	ctx := context.Background()
	first, err := s.EnsureDefaultSpace(ctx, owner)
	again, _ := s.EnsureDefaultSpace(ctx, owner)
	if err != nil || first.Name != store.DefaultSpaceName || again.ID != first.ID {
		t.Fatalf("the default space must be created once: %v %+v %+v", err, first, again)
	}
	work, err := s.CreateSpace(ctx, store.Space{UserID: owner, Name: "Work", CourrierAccount: 4})
	if err != nil {
		t.Fatal(err)
	}
	checkSpaceOwnership(t, s, work, stranger)
	checkSpaceFilters(t, s, first, work)
	checkSpaceDelete(t, s, first, work)
}

func checkSpaceOwnership(t *testing.T, s store.Store, work store.Space, stranger int64) {
	t.Helper()
	ctx := context.Background()
	if _, err := s.GetSpace(ctx, stranger, work.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger read: %v", err)
	}
	if _, err := s.UpdateSpace(ctx, store.Space{ID: work.ID, UserID: stranger, Name: "x"}); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger update: %v", err)
	}
	if err := s.DeleteSpace(ctx, stranger, work.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger delete: %v", err)
	}
	session := store.Session{ID: store.NewID(), UserID: stranger, SpaceID: work.ID, Channel: store.ChannelWeb}
	if _, err := s.CreateSession(ctx, session); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("a session must not land in another user's space: %v", err)
	}
	job := store.Job{UserID: stranger, SpaceID: work.ID, Name: "n", Schedule: "* * * * *", Prompt: "p"}
	if _, err := s.CreateJob(ctx, job); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("a job must not land in another user's space: %v", err)
	}
	renamed, err := s.UpdateSpace(ctx, store.Space{ID: work.ID, UserID: work.UserID, Name: "Office", OpusWorkspace: "w"})
	if err != nil || renamed.Name != "Office" || renamed.OpusWorkspace != "w" || renamed.CourrierAccount != 0 {
		t.Fatalf("update: %v %+v", err, renamed)
	}
}

func checkSpaceFilters(t *testing.T, s store.Store, home, work store.Space) {
	t.Helper()
	ctx := context.Background()
	owner := home.UserID
	if _, err := s.CreateSession(ctx, store.Session{ID: store.NewID(), UserID: owner, Channel: store.ChannelWeb}); err != nil {
		t.Fatal(err)
	}
	inWork, err := s.CreateSession(ctx, store.Session{ID: store.NewID(), UserID: owner, SpaceID: work.ID, Channel: store.ChannelWeb})
	if err != nil || inWork.SpaceID != work.ID {
		t.Fatalf("create in work: %v %+v", err, inWork)
	}
	job, err := s.CreateJob(ctx, store.Job{UserID: owner, SpaceID: work.ID, Name: "n", Schedule: "* * * * *", Prompt: "p"})
	if err != nil || job.SpaceID != work.ID {
		t.Fatalf("create job in work: %v %+v", err, job)
	}
	homeSessions, _ := s.ListSessions(ctx, owner, home.ID)
	workSessions, _ := s.ListSessions(ctx, owner, work.ID)
	all, _ := s.ListSessions(ctx, owner, "")
	if len(homeSessions) != 1 || len(workSessions) != 1 || workSessions[0].ID != inWork.ID || len(all) != 2 {
		t.Fatalf("session filter: home %d, work %d, all %d", len(homeSessions), len(workSessions), len(all))
	}
	homeJobs, _ := s.ListJobs(ctx, owner, home.ID)
	workJobs, _ := s.ListJobs(ctx, owner, work.ID)
	if len(homeJobs) != 0 || len(workJobs) != 1 {
		t.Fatalf("job filter: home %d, work %d", len(homeJobs), len(workJobs))
	}
}

func checkSpaceDelete(t *testing.T, s store.Store, home, work store.Space) {
	t.Helper()
	ctx := context.Background()
	owner := home.UserID
	if err := s.DeleteSpace(ctx, owner, work.ID); err != nil {
		t.Fatal(err)
	}
	sessions, _ := s.ListSessions(ctx, owner, "")
	jobs, _ := s.ListJobs(ctx, owner, "")
	if len(sessions) != 1 || len(jobs) != 0 {
		t.Fatalf("deleting a space must delete its sessions and jobs: %d sessions, %d jobs", len(sessions), len(jobs))
	}
	if err := s.DeleteSpace(ctx, owner, home.ID); !errors.Is(err, store.ErrLastSpace) {
		t.Fatalf("the last space must stay: %v", err)
	}
	spaces, err := s.ListSpaces(ctx, owner)
	if err != nil || len(spaces) != 1 || spaces[0].ID != home.ID {
		t.Fatalf("list after delete: %v %+v", err, spaces)
	}
}
