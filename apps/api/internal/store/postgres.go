package store

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type Postgres struct {
	db *gorm.DB
}

func NewPostgres(db *gorm.DB) *Postgres { return &Postgres{db: db} }

type sessionRow struct {
	ID        string
	UserID    int64
	SpaceID   string
	Channel   string
	Title     string
	Backend   string
	Model     string
	CreatedAt time.Time
	UpdatedAt time.Time

	MemorySeenSeq int
}

func (r sessionRow) session() Session { return Session(r) }

const sessionColumns = "id, user_id, space_id, channel, title, backend, model, created_at, updated_at, memory_seen_seq"

func (p *Postgres) CreateSession(ctx context.Context, session Session) (Session, error) {
	spaceID, err := p.spaceFor(ctx, session.UserID, session.SpaceID)
	if err != nil {
		return Session{}, err
	}
	var rows []sessionRow
	err = p.db.WithContext(ctx).Raw(
		`INSERT INTO sessions (id, user_id, space_id, channel, title, backend, model)
		SELECT ?, ?, id, ?, ?, ?, ? FROM spaces WHERE id = ? AND user_id = ? RETURNING `+sessionColumns,
		session.ID, session.UserID, session.Channel, session.Title, session.Backend, session.Model, spaceID, session.UserID,
	).Scan(&rows).Error
	if err != nil {
		return Session{}, err
	}
	if len(rows) == 0 {
		return Session{}, ErrNotFound
	}
	return rows[0].session(), nil
}

func (p *Postgres) ListSessions(ctx context.Context, userID int64, spaceID string) ([]Session, error) {
	var rows []sessionRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT "+sessionColumns+" FROM sessions WHERE user_id = ? AND channel NOT IN (?, ?) AND (? = '' OR space_id = ?) ORDER BY updated_at DESC, id",
		userID, ChannelIdeas, ChannelMemory, spaceID, spaceID,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Session, len(rows))
	for i, row := range rows {
		out[i] = row.session()
	}
	return out, nil
}

func (p *Postgres) GetSession(ctx context.Context, userID int64, id string) (Session, error) {
	var rows []sessionRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT "+sessionColumns+" FROM sessions WHERE id = ? AND user_id = ?", id, userID,
	).Scan(&rows).Error
	if err != nil {
		return Session{}, err
	}
	if len(rows) == 0 {
		return Session{}, ErrNotFound
	}
	return rows[0].session(), nil
}

func (p *Postgres) RenameSession(ctx context.Context, id, title string) error {
	return p.db.WithContext(ctx).Exec("UPDATE sessions SET title = ?, updated_at = now() WHERE id = ?", title, id).Error
}

func (p *Postgres) DeleteSession(ctx context.Context, userID int64, id string) error {
	result := p.db.WithContext(ctx).Exec("DELETE FROM sessions WHERE id = ? AND user_id = ?", id, userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

type messageRow struct {
	Seq     int
	Payload []byte
}

func (p *Postgres) Messages(ctx context.Context, sessionID string) ([]Message, error) {
	var rows []messageRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT seq, payload FROM messages WHERE session_id = ? ORDER BY seq", sessionID,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Message, len(rows))
	for i, row := range rows {
		out[i] = Message{Seq: row.Seq, Payload: json.RawMessage(row.Payload)}
	}
	return out, nil
}

func (p *Postgres) ReplaceMessages(ctx context.Context, sessionID string, payloads []json.RawMessage) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM messages WHERE session_id = ?", sessionID).Error; err != nil {
			return err
		}
		for seq, payload := range payloads {
			if err := tx.Exec(
				"INSERT INTO messages (session_id, seq, payload) VALUES (?, ?, ?::jsonb)",
				sessionID, seq, string(payload),
			).Error; err != nil {
				return err
			}
		}
		return tx.Exec("UPDATE sessions SET updated_at = now() WHERE id = ?", sessionID).Error
	})
}

type runRow struct {
	ID         string
	SessionID  string
	Status     string
	StopReason string
	Usage      []byte
	CostUSD    float64
	StartedAt  time.Time
	FinishedAt *time.Time
}

