package suite

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// Defaults are the accounts a Bulle space picks for the suite tools. A tool
// call that names its own account, calendar or workspace keeps it; one that
// omits it gets the space's default, and a zero default leaves the app's own.
//
// SharedSpace is the name of the Registre shared space the Bulle space
// mirrors. Every suite app holds that space under the same name, so an unset
// Opus workspace or Mycelium space follows it without anyone typing an id.
type Defaults struct {
	CourrierAccount int64
	AgendaCalendar  int64
	OpusWorkspace   string
	MyceliumSpace   string
	SharedSpace     string
}

// SpaceDefaults are the defaults a Bulle space gives the suite tools.
func SpaceDefaults(space store.Space) Defaults {
	defaults := Defaults{
		CourrierAccount: space.CourrierAccount,
		AgendaCalendar:  space.AgendaCalendar,
		OpusWorkspace:   space.OpusWorkspace,
		MyceliumSpace:   space.MyceliumSpaceID,
	}
	if space.RegistreSpaceID != "" {
		defaults.SharedSpace = space.Name
	}
	return defaults
}

// SpaceOf answers the Mycelium space a Bulle space's memory and artifacts
// live in, resolved exactly as the agent's memory tools resolve it.
func (m Mycelium) SpaceOf(ctx context.Context, userID int64, space store.Space, saved store.MyceliumSpaces) (string, error) {
	personal := func(ctx context.Context) (string, error) { return m.PersonalSpace(ctx, userID, saved) }
	return SpaceDefaults(space).memorySpace(m, userID, personal).resolve(ctx)
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
	switch {
	case requested != "":
		return requested
	case d.OpusWorkspace != "":
		return d.OpusWorkspace
	}
	return d.SharedSpace
}

func (d Defaults) memorySpace(m Mycelium, userID int64, personal MemorySpace) MemorySpace {
	switch {
	case d.MyceliumSpace != "":
		return func(context.Context) (string, error) { return d.MyceliumSpace, nil }
	case d.SharedSpace != "":
		return func(ctx context.Context) (string, error) { return m.SpaceNamed(ctx, userID, d.SharedSpace) }
	}
	return personal
}
