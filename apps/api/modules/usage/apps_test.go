package usage

import (
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestSummarizeCountsAppCallsAsTheirOwnChannel(t *testing.T) {
	now := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	runs := []store.UsageRun{
		run("2026-11-03T08:00:00Z", store.ChannelWeb, "", 100, 0.5),
		run("2026-11-03T09:00:00Z", store.ChannelApps, "mimo", 40, 0.25),
	}
	got := summarize(runs, now, "haiku")
	if got.Today.CostUSD != 0.75 || got.Today.InputTokens != 140 {
		t.Errorf("today %+v must include the apps call", got.Today)
	}
	if len(got.Channels) != 2 || got.Channels[1].Name != "Apps" || got.Channels[1].CostUSD != 0.25 {
		t.Errorf("channels %+v", got.Channels)
	}
	if got.Models[1].Name != "mimo" {
		t.Errorf("models %+v", got.Models)
	}
}