func (r runRow) run() Run {
	return Run{
		ID: r.ID, SessionID: r.SessionID, Status: r.Status, StopReason: r.StopReason,
		Usage: json.RawMessage(r.Usage), CostUSD: r.CostUSD, StartedAt: r.StartedAt, FinishedAt: r.FinishedAt,
	}
}

const runColumns = "r.id, r.session_id, r.status, r.stop_reason, r.usage, r.cost_usd::float8 AS cost_usd, r.started_at, r.finished_at"

func (p *Postgres) CreateRun(ctx context.Context, run Run) error {
	err := p.db.WithContext(ctx).Exec(
		"INSERT INTO runs (id, session_id, status) VALUES (?, ?, ?)", run.ID, run.SessionID, RunRunning,
	).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrSessionBusy
	}
	return err
}

func (p *Postgres) FinishRun(ctx context.Context, id, status, stop string, usage json.RawMessage, costUSD float64) error {
	return p.db.WithContext(ctx).Exec(
		"UPDATE runs SET status = ?, stop_reason = ?, usage = ?::jsonb, cost_usd = ?, finished_at = now() WHERE id = ?",
		status, stop, jsonOr(usage, "{}"), costUSD, id,
	).Error
}

func (p *Postgres) GetRun(ctx context.Context, userID int64, id string) (Run, error) {
	var rows []runRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT "+runColumns+" FROM runs r JOIN sessions s ON s.id = r.session_id WHERE r.id = ? AND s.user_id = ?", id, userID,
	).Scan(&rows).Error
	if err != nil {
		return Run{}, err
	}
	if len(rows) == 0 {
		return Run{}, ErrNotFound
	}
	return rows[0].run(), nil
}

func (p *Postgres) ActiveRun(ctx context.Context, sessionID string) (Run, bool, error) {
	var rows []runRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT "+runColumns+" FROM runs r WHERE r.session_id = ? AND r.status = 'running'", sessionID,
	).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return Run{}, false, err
	}
	return rows[0].run(), true, nil
}

func (p *Postgres) LatestRun(ctx context.Context, sessionID string) (Run, bool, error) {
	var rows []runRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT "+runColumns+" FROM runs r WHERE r.session_id = ? ORDER BY r.started_at DESC LIMIT 1", sessionID,
	).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return Run{}, false, err
	}
	return rows[0].run(), true, nil
}

func (p *Postgres) InterruptRunning(ctx context.Context) ([]Run, error) {
	var rows []runRow
	err := p.db.WithContext(ctx).Raw(
		"UPDATE runs r SET status = 'interrupted', stop_reason = 'interrupted', finished_at = now() WHERE r.status = 'running' RETURNING " + runColumns,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Run, len(rows))
	for i, row := range rows {
		out[i] = row.run()
	}
	return out, nil
}

type eventRow struct {
	RunID   string
	Seq     int
	Type    string
	Payload []byte
}

func (p *Postgres) AppendEvent(ctx context.Context, event Event) error {
	return p.db.WithContext(ctx).Exec(
		"INSERT INTO run_events (run_id, seq, type, payload) VALUES (?, ?, ?, ?::jsonb)",
		event.RunID, event.Seq, event.Type, jsonOr(event.Payload, "{}"),
	).Error
}

func (p *Postgres) Events(ctx context.Context, runID string, after int) ([]Event, error) {
	var rows []eventRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT run_id, seq, type, payload FROM run_events WHERE run_id = ? AND seq > ? ORDER BY seq", runID, after,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Event, len(rows))
	for i, row := range rows {
		out[i] = Event{RunID: row.RunID, Seq: row.Seq, Type: row.Type, Payload: json.RawMessage(row.Payload)}
	}
	return out, nil
}

func (p *Postgres) LastSeq(ctx context.Context, runID string) (int, error) {
	var seq int
	err := p.db.WithContext(ctx).Raw("SELECT COALESCE(MAX(seq), 0) FROM run_events WHERE run_id = ?", runID).Scan(&seq).Error
	return seq, err
}

