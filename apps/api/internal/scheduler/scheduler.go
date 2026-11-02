package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

const advisoryLockID int64 = 7_300_401

type Runner interface {
	Start(ctx context.Context, userID int64, sessionID, text string) (string, error)
	Defaults() (backend, model string)
	Subscribe(runID string) (<-chan struct{}, func())
}

type Config struct {
	DB     *gorm.DB
	Store  store.Store
	Runner Runner
	Logger *slog.Logger
	Tick   time.Duration
}

type Scheduler struct {
	cfg    Config
	parser cron.Parser
	mu     sync.Mutex
	stopCh chan struct{}
	doneCh chan struct{}
}

func New(cfg Config) *Scheduler {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Tick <= 0 {
		cfg.Tick = time.Minute
	}
	return &Scheduler{
		cfg:    cfg,
		parser: cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor),
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	go s.loop()
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	select {
	case <-s.stopCh:
		s.mu.Unlock()
		return
	default:
		close(s.stopCh)
	}
	s.mu.Unlock()
	<-s.doneCh
}

func (s *Scheduler) loop() {
	defer close(s.doneCh)
	ticker := time.NewTicker(s.cfg.Tick)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case now := <-ticker.C:
			s.Tick(context.Background(), now)
		}
	}
}

func (s *Scheduler) Tick(ctx context.Context, now time.Time) {
	if s.cfg.DB != nil {
		tx := s.cfg.DB.Begin()
		var acquired bool
		err := tx.Raw("SELECT pg_try_advisory_xact_lock(?)", advisoryLockID).Scan(&acquired).Error
		if err != nil || !acquired {
			tx.Rollback()
			return
		}
		defer tx.Commit()
	}

	jobs, err := s.cfg.Store.ListDueJobs(ctx)
	if err != nil {
		s.cfg.Logger.Error("scheduler listing jobs failed", slog.Any("error", err))
		return
	}

	minute := now.Truncate(time.Minute)
	for _, job := range jobs {
		if !job.Enabled {
			continue
		}
		if !s.isJobDue(job.Schedule, now) {
			continue
		}
		if s.alreadyRanInMinute(ctx, job.ID, minute) {
			continue
		}
		s.triggerJob(ctx, job, now)
	}
}

func (s *Scheduler) isJobDue(expr string, now time.Time) bool {
	sched, err := s.parser.Parse(expr)
	if err != nil {
		s.cfg.Logger.Warn("invalid cron schedule", slog.String("expr", expr), slog.Any("error", err))
		return false
	}
	minute := now.Truncate(time.Minute)
	prev := minute.Add(-time.Second)
	next := sched.Next(prev)
	return next.Equal(minute)
}

func (s *Scheduler) alreadyRanInMinute(ctx context.Context, jobID string, minute time.Time) bool {
	runs, err := s.cfg.Store.ListJobRuns(ctx, jobID)
	if err != nil || len(runs) == 0 {
		return false
	}
	return !runs[0].StartedAt.Before(minute)
}

func (s *Scheduler) triggerJob(ctx context.Context, job store.Job, now time.Time) {
	backend, model := "", ""
	if s.cfg.Runner != nil {
		backend, model = s.cfg.Runner.Defaults()
	}

	session, err := s.cfg.Store.CreateSession(ctx, store.Session{
		ID:      store.NewID(),
		UserID:  job.UserID,
		Channel: store.ChannelCron,
		Title:   job.Name,
		Backend: backend,
		Model:   model,
	})
	if err != nil {
		s.cfg.Logger.Error("scheduler create session failed", slog.String("job_id", job.ID), slog.Any("error", err))
		return
	}

	runID, err := s.cfg.Runner.Start(ctx, job.UserID, session.ID, job.Prompt)
	if err != nil {
		s.cfg.Logger.Error("scheduler trigger runner failed", slog.String("job_id", job.ID), slog.Any("error", err))
		return
	}

	startedAt := now
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}

	jobRun := store.JobRun{
		ID:        store.NewID(),
		JobID:     job.ID,
		RunID:     runID,
		Status:    store.RunRunning,
		StartedAt: startedAt,
	}
	if err := s.cfg.Store.CreateJobRun(ctx, jobRun); err != nil {
		s.cfg.Logger.Error("scheduler create job run failed", slog.String("job_id", job.ID), slog.Any("error", err))
		return
	}

	go s.trackRun(jobRun.ID, runID, job.UserID)
}

func (s *Scheduler) trackRun(jobRunID, runID string, userID int64) {
	ctx := context.Background()
	notifyCh, unsubscribe := s.cfg.Runner.Subscribe(runID)
	defer unsubscribe()

	checkDone := func() bool {
		run, err := s.cfg.Store.GetRun(ctx, userID, runID)
		if err == nil && run.Status != store.RunRunning {
			_ = s.cfg.Store.UpdateJobRunStatus(ctx, jobRunID, run.Status, run.FinishedAt)
			return true
		}
		return false
	}

	if checkDone() {
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-notifyCh:
			if checkDone() {
				return
			}
		case <-ticker.C:
			if checkDone() {
				return
			}
		}
	}
}
