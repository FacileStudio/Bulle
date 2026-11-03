package usage

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func run(at string, channel, model string, input int64, cost float64) store.UsageRun {
	started, _ := time.Parse(time.RFC3339, at)
	usage, _ := json.Marshal(map[string]int64{"input_tokens": input, "output_tokens": 1})
	return store.UsageRun{StartedAt: started, Channel: channel, Model: model, Usage: usage, CostUSD: cost}
}

func TestSummarizeBucketsByTheUsersDay(t *testing.T) {
	paris, _ := time.LoadLocation("Europe/Paris")
	now := time.Date(2026, 11, 3, 10, 0, 0, 0, paris)
	runs := []store.UsageRun{
		run("2026-11-02T23:30:00Z", "web", "", 100, 0.5),
		run("2026-11-02T22:30:00Z", "cron", "m2", 10, 0.25),
		run("2026-10-28T12:00:00Z", "web", "", 1, 1),
		run("2026-09-01T12:00:00Z", "web", "", 1, 9),
	}
	got := summarize(runs, now, "haiku")
	if got.Today.Runs != 1 || got.Today.InputTokens != 100 {
		t.Errorf("23:30 UTC is 00:30 in Paris, so only it is today: %+v", got.Today)
	}
	if got.Week.Runs != 3 || got.Month.Runs != 3 || got.Month.CostUSD != 1.75 {
		t.Errorf("week %+v month %+v; runs older than 30 days must be ignored", got.Week, got.Month)
	}
	if len(got.Days) != 30 || got.Days[29].Date != "2026-11-03" || got.Days[28].Runs != 1 {
		t.Errorf("days must end today and place 23:30 Paris on Nov 2: %+v", got.Days[27:])
	}
	if got.Models[0].Name != "haiku" || got.Models[0].Runs != 2 || got.Channels[1].Name != "Background jobs" {
		t.Errorf("models %+v channels %+v", got.Models, got.Channels)
	}
}
