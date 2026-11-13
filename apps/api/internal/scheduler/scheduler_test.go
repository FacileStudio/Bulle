package scheduler_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/FacileStudio/Bulle/apps/api/internal/scheduler"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

type mockRunner struct {
	mu        sync.Mutex
	started   []runRequest
	notifyCh  chan struct{}
	backend   string
	model     string
	nextRunID string
	onStart   func(sessionID string)
	stopped   []string
	accessErr error
	startErr  error
}

type runRequest struct {
	userID    int64
	sessionID string
	text      string
}

func newMockRunner() *mockRunner {
	return &mockRunner{
		notifyCh:  make(chan struct{}, 10),
		backend:   "anthropic",
		model:     "claude-3-5-sonnet",
		nextRunID: "mock-run-1",
	}
}

func (m *mockRunner) Start(_ context.Context, userID int64, sessionID, text string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.started = append(m.started, runRequest{userID: userID, sessionID: sessionID, text: text})
	if m.onStart != nil {
		m.onStart(sessionID)
	}
	if m.startErr != nil {
		return "", m.startErr
	}
	return m.nextRunID, nil
}

func (m *mockRunner) CheckAccess(context.Context, int64) error {
	return m.accessErr
}

func (m *mockRunner) Defaults() (string, string) {
	return m.backend, m.model
}

func (m *mockRunner) Stop(_ context.Context, _ int64, runID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopped = append(m.stopped, runID)
	return nil
}

func (m *mockRunner) stoppedRuns() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.stopped...)
}

func (m *mockRunner) Subscribe(_ string) (<-chan struct{}, func()) {
	return m.notifyCh, func() {}
}

func TestSchedulerTick(t *testing.T) {
	st := store.NewMemory()
	runner := newMockRunner()
	sched := scheduler.New(scheduler.Config{
		Store:  st,
		Runner: runner,
	})

	now := time.Date(2026, 11, 2, 10, 15, 0, 0, time.UTC)

	job1, err := st.CreateJob(context.Background(), store.Job{
		UserID:   1,
		Name:     "every-five",
		Schedule: "*/5 * * * *",
		Prompt:   "five min check",
		Enabled:  true,
	})
	if err != nil {
		t.Fatal(err)
	}

	job2, err := st.CreateJob(context.Background(), store.Job{
		UserID:   1,
		Name:     "disabled-job",
		Schedule: "*/5 * * * *",
		Prompt:   "disabled prompt",
		Enabled:  false,
	})
	if err != nil {
		t.Fatal(err)
	}

	job3, err := st.CreateJob(context.Background(), store.Job{
		UserID:   1,
		Name:     "hourly-job",
		Schedule: "@hourly",
		Prompt:   "hourly prompt",
		Enabled:  true,
	})
	if err != nil {
		t.Fatal(err)
	}

	sched.Tick(context.Background(), now)

	runner.mu.Lock()
	startedCount := len(runner.started)
	runner.mu.Unlock()

	if startedCount != 1 {
		t.Fatalf("expected 1 started run, got %d", startedCount)
	}

	runs, err := st.ListJobRuns(context.Background(), job1.ID)
	if err != nil || len(runs) != 1 {
		t.Fatalf("expected 1 job run for job1, got %v (len %d)", err, len(runs))
	}
	if runs[0].Status != store.RunRunning {
		t.Fatalf("expected run status running, got %s", runs[0].Status)
	}

	disabledRuns, err := st.ListJobRuns(context.Background(), job2.ID)
	if err != nil || len(disabledRuns) != 0 {
		t.Fatalf("expected 0 job runs for disabled job, got %d", len(disabledRuns))
	}

	hourlyRuns, err := st.ListJobRuns(context.Background(), job3.ID)
	if err != nil || len(hourlyRuns) != 0 {
		t.Fatalf("expected 0 job runs for hourly job at 10:15, got %d", len(hourlyRuns))
	}

	sched.Tick(context.Background(), now)
	runner.mu.Lock()
	startedCount = len(runner.started)
	runner.mu.Unlock()
	if startedCount != 1 {
		t.Fatalf("expected idempotent tick not to double run, got %d", startedCount)
	}

	hourlyTime := time.Date(2026, 11, 2, 11, 0, 0, 0, time.UTC)
	sched.Tick(context.Background(), hourlyTime)

	runner.mu.Lock()
	startedCount = len(runner.started)
	runner.mu.Unlock()

	if startedCount != 3 {
		t.Fatalf("expected 3 started runs (job1 + job3), got %d", startedCount)
	}
}

func TestSchedulerLifecycle(t *testing.T) {
	st := store.NewMemory()
	runner := newMockRunner()
	sched := scheduler.New(scheduler.Config{
		Store:  st,
		Runner: runner,
		Tick:   50 * time.Millisecond,
	})

	sched.Start()
	sched.Stop()
}

type recordedPush struct {
	mu    sync.Mutex
	seen  []push.Notification
	users []int64
}

func (r *recordedPush) Notify(_ context.Context, userID int64, n push.Notification) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seen = append(r.seen, n)
	r.users = append(r.users, userID)
}

func (r *recordedPush) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.seen)
}

