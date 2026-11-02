package assistant

import (
	"context"
	"slices"
	"strconv"
	"time"

	"github.com/FacileStudio/Bulle/core/settings"
)

// CalendarSource reads today's events through the agenda CLI.
type CalendarSource struct {
	Suite    Suite
	Settings settings.Assistant
	Now      func() time.Time
}

// Fetch returns the events starting between today and tomorrow, cancelled ones
// excluded.
func (c *CalendarSource) Fetch(ctx context.Context) ([]Item, error) {
	cals, err := c.Suite.Calendars(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now
	if c.Now != nil {
		now = c.Now
	}
	loc, err := time.LoadLocation(c.Settings.WithDefaults().Timezone)
	if err != nil {
		loc = time.UTC
	}
	from := now().In(loc)
	var items []Item
	for _, cal := range cals {
		if !pickedID(c.Settings.Sources.Calendars, cal.ID) {
			continue
		}
		got, err := c.events(ctx, cal, from)
		if err != nil {
			return nil, err
		}
		items = append(items, got...)
	}
	return items, nil
}

func (c *CalendarSource) events(ctx context.Context, cal Calendar, from time.Time) ([]Item, error) {
	evs, err := c.Suite.Events(ctx, cal.ID, from, from.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}
	var items []Item
	for _, e := range evs {
		if e.Status == "cancelled" {
			continue
		}
		items = append(items, Item{
			Source:  SourceCalendar,
			ID:      strconv.Itoa(e.ID),
			Title:   e.Title,
			Account: cal.Name,
			When:    e.StartAt,
			Status:  e.Status,
		})
	}
	return items, nil
}

func pickedID(ids []int, id int) bool {
	return len(ids) == 0 || slices.Contains(ids, id)
}
