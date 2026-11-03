package spaces

import (
	"strings"
	"unicode/utf8"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

const (
	maxNameRunes = 60
	maxIconRunes = 32
)

// apply folds a request into a space. A space linked to Registre keeps
// Registre's name: renaming it happens in Registre, for every app at once.
func apply(space store.Space, body SpaceRequest) (store.Space, error) {
	if space.RegistreSpaceID != "" && body.Name != nil && strings.TrimSpace(*body.Name) != space.Name {
		return space, troncerrors.Conflict("this space is shared through Registre: rename it there")
	}
	setText(&space.Name, body.Name)
	setText(&space.Icon, body.Icon)
	setText(&space.MyceliumSpaceID, body.MyceliumSpaceID)
	setText(&space.OpusWorkspace, body.OpusWorkspace)
	if body.AgendaCalendar != nil {
		space.AgendaCalendar = *body.AgendaCalendar
	}
	if body.CourrierAccount != nil {
		space.CourrierAccount = *body.CourrierAccount
	}
	return space, validate(space)
}

func setText(field, value *string) {
	if value != nil {
		*field = strings.TrimSpace(*value)
	}
}

func validate(space store.Space) error {
	switch {
	case space.Name == "":
		return troncerrors.Invalid("space name is required")
	case utf8.RuneCountInString(space.Name) > maxNameRunes:
		return troncerrors.Invalid("space name is too long")
	case utf8.RuneCountInString(space.Icon) > maxIconRunes:
		return troncerrors.Invalid("space icon is too long")
	case space.AgendaCalendar < 0 || space.CourrierAccount < 0:
		return troncerrors.Invalid("account and calendar ids cannot be negative")
	}
	return nil
}

func toResponse(space store.Space) SpaceResponse {
	return SpaceResponse{
		ID:              space.ID,
		Name:            space.Name,
		Icon:            space.Icon,
		MyceliumSpaceID: space.MyceliumSpaceID,
		OpusWorkspace:   space.OpusWorkspace,
		AgendaCalendar:  space.AgendaCalendar,
		CourrierAccount: space.CourrierAccount,
		RegistreSpaceID: space.RegistreSpaceID,
		CreatedAt:       space.CreatedAt.UTC(),
		UpdatedAt:       space.UpdatedAt.UTC(),
	}
}
