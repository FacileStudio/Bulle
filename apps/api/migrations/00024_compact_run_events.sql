-- +goose Up
ALTER TABLE runs ADD COLUMN events_compacted BOOLEAN NOT NULL DEFAULT false;

WITH due AS (
    SELECT id FROM runs WHERE status <> 'running'
),
ranked AS (
    SELECT e.run_id, e.seq, e.type, e.payload ->> 'delta' AS delta,
           row_number() OVER (PARTITION BY e.run_id ORDER BY e.seq)
             - row_number() OVER (PARTITION BY e.run_id, e.type ORDER BY e.seq) AS grp
    FROM run_events e JOIN due d ON d.id = e.run_id
),
islands AS (
    SELECT run_id, type, grp, min(seq) AS first_seq, string_agg(delta, '' ORDER BY seq) AS text
    FROM ranked
    WHERE type IN ('TEXT_MESSAGE_CONTENT', 'THINKING')
    GROUP BY run_id, type, grp
    HAVING count(*) > 1
),
merged AS (
    UPDATE run_events e SET payload = jsonb_build_object('delta', i.text)
    FROM islands i WHERE e.run_id = i.run_id AND e.seq = i.first_seq
    RETURNING e.run_id
)
DELETE FROM run_events e
USING ranked r, islands i
WHERE e.run_id = r.run_id AND e.seq = r.seq
  AND r.run_id = i.run_id AND r.type = i.type AND r.grp = i.grp AND r.seq > i.first_seq;

UPDATE runs SET events_compacted = true WHERE status <> 'running';

CREATE INDEX runs_events_uncompacted_idx ON runs (finished_at) WHERE NOT events_compacted;

-- +goose Down
DROP INDEX IF EXISTS runs_events_uncompacted_idx;
ALTER TABLE runs DROP COLUMN IF EXISTS events_compacted;