func (p *Postgres) CreateApproval(ctx context.Context, runID, callID string) error {
	return p.db.WithContext(ctx).Exec(
		"INSERT INTO approvals (run_id, tool_call_id) VALUES (?, ?) ON CONFLICT DO NOTHING", runID, callID,
	).Error
}

func (p *Postgres) DecideApproval(ctx context.Context, runID, callID, status, decidedBy string) (bool, error) {
	result := p.db.WithContext(ctx).Exec(
		"UPDATE approvals SET status = ?, decided_by = ?, decided_at = now() WHERE run_id = ? AND tool_call_id = ? AND status = 'pending'",
		status, decidedBy, runID, callID,
	)
	return result.RowsAffected > 0, result.Error
}

type assistantProfileRow struct {
	UserID    int64
	Name      string
	AvatarURL string
	Archetype string
	Traits    []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (p *Postgres) GetAssistantProfile(ctx context.Context, userID int64) (AssistantProfile, error) {
	var rows []assistantProfileRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT user_id, name, avatar_url, archetype, traits, created_at, updated_at FROM assistant_profiles WHERE user_id = ?",
		userID,
	).Scan(&rows).Error
	if err != nil {
		return AssistantProfile{}, err
	}
	if len(rows) == 0 {
		return AssistantProfile{}, ErrNotFound
	}
	row := rows[0]
	traits := []string{}
	if len(row.Traits) > 0 {
		_ = json.Unmarshal(row.Traits, &traits)
	}
	if traits == nil {
		traits = []string{}
	}
	return AssistantProfile{
		UserID:    row.UserID,
		Name:      row.Name,
		AvatarURL: row.AvatarURL,
		Archetype: row.Archetype,
		Traits:    traits,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (p *Postgres) SaveAssistantProfile(ctx context.Context, profile AssistantProfile) error {
	traitsJSON, err := json.Marshal(profile.Traits)
	if err != nil {
		return err
	}
	if profile.Traits == nil {
		traitsJSON = []byte("[]")
	}
	return p.db.WithContext(ctx).Exec(`
		INSERT INTO assistant_profiles (user_id, name, avatar_url, archetype, traits, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?::jsonb, now(), now())
		ON CONFLICT (user_id) DO UPDATE SET
			name = EXCLUDED.name,
			avatar_url = EXCLUDED.avatar_url,
			archetype = EXCLUDED.archetype,
			traits = EXCLUDED.traits,
			updated_at = now()
	`, profile.UserID, profile.Name, profile.AvatarURL, profile.Archetype, string(traitsJSON)).Error
}

type jobRow struct {
	ID        string
	UserID    string
	SpaceID   string
	Name      string
	Schedule  string
	Prompt    string
	Profile   *string
	Delivery  *string
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r jobRow) job() Job {
	var uid int64
	if id, err := strconv.ParseInt(r.UserID, 10, 64); err == nil {
		uid = id
	}
	var profile, delivery string
	if r.Profile != nil {
		profile = *r.Profile
	}
	if r.Delivery != nil {
		delivery = *r.Delivery
	}
	return Job{
		ID:        r.ID,
		UserID:    uid,
		SpaceID:   r.SpaceID,
		Name:      r.Name,
		Schedule:  r.Schedule,
		Prompt:    r.Prompt,
		Profile:   profile,
		Delivery:  delivery,
		Enabled:   r.Enabled,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

const jobColumns = "id, user_id, space_id, name, schedule, prompt, profile, delivery, enabled, created_at, updated_at"

func (p *Postgres) CreateJob(ctx context.Context, job Job) (Job, error) {
	if job.ID == "" {
		job.ID = NewID()
	}
	spaceID, err := p.spaceFor(ctx, job.UserID, job.SpaceID)
	if err != nil {
		return Job{}, err
	}
	var rows []jobRow
	err = p.db.WithContext(ctx).Raw(
		`INSERT INTO jobs (id, user_id, space_id, name, schedule, prompt, profile, delivery, enabled)
		SELECT ?, ?, id, ?, ?, ?, ?, ?, ? FROM spaces WHERE id = ? AND user_id = ? RETURNING `+jobColumns,
		job.ID, strconv.FormatInt(job.UserID, 10), job.Name, job.Schedule, job.Prompt, job.Profile, job.Delivery, job.Enabled,
		spaceID, job.UserID,
	).Scan(&rows).Error
	if err != nil {
		return Job{}, err
	}
	if len(rows) == 0 {
		return Job{}, ErrNotFound
	}
	return rows[0].job(), nil
}

func (p *Postgres) GetJob(ctx context.Context, userID int64, id string) (Job, error) {
	uidStr := strconv.FormatInt(userID, 10)
	var rows []jobRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT "+jobColumns+" FROM jobs WHERE id = ? AND user_id = ?", id, uidStr,
	).Scan(&rows).Error
	if err != nil {
		return Job{}, err
	}
	if len(rows) == 0 {
		return Job{}, ErrNotFound
	}
	return rows[0].job(), nil
}

func (p *Postgres) ListJobs(ctx context.Context, userID int64, spaceID string) ([]Job, error) {
	uidStr := strconv.FormatInt(userID, 10)
	var rows []jobRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT "+jobColumns+" FROM jobs WHERE user_id = ? AND (? = '' OR space_id = ?) ORDER BY created_at DESC, id",
		uidStr, spaceID, spaceID,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Job, len(rows))
	for i, row := range rows {
		out[i] = row.job()
	}
	return out, nil
}

func (p *Postgres) UpdateJob(ctx context.Context, job Job) error {
	uidStr := strconv.FormatInt(job.UserID, 10)
	result := p.db.WithContext(ctx).Exec(
		"UPDATE jobs SET name = ?, schedule = ?, prompt = ?, profile = ?, delivery = ?, enabled = ?, updated_at = now() WHERE id = ? AND user_id = ?",
		job.Name, job.Schedule, job.Prompt, job.Profile, job.Delivery, job.Enabled, job.ID, uidStr,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) DeleteJob(ctx context.Context, userID int64, id string) error {
	uidStr := strconv.FormatInt(userID, 10)
	result := p.db.WithContext(ctx).Exec("DELETE FROM jobs WHERE id = ? AND user_id = ?", id, uidStr)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) ListDueJobs(ctx context.Context) ([]Job, error) {
	var rows []jobRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT " + jobColumns + " FROM jobs WHERE enabled = true ORDER BY created_at ASC, id",
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]Job, len(rows))
	for i, row := range rows {
		out[i] = row.job()
	}
	return out, nil
}

type jobRunRow struct {
	ID         string
	JobID      string
	RunID      string
	Status     string
	StartedAt  time.Time
	FinishedAt *time.Time
}

func (r jobRunRow) jobRun() JobRun {
	return JobRun{
		ID:         r.ID,
		JobID:      r.JobID,
		RunID:      r.RunID,
		Status:     r.Status,
		StartedAt:  r.StartedAt,
		FinishedAt: r.FinishedAt,
	}
}

const jobRunColumns = "id, job_id, run_id, status, started_at, finished_at"

func (p *Postgres) CreateJobRun(ctx context.Context, run JobRun) error {
	if run.ID == "" {
		run.ID = NewID()
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now()
	}
	return p.db.WithContext(ctx).Exec(
		"INSERT INTO job_runs (id, job_id, run_id, status, started_at, finished_at) VALUES (?, ?, ?, ?, ?, ?)",
		run.ID, run.JobID, run.RunID, run.Status, run.StartedAt, run.FinishedAt,
	).Error
}

func (p *Postgres) UpdateJobRunStatus(ctx context.Context, id, status string, finishedAt *time.Time) error {
	result := p.db.WithContext(ctx).Exec(
		"UPDATE job_runs SET status = ?, finished_at = ? WHERE id = ? OR run_id = ?",
		status, finishedAt, id, id,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) ListJobRuns(ctx context.Context, jobID string) ([]JobRun, error) {
	var rows []jobRunRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT "+jobRunColumns+" FROM job_runs WHERE job_id = ? ORDER BY started_at DESC, id", jobID,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]JobRun, len(rows))
	for i, row := range rows {
		out[i] = row.jobRun()
	}
	return out, nil
}
