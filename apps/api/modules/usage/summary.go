package usage

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

const monthDays = 30

type runUsage struct {
	InputTokens         int64 `json:"input_tokens"`
	OutputTokens        int64 `json:"output_tokens"`
	CacheReadTokens     int64 `json:"cache_read_tokens"`
	CacheCreationTokens int64 `json:"cache_creation_tokens"`
}

func (t *Totals) add(run store.UsageRun) {
	var u runUsage
	_ = json.Unmarshal(run.Usage, &u)
	t.Runs++
	t.InputTokens += u.InputTokens
	t.OutputTokens += u.OutputTokens
	t.CacheReadTokens += u.CacheReadTokens
	t.CacheCreationTokens += u.CacheCreationTokens
	t.CostUSD += run.CostUSD
}

func monthStart(now time.Time) time.Time {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return today.AddDate(0, 0, 1-monthDays)
}

func summarize(runs []store.UsageRun, now time.Time, defaultModel string) SummaryResponse {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	week := today.AddDate(0, 0, -6)
	out := SummaryResponse{Timezone: now.Location().String(), Days: make([]Day, monthDays)}
	index := map[string]int{}
	for i := range out.Days {
		date := monthStart(now).AddDate(0, 0, i).Format(time.DateOnly)
		out.Days[i].Date = date
		index[date] = i
	}
	models, channels := map[string]*Totals{}, map[string]*Totals{}
	for _, run := range runs {
		started := run.StartedAt.In(now.Location())
		day, ok := index[started.Format(time.DateOnly)]
		if !ok {
			continue
		}
		out.Days[day].add(run)
		out.Month.add(run)
		if !started.Before(week) {
			out.Week.add(run)
		}
		if !started.Before(today) {
			out.Today.add(run)
		}
		bucket(models, modelName(run.Model, defaultModel)).add(run)
		bucket(channels, channelName(run.Channel)).add(run)
	}
	out.Models, out.Channels = ranked(models), ranked(channels)
	return out
}

func bucket(groups map[string]*Totals, name string) *Totals {
	if groups[name] == nil {
		groups[name] = &Totals{}
	}
	return groups[name]
}

func modelName(model, fallback string) string {
	if model != "" {
		return model
	}
	if fallback != "" {
		return fallback
	}
	return "default"
}

func channelName(channel string) string {
	switch channel {
	case store.ChannelCron:
		return "Background jobs"
	case store.ChannelIdeas:
		return "Ideas"
	}
	return "Chat"
}

func ranked(groups map[string]*Totals) []Group {
	out := make([]Group, 0, len(groups))
	for name, totals := range groups {
		out = append(out, Group{Name: name, Totals: *totals})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CostUSD != out[j].CostUSD {
			return out[i].CostUSD > out[j].CostUSD
		}
		return out[i].Name < out[j].Name
	})
	return out
}
