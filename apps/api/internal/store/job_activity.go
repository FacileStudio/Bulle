package store

import (
	"context"
	"strconv"
)

// JobActivity is a job's latest run: whether it is still going and the
// conversation its output is written to.
type JobActivity struct {
	Running   bool
	SessionID string
}

type jobActivityRow struct {
	JobID     string
	SessionID string
	Status    string
}

// JobActivity reads each job's latest run from the runs table, which the
// startup sweep marks interrupted, so a crash never leaves a job running forever.
func (p *Postgres) JobActivity(ctx context.Context, userID int64) (map[string]JobActivity, error) {
	var rows []jobActivityRow
	err := p.db.WithContext(ctx).Raw(
		`SELECT DISTINCT ON (jr.job_id) jr.job_id, r.session_id, r.status
		FROM job_runs jr JOIN runs r ON r.id = jr.run_id JOIN jobs j ON j.id = jr.job_id
		WHERE j.user_id = ? ORDER BY jr.job_id, jr.started_at DESC`,
		strconv.FormatInt(userID, 10),
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]JobActivity, len(rows))
	for _, row := range rows {
		out[row.JobID] = JobActivity{Running: row.Status == RunRunning, SessionID: row.SessionID}
	}
	return out, nil
}

// JobActivity is the in-memory twin of Postgres.JobActivity.
func (m *Memory) JobActivity(_ context.Context, userID int64) (map[string]JobActivity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	latest := map[string]JobRun{}
	for _, run := range m.jobRuns {
		job, ok := m.jobs[run.JobID]
		if !ok || job.UserID != userID {
			continue
		}
		if prev, seen := latest[run.JobID]; !seen || run.StartedAt.After(prev.StartedAt) {
			latest[run.JobID] = run
		}
	}
	out := make(map[string]JobActivity, len(latest))
	for jobID, jr := range latest {
		run := m.runs[jr.RunID]
		out[jobID] = JobActivity{Running: run.Status == RunRunning, SessionID: run.SessionID}
	}
	return out, nil
}
