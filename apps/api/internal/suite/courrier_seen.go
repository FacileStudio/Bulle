package suite

import (
	"context"
	"encoding/json"
	"time"
)

// markSeen stamps assistant_seen_at on each listed email an earlier call
// already showed the assistant, and records the others as seen now. The read
// flag stays the user's: opening mail on a phone must not hide it from the
// assistant, and the assistant reading mail must not hide it from the user.
// A page it cannot parse or record passes through unchanged.
func (s *toolSet) markSeen(ctx context.Context) func(raw []byte, err error) ([]byte, error) {
	return func(raw []byte, err error) ([]byte, error) {
		if err != nil || s.seeMail == nil {
			return raw, err
		}
		var page map[string]json.RawMessage
		var emails []map[string]any
		if json.Unmarshal(raw, &page) != nil || json.Unmarshal(page["emails"], &emails) != nil {
			return raw, nil
		}
		ids := make([]int64, 0, len(emails))
		for _, email := range emails {
			if id, ok := email["id"].(float64); ok {
				ids = append(ids, int64(id))
			}
		}
		seen, seeErr := s.seeMail(ctx, ids)
		if seeErr != nil || len(seen) == 0 {
			return raw, nil
		}
		for _, email := range emails {
			id, _ := email["id"].(float64)
			if at, ok := seen[int64(id)]; ok {
				email["assistant_seen_at"] = at.UTC().Format(time.RFC3339)
			}
		}
		marked, encodeErr := json.Marshal(emails)
		if encodeErr != nil {
			return raw, nil
		}
		page["emails"] = marked
		return json.Marshal(page)
	}
}
