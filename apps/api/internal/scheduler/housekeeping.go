package scheduler

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const (
	sweepEvery          = time.Hour
	compactAfter        = time.Hour
	compactBatch        = 500
	orphanAttachmentTTL = 30 * 24 * time.Hour
)

const compactEvents = `
WITH ranked AS (
    SELECT e.run_id, e.seq, e.type, e.payload ->> 'delta' AS delta,
           row_number() OVER (PARTITION BY e.run_id ORDER BY e.seq)
             - row_number() OVER (PARTITION BY e.run_id, e.type ORDER BY e.seq) AS grp
    FROM run_events e WHERE e.run_id IN ?
),
islands AS (
    SELECT run_id, type, grp, min(seq) AS first_seq, string_agg(delta, '' ORDER BY seq) AS text
    FROM ranked WHERE type IN ('TEXT_MESSAGE_CONTENT', 'THINKING')
    GROUP BY run_id, type, grp HAVING count(*) > 1
),
merged AS (
    UPDATE run_events e SET payload = jsonb_build_object('delta', i.text)
    FROM islands i WHERE e.run_id = i.run_id AND e.seq = i.first_seq
    RETURNING e.run_id
)
DELETE FROM run_events e USING ranked r, islands i
WHERE e.run_id = r.run_id AND e.seq = r.seq
  AND r.run_id = i.run_id AND r.type = i.type AND r.grp = i.grp AND r.seq > i.first_seq`

const deleteOrphanAttachments = `
DELETE FROM attachments a WHERE a.created_at < ? AND NOT EXISTS (
    SELECT 1 FROM messages m JOIN sessions s ON s.id = m.session_id
    WHERE s.user_id = a.user_id
      AND m.payload @> jsonb_build_object('parts', jsonb_build_array(jsonb_build_object('type', 'attachment', 'id', a.id)))
)`

// sweep runs the hourly housekeeping: it merges the stream deltas of runs that
// ended over an hour ago into one row per streak, and deletes attachments no
// message has referenced for 30 days.
func (s *Scheduler) sweep(ctx context.Context, now time.Time) {
	if s.cfg.DB == nil || now.Sub(s.swept) < sweepEvery {
		return
	}
	s.swept = now
	db := s.cfg.DB.WithContext(ctx)
	if err := compactRuns(db, now.Add(-compactAfter)); err != nil {
		s.cfg.Logger.Warn("compacting run events", "error", err)
	}
	if err := db.Exec(deleteOrphanAttachments, now.Add(-orphanAttachmentTTL)).Error; err != nil {
		s.cfg.Logger.Warn("deleting orphaned attachments", "error", err)
	}
}

func compactRuns(db *gorm.DB, before time.Time) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var ids []string
		err := tx.Raw(`SELECT id FROM runs WHERE NOT events_compacted AND status <> 'running' AND finished_at < ? LIMIT ?`,
			before, compactBatch).Scan(&ids).Error
		if err != nil || len(ids) == 0 {
			return err
		}
		if err := tx.Exec(compactEvents, ids).Error; err != nil {
			return err
		}
		return tx.Exec(`UPDATE runs SET events_compacted = true WHERE id IN ?`, ids).Error
	})
}
