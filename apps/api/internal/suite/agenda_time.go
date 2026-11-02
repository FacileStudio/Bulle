package suite

import (
	"encoding/json"
	"fmt"
	"time"
)

func zone(requested *string, stored map[string]any) (*time.Location, error) {
	name, _ := stored["timezone"].(string)
	if requested != nil {
		name = *requested
	}
	if name == "" {
		return time.UTC, nil
	}
	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("%q is not an IANA timezone", name)
	}
	return location, nil
}

func storedTime(stored map[string]any, key string) time.Time {
	text, _ := stored[key].(string)
	parsed, _ := time.Parse(time.RFC3339, text)
	return parsed
}

func keptDuration(stored map[string]any, allDay *bool) time.Duration {
	start, end := storedTime(stored, "start_at"), storedTime(stored, "end_at")
	if !start.IsZero() && end.After(start) {
		return end.Sub(start)
	}
	isAllDay, _ := stored["is_all_day"].(bool)
	if allDay != nil {
		isAllDay = *allDay
	}
	if isAllDay {
		return 24 * time.Hour
	}
	return time.Hour
}

func toMap(value any) (map[string]any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	return out, json.Unmarshal(data, &out)
}
