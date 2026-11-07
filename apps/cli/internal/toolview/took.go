package toolview

import (
	"fmt"
	"time"
)

// Took formats how long a call ran: whole milliseconds under a second, one
// decimal of seconds after that, so "2.347s" reads as "2.3s".
func Took(spent time.Duration) string {
	switch {
	case spent < time.Second:
		return max(spent.Round(time.Millisecond), time.Millisecond).String()
	case spent < time.Minute:
		return fmt.Sprintf("%.1fs", spent.Seconds())
	default:
		return spent.Round(100 * time.Millisecond).String()
	}
}
