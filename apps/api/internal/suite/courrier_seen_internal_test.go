package suite

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestMarkSeenStampsOnlyEmailsSeenBefore(t *testing.T) {
	earlier := time.Date(2026, 11, 6, 9, 0, 0, 0, time.UTC)
	var asked []int64
	set := &toolSet{seeMail: func(_ context.Context, ids []int64) (map[int64]time.Time, error) {
		asked = ids
		return map[int64]time.Time{7: earlier}, nil
	}}
	raw, err := set.markSeen(context.Background())([]byte(`{"emails":[{"id":7},{"id":8}],"accounts":[]}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	var page struct {
		Emails   []map[string]any `json:"emails"`
		Accounts []any            `json:"accounts"`
	}
	if err := json.Unmarshal(raw, &page); err != nil {
		t.Fatal(err)
	}
	if len(asked) != 2 || page.Accounts == nil {
		t.Fatalf("asked %v, page %s", asked, raw)
	}
	if page.Emails[0]["assistant_seen_at"] != "2026-11-06T09:00:00Z" || page.Emails[1]["assistant_seen_at"] != nil {
		t.Fatalf("stamps wrong: %s", raw)
	}
}
