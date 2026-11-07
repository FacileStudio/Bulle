package runner

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestJobScheduleNeedsFifteenMinutes(t *testing.T) {
	now := time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)
	cases := map[string]bool{
		"0 9 * * *":      true,
		"*/15 * * * *":   true,
		"@every 20m":     true,
		"*/5 * * * *":    false,
		"0,59 * * * *":   false,
		"* * 1 * *":      false,
		"0,5 0,23 * * 1": false,
	}
	for schedule, ok := range cases {
		if err := checkInterval(schedule, now); (err == nil) != ok {
			t.Errorf("%s: err %v, want ok %v", schedule, err, ok)
		}
	}
}

func TestJobsStopAtTheCap(t *testing.T) {
	ctx := context.Background()
	limited := jobLimits{Store: store.NewMemory()}
	for i := range MaxJobsPerUser {
		job := store.Job{UserID: 1, Name: "j" + strconv.Itoa(i), Schedule: "0 9 * * *", Prompt: "p", Enabled: true}
		if _, err := limited.CreateJob(ctx, job); err != nil {
			t.Fatalf("job %d: %v", i, err)
		}
	}
	if _, err := limited.CreateJob(ctx, store.Job{UserID: 1, Schedule: "0 9 * * *"}); err == nil {
		t.Fatal("job past the cap was created")
	}
	if _, err := limited.CreateJob(ctx, store.Job{UserID: 2, Schedule: "0 9 * * *", Name: "j", Prompt: "p"}); err != nil {
		t.Fatalf("another user is capped too: %v", err)
	}
}
