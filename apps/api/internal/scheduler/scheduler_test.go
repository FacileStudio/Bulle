package scheduler_test

import (
	"context"
	"sync"
	"testing"
	"time"

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
	return m.nextRunID, nil
}

func (m *mockRunner) Defaults() (string, string) {
	return m.backend, m.model
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
