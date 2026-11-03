package suite

import "context"

// Defaults are the accounts a Bulle space picks for the suite tools. A tool
// call that names its own account, calendar or workspace keeps it; one that
// omits it gets the space's default, and a zero default leaves the app's own.
type Defaults struct {
	CourrierAccount int64
	AgendaCalendar  int64
	OpusWorkspace   string
	MyceliumSpace   string
}

func (d Defaults) account(requested int64) int64 {
	if requested != 0 {
		return requested
	}
	return d.CourrierAccount
}

func (d Defaults) calendar(requested int64) int64 {
	if requested != 0 {
		return requested
	}
	return d.AgendaCalendar
}

func (d Defaults) workspace(requested string) string {
	if requested != "" {
		return requested
	}
	return d.OpusWorkspace
}

func (d Defaults) memorySpace(personal MemorySpace) MemorySpace {
	if d.MyceliumSpace == "" {
		return personal
	}
	return func(context.Context) (string, error) { return d.MyceliumSpace, nil }
}
