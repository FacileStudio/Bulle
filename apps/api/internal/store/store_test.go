package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/migrations"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/testdb"
	"gorm.io/gorm"
)

var testDB *gorm.DB

var dbConfig = testdb.Config{
	Prefix: "bulle_store_test",
	Migrate: func(db *gorm.DB) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return migrate.Run(context.Background(), migrate.Config{
			DB:     sqlDB,
			FS:     migrations.FS,
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			LockID: 7_300_301,
		})
	},
}

func TestMain(m *testing.M) {
	if url, ok := testdb.URL(); ok {
		db, err := testdb.Open(url, dbConfig)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		testDB = db
	}
	os.Exit(m.Run())
}

func TestMemoryStore(t *testing.T) {
	exercise(t, store.NewMemory(), 1, 2)
	exerciseAttachments(t, store.NewMemory(), 1, 2)
}

func TestPostgresStore(t *testing.T) {
	if testDB == nil {
		t.Skip(testdb.SkipReason("sh scripts/postgres-up.sh"))
	}
	if err := testdb.Truncate(testDB, dbConfig); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{1, 2} {
		err := testDB.Exec("INSERT INTO users (id, email) VALUES (?, ?)", id, fmt.Sprintf("u%d@facile.studio", id)).Error
		if err != nil {
			t.Fatal(err)
		}
	}
	exercise(t, store.NewPostgres(testDB), 1, 2)
	exerciseAttachments(t, store.NewPostgres(testDB), 1, 2)
}

