package suite

import (
	"context"
	"errors"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

type spaceListInput struct{}

type spaceListEntry struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Icon            string `json:"icon"`
	Shared          bool   `json:"shared"`
	MyceliumSpaceID string `json:"mycelium_space_id"`
	OpusWorkspace   string `json:"opus_workspace"`
	AgendaCalendar  int64  `json:"agenda_calendar"`
	CourrierAccount int64  `json:"courrier_account"`
	Current         bool   `json:"current"`
}

func spacesTools(set *toolSet, rt Runtime) {
	readTool(set, "spaces_list",
		"List the user's Bulle spaces and the suite accounts each one points at. The space of this conversation is the one marked current.",
		func(ctx context.Context, _ spaceListInput) (string, error) {
			if rt.Spaces == nil {
				return "", errors.New("spaces are unavailable in this conversation")
			}
			spaces, err := rt.Spaces(ctx)
			if err != nil {
				return "", err
			}
			return jsonResult(spaceListEntries(spaces, set.defaults), nil)
		})
}

func spaceListEntries(spaces []store.Space, d Defaults) []spaceListEntry {
	entries := make([]spaceListEntry, 0, len(spaces))
	for _, space := range spaces {
		entries = append(entries, spaceListEntry{
			ID:              space.ID,
			Name:            space.Name,
			Icon:            space.Icon,
			Shared:          space.RegistreSpaceID != "",
			MyceliumSpaceID: space.MyceliumSpaceID,
			OpusWorkspace:   space.OpusWorkspace,
			AgendaCalendar:  space.AgendaCalendar,
			CourrierAccount: space.CourrierAccount,
			Current:         spaceListCurrent(space, d),
		})
	}
	return entries
}

func spaceListCurrent(space store.Space, d Defaults) bool {
	if space.MyceliumSpaceID != d.MyceliumSpace || space.OpusWorkspace != d.OpusWorkspace ||
		space.AgendaCalendar != d.AgendaCalendar || space.CourrierAccount != d.CourrierAccount {
		return false
	}
	if d.SharedSpace != "" {
		return space.Name == d.SharedSpace && space.RegistreSpaceID != ""
	}
	return space.RegistreSpaceID == ""
}