func TestPushDeliveryNotifiesWhenTheRunEnds(t *testing.T) {
	cases := []struct {
		name     string
		delivery string
		status   string
		body     string
	}{
		{"success", "push", store.RunFinished, "digest is ready"},
		{"failure", "push", store.RunError, "digest failed"},
		{"no delivery", "", store.RunFinished, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st := store.NewMemory()
			runner := newMockRunner()
			sent := &recordedPush{}
			sched := scheduler.New(scheduler.Config{Store: st, Runner: runner, Push: sent})
			ctx := context.Background()

			if _, err := st.CreateJob(ctx, store.Job{UserID: 3, Name: "digest", Schedule: "* * * * *", Prompt: "p", Delivery: c.delivery, Enabled: true}); err != nil {
				t.Fatal(err)
			}
			runner.onStart = func(sessionID string) {
				_ = st.CreateRun(ctx, store.Run{ID: runner.nextRunID, SessionID: sessionID, Status: store.RunRunning, StartedAt: time.Now()})
			}
			sched.Tick(ctx, time.Date(2026, 11, 2, 10, 15, 0, 0, time.UTC))
			if sent.count() != 0 {
				t.Fatal("notified before the run ended")
			}
			if err := st.FinishRun(ctx, runner.nextRunID, c.status, "", nil, 0); err != nil {
				t.Fatal(err)
			}
			runner.notifyCh <- struct{}{}

			if c.body == "" {
				time.Sleep(150 * time.Millisecond)
				if sent.count() != 0 {
					t.Fatal("notified without push delivery")
				}
				return
			}
			deadline := time.Now().Add(3 * time.Second)
			for sent.count() == 0 && time.Now().Before(deadline) {
				time.Sleep(10 * time.Millisecond)
			}
			sent.mu.Lock()
			defer sent.mu.Unlock()
			if len(sent.seen) != 1 || sent.seen[0].Body != c.body || sent.seen[0].Title != "Bulle" || sent.users[0] != 3 {
				t.Fatalf("got %+v for users %v", sent.seen, sent.users)
			}
			if !strings.HasPrefix(sent.seen[0].Path, "sessions/") || len(sent.seen[0].Path) == len("sessions/") {
				t.Fatalf("path %q", sent.seen[0].Path)
			}
		})
	}
}

func TestScheduledRunIsStoppedWhenItsJobIsDeletedWhileStarting(t *testing.T) {
	st := store.NewMemory()
	runner := newMockRunner()
	sched := scheduler.New(scheduler.Config{Store: st, Runner: runner})
	ctx := context.Background()

	job, err := st.CreateJob(ctx, store.Job{UserID: 4, Name: "racy", Schedule: "* * * * *", Prompt: "p", Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	runner.onStart = func(string) {
		if err := st.DeleteJob(ctx, 4, job.ID); err != nil {
			t.Error(err)
		}
	}
	sched.Tick(ctx, time.Date(2026, 11, 2, 10, 15, 0, 0, time.UTC))

	if got := runner.stoppedRuns(); len(got) != 1 || got[0] != runner.nextRunID {
		t.Fatalf("stopped %v, want [%s]", got, runner.nextRunID)
	}
}

func TestScheduledRunIsStoppedWhenItsJobIsDeletedWhileRunning(t *testing.T) {
	st := store.NewMemory()
	runner := newMockRunner()
	sched := scheduler.New(scheduler.Config{Store: st, Runner: runner})
	ctx := context.Background()

	job, err := st.CreateJob(ctx, store.Job{UserID: 5, Name: "long", Schedule: "* * * * *", Prompt: "p", Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	runner.onStart = func(sessionID string) {
		_ = st.CreateRun(ctx, store.Run{ID: runner.nextRunID, SessionID: sessionID, Status: store.RunRunning, StartedAt: time.Now()})
	}
	sched.Tick(ctx, time.Date(2026, 11, 2, 10, 15, 0, 0, time.UTC))
	if got := runner.stoppedRuns(); len(got) != 0 {
		t.Fatalf("stopped %v before the job was deleted", got)
	}

	if err := st.DeleteJob(ctx, 5, job.ID); err != nil {
		t.Fatal(err)
	}
	runner.notifyCh <- struct{}{}

	deadline := time.Now().Add(3 * time.Second)
	for len(runner.stoppedRuns()) == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if got := runner.stoppedRuns(); len(got) != 1 || got[0] != runner.nextRunID {
		t.Fatalf("stopped %v, want [%s]", got, runner.nextRunID)
	}
}

func TestScheduledRunThatCannotStartLeavesNoConversation(t *testing.T) {
	cases := map[string]func(*mockRunner){
		"no ai access":  func(m *mockRunner) { m.accessErr = errors.New("no ai.access") },
		"start refused": func(m *mockRunner) { m.startErr = errors.New("too many runs") },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			st := store.NewMemory()
			rn := newMockRunner()
			setup(rn)
			sched := scheduler.New(scheduler.Config{Store: st, Runner: rn})
			job, err := st.CreateJob(ctx, store.Job{UserID: 1, Name: "minutely", Schedule: "* * * * *", Prompt: "check", Enabled: true})
			if err != nil {
				t.Fatal(err)
			}

			start := time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)
			for i := range 3 {
				sched.Tick(ctx, start.Add(time.Duration(i)*time.Minute))
			}

			sessions, err := st.ListSessions(ctx, 1, "")
			if err != nil {
				t.Fatal(err)
			}
			if len(sessions) != 0 {
				t.Fatalf("expected no conversation after failed starts, got %d", len(sessions))
			}
			runs, err := st.ListJobRuns(ctx, job.ID)
			if err != nil || len(runs) != 0 {
				t.Fatalf("expected no job run, got %d (%v)", len(runs), err)
			}
		})
	}
}