func exercise(t *testing.T, s store.Store, owner, stranger int64) {
	t.Helper()
	ctx := context.Background()

	created, err := s.CreateSession(ctx, store.Session{ID: store.NewID(), UserID: owner, Channel: store.ChannelWeb, Title: "first", Backend: "anthropic", Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("timestamps not set: %+v", created)
	}

	if _, err := s.GetSession(ctx, stranger, created.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger read: %v", err)
	}
	if err := s.DeleteSession(ctx, stranger, created.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger delete: %v", err)
	}
	listed, err := s.ListSessions(ctx, owner, "")
	if err != nil || len(listed) != 1 || listed[0].Title != "first" {
		t.Fatalf("list: %v %+v", err, listed)
	}
	if other, _ := s.ListSessions(ctx, stranger, ""); len(other) != 0 {
		t.Fatalf("stranger sees %d sessions", len(other))
	}
	if err := s.RenameSession(ctx, created.ID, "renamed"); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetSession(ctx, owner, created.ID)
	if err != nil || got.Title != "renamed" {
		t.Fatalf("rename: %v %+v", err, got)
	}

	payloads := []json.RawMessage{json.RawMessage(`{"role":"user"}`), json.RawMessage(`{"role":"assistant"}`)}
	if err := s.ReplaceMessages(ctx, created.ID, payloads); err != nil {
		t.Fatal(err)
	}
	if err := s.ReplaceMessages(ctx, created.ID, append(payloads, json.RawMessage(`{"role":"user","n":3}`))); err != nil {
		t.Fatal(err)
	}
	messages, err := s.Messages(ctx, created.ID)
	if err != nil || len(messages) != 3 || messages[2].Seq != 2 {
		t.Fatalf("messages: %v %+v", err, messages)
	}
	var decoded map[string]any
	if err := json.Unmarshal(messages[2].Payload, &decoded); err != nil || decoded["n"].(float64) != 3 {
		t.Fatalf("payload round trip: %v %s", err, messages[2].Payload)
	}

	run := store.Run{ID: store.NewID(), SessionID: created.ID}
	if err := s.CreateRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateRun(ctx, store.Run{ID: store.NewID(), SessionID: created.ID}); !errors.Is(err, store.ErrSessionBusy) {
		t.Fatalf("second active run: %v", err)
	}
	active, ok, err := s.ActiveRun(ctx, created.ID)
	if err != nil || !ok || active.ID != run.ID {
		t.Fatalf("active run: %v %v %+v", err, ok, active)
	}
	if _, err := s.GetRun(ctx, stranger, run.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger run read: %v", err)
	}

	for seq, kind := range []string{"RUN_STARTED", "TEXT_MESSAGE_CONTENT", "RUN_FINISHED"} {
		event := store.Event{RunID: run.ID, Seq: seq + 1, Type: kind, Payload: json.RawMessage(`{"k":1}`)}
		if err := s.AppendEvent(ctx, event); err != nil {
			t.Fatal(err)
		}
	}
	events, err := s.Events(ctx, run.ID, 1)
	if err != nil || len(events) != 2 || events[0].Seq != 2 || events[1].Type != "RUN_FINISHED" {
		t.Fatalf("events after 1: %v %+v", err, events)
	}
	if last, err := s.LastSeq(ctx, run.ID); err != nil || last != 3 {
		t.Fatalf("last seq: %v %d", err, last)
	}

	if err := s.CreateApproval(ctx, run.ID, "call-1"); err != nil {
		t.Fatal(err)
	}
	if decided, err := s.DecideApproval(ctx, run.ID, "call-1", store.ApprovalApproved, "1"); err != nil || !decided {
		t.Fatalf("first decision: %v %v", err, decided)
	}
	if decided, err := s.DecideApproval(ctx, run.ID, "call-1", store.ApprovalDenied, "1"); err != nil || decided {
		t.Fatalf("second decision must not apply: %v %v", err, decided)
	}

	if err := s.FinishRun(ctx, run.ID, store.RunFinished, "end", json.RawMessage(`{"input_tokens":3}`), 0.0125); err != nil {
		t.Fatal(err)
	}
	finished, err := s.GetRun(ctx, owner, run.ID)
	if err != nil || finished.Status != store.RunFinished || finished.CostUSD != 0.0125 || finished.FinishedAt == nil {
		t.Fatalf("finished run: %v %+v", err, finished)
	}
	if _, ok, _ := s.ActiveRun(ctx, created.ID); ok {
		t.Fatal("finished run still active")
	}
	latest, ok, err := s.LatestRun(ctx, created.ID)
	if err != nil || !ok || latest.ID != run.ID || latest.Status != store.RunFinished {
		t.Fatalf("latest run: %v %v %+v", err, ok, latest)
	}

	dangling := store.Run{ID: store.NewID(), SessionID: created.ID}
	if err := s.CreateRun(ctx, dangling); err != nil {
		t.Fatalf("new run after finish: %v", err)
	}
	interrupted, err := s.InterruptRunning(ctx)
	if err != nil || len(interrupted) != 1 || interrupted[0].ID != dangling.ID || interrupted[0].Status != store.RunInterrupted {
		t.Fatalf("interrupt: %v %+v", err, interrupted)
	}

	if err := s.DeleteSession(ctx, owner, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetSession(ctx, owner, created.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("deleted session: %v", err)
	}
	if remaining, _ := s.Events(ctx, run.ID, 0); len(remaining) != 0 {
		t.Fatalf("events survive session delete: %d", len(remaining))
	}

	if _, err := s.GetAssistantProfile(ctx, owner); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("assistant profile should not exist yet: %v", err)
	}
	profile := store.AssistantProfile{
		UserID:    owner,
		Name:      "Atlas",
		AvatarURL: "https://example.com/avatar.png",
		Archetype: "pragmatic_senior",
		Traits:    []string{"concise", "strict"},
	}
	if err := s.SaveAssistantProfile(ctx, profile); err != nil {
		t.Fatalf("save assistant profile: %v", err)
	}
	gotProfile, err := s.GetAssistantProfile(ctx, owner)
	if err != nil {
		t.Fatalf("get assistant profile: %v", err)
	}
	if gotProfile.Name != "Atlas" || gotProfile.AvatarURL != "https://example.com/avatar.png" || gotProfile.Archetype != "pragmatic_senior" {
		t.Fatalf("profile mismatch: %+v", gotProfile)
	}
	if len(gotProfile.Traits) != 2 || gotProfile.Traits[0] != "concise" || gotProfile.Traits[1] != "strict" {
		t.Fatalf("traits mismatch: %+v", gotProfile.Traits)
	}
	if gotProfile.CreatedAt.IsZero() || gotProfile.UpdatedAt.IsZero() {
		t.Fatalf("timestamps not set: %+v", gotProfile)
	}

	profile.Name = "Nova"
	profile.Traits = []string{"creative"}
	if err := s.SaveAssistantProfile(ctx, profile); err != nil {
		t.Fatalf("update assistant profile: %v", err)
	}
	updatedProfile, err := s.GetAssistantProfile(ctx, owner)
	if err != nil {
		t.Fatalf("get updated profile: %v", err)
	}
	if updatedProfile.Name != "Nova" || len(updatedProfile.Traits) != 1 || updatedProfile.Traits[0] != "creative" {
		t.Fatalf("updated profile mismatch: %+v", updatedProfile)
	}
	if _, err := s.GetAssistantProfile(ctx, stranger); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger assistant profile should not exist: %v", err)
	}

	job, err := s.CreateJob(ctx, store.Job{
		UserID:   owner,
		Name:     "daily-digest",
		Schedule: "0 9 * * *",
		Prompt:   "summarize daily news",
		Profile:  "standard",
		Delivery: "web",
		Enabled:  true,
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if job.ID == "" || job.CreatedAt.IsZero() || job.UpdatedAt.IsZero() {
		t.Fatalf("job metadata not populated: %+v", job)
	}

	gotJob, err := s.GetJob(ctx, owner, job.ID)
	if err != nil || gotJob.Name != "daily-digest" || gotJob.Schedule != "0 9 * * *" {
		t.Fatalf("get job: %v %+v", err, gotJob)
	}
	if _, err := s.GetJob(ctx, stranger, job.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger job read: %v", err)
	}

	jobs, err := s.ListJobs(ctx, owner, "")
	if err != nil || len(jobs) != 1 || jobs[0].ID != job.ID {
		t.Fatalf("list jobs owner: %v %+v", err, jobs)
	}
	if strangerJobs, err := s.ListJobs(ctx, stranger, ""); err != nil || len(strangerJobs) != 0 {
		t.Fatalf("list jobs stranger: %v %+v", err, strangerJobs)
	}

	dueJobs, err := s.ListDueJobs(ctx)
	if err != nil || len(dueJobs) < 1 {
		t.Fatalf("list due jobs: %v %+v", err, dueJobs)
	}

	job.Name = "morning-digest"
	job.Enabled = false
	if err := s.UpdateJob(ctx, job); err != nil {
		t.Fatalf("update job: %v", err)
	}

	strangerJob := job
	strangerJob.UserID = stranger
	if err := s.UpdateJob(ctx, strangerJob); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger update job: %v", err)
	}

	gotJob, err = s.GetJob(ctx, owner, job.ID)
	if err != nil || gotJob.Name != "morning-digest" || gotJob.Enabled {
		t.Fatalf("get updated job: %v %+v", err, gotJob)
	}

	dueAfterDisable, err := s.ListDueJobs(ctx)
	if err != nil {
		t.Fatalf("list due after disable: %v", err)
	}
	for _, dj := range dueAfterDisable {
		if dj.ID == job.ID {
			t.Fatalf("disabled job still in due list: %+v", dj)
		}
	}

	cronSession, err := s.CreateSession(ctx, store.Session{ID: store.NewID(), UserID: owner, Channel: store.ChannelCron, Title: "cron", Backend: "anthropic", Model: "m"})
	if err != nil {
		t.Fatalf("create cron session: %v", err)
	}
	cronRun := store.Run{ID: store.NewID(), SessionID: cronSession.ID}
	if err := s.CreateRun(ctx, cronRun); err != nil {
		t.Fatalf("create cron run: %v", err)
	}
	jobRun := store.JobRun{
		ID:        store.NewID(),
		JobID:     job.ID,
		RunID:     cronRun.ID,
		Status:    store.RunRunning,
		StartedAt: time.Now(),
	}
	if err := s.CreateJobRun(ctx, jobRun); err != nil {
		t.Fatalf("create job run: %v", err)
	}

	jobRuns, err := s.ListJobRuns(ctx, job.ID)
	if err != nil || len(jobRuns) != 1 || jobRuns[0].ID != jobRun.ID || jobRuns[0].Status != store.RunRunning {
		t.Fatalf("list job runs: %v %+v", err, jobRuns)
	}

	activity, err := s.JobActivity(ctx, owner)
	if err != nil || !activity[job.ID].Running || activity[job.ID].SessionID != cronSession.ID {
		t.Fatalf("job activity while running: %v %+v", err, activity)
	}
	if other, _ := s.JobActivity(ctx, stranger); len(other) != 0 {
		t.Fatalf("stranger sees job activity: %+v", other)
	}
	if err := s.FinishRun(ctx, cronRun.ID, store.RunFinished, "", nil, 0); err != nil {
		t.Fatalf("finish cron run: %v", err)
	}
	if activity, _ = s.JobActivity(ctx, owner); activity[job.ID].Running {
		t.Fatalf("job still running after its run finished: %+v", activity)
	}

	now := time.Now()
	if err := s.UpdateJobRunStatus(ctx, jobRun.ID, store.RunFinished, &now); err != nil {
		t.Fatalf("update job run status: %v", err)
	}

	jobRuns, err = s.ListJobRuns(ctx, job.ID)
	if err != nil || len(jobRuns) != 1 || jobRuns[0].Status != store.RunFinished || jobRuns[0].FinishedAt == nil {
		t.Fatalf("list job runs after update: %v %+v", err, jobRuns)
	}

	if err := s.DeleteJob(ctx, stranger, job.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("stranger delete job: %v", err)
	}
	if err := s.DeleteJob(ctx, owner, job.ID); err != nil {
		t.Fatalf("delete job: %v", err)
	}
	if _, err := s.GetJob(ctx, owner, job.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("get deleted job: %v", err)
	}
}
