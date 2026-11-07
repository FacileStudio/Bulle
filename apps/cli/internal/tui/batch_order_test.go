package tui

import (
	"slices"
	"testing"
	"time"
)

// Batches list in the order they started, whatever the map or the key spelling
// would give: detach10 started after detach2.
func TestBatchKeysFollowStartOrder(t *testing.T) {
	m := sized()
	now := time.Now()
	m.parallelTasks = map[string][]parallelTaskInfo{
		"detach10": {{Task: "late", Began: now.Add(2 * time.Second)}},
		"detach2":  {{Task: "early", Began: now}},
		"psa-1":    {{Task: "middle", Began: now.Add(time.Second)}},
	}
	for range 20 {
		if got := m.batchKeys(); !slices.Equal(got, []string{"detach2", "psa-1", "detach10"}) {
			t.Fatalf("batchKeys = %v", got)
		}
	}
}
