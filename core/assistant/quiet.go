package assistant

import (
	"encoding/json"
	"time"

	"github.com/FacileStudio/Bulle/core/settings"
)

type quietFile struct {
	Until time.Time `json:"until"`
}

// InQuiet reports whether now falls in the quiet window or before the
// /quiet-until timestamp stored in the "quiet" state.
func InQuiet(now time.Time, a settings.Assistant, st State) bool {
	if quietUntil(st).After(now) {
		return true
	}
	a = a.WithDefaults()
	loc, err := time.LoadLocation(a.Timezone)
	if err != nil {
		loc = time.UTC
	}
	local := now.In(loc)
	cur := local.Hour()*60 + local.Minute()
	start, ok1 := parseClock(a.QuietStart)
	end, ok2 := parseClock(a.QuietEnd)
	if !ok1 || !ok2 || start == end {
		return false
	}
	if start < end {
		return cur >= start && cur < end
	}
	return cur >= start || cur < end
}

func quietUntil(st State) time.Time {
	data, err := st.Load("quiet")
	if err != nil || data == nil {
		return time.Time{}
	}
	var q quietFile
	if json.Unmarshal(data, &q) != nil {
		return time.Time{}
	}
	return q.Until
}

func parseClock(s string) (int, bool) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return 0, false
	}
	return t.Hour()*60 + t.Minute(), true
}

// SetQuiet silences interrupts until the given time in the "quiet" state.
func SetQuiet(st State, until time.Time) error {
	return st.Update("quiet", func([]byte) ([]byte, error) {
		return json.Marshal(quietFile{Until: until})
	})
}
