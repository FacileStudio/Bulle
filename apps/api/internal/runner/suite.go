package runner

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

func (r *Runner) suiteTools(userID int64, space store.Space) ([]nacelle.Tool, string) {
	note := spacePrompt(space)
	if r.cfg.Suite == nil {
		return nil, note
	}
	defaults := suite.Defaults{
		CourrierAccount: space.CourrierAccount,
		AgendaCalendar:  space.AgendaCalendar,
		OpusWorkspace:   space.OpusWorkspace,
		MyceliumSpace:   space.MyceliumSpaceID,
	}
	tools, suiteNote, err := suite.Tools(r.cfg.Suite, userID, defaults, r.suiteRuntime(userID))
	if err != nil {
		r.cfg.Logger.Error("building suite tools", "error", err)
		return nil, note
	}
	return tools, appendSection(note, suiteNote)
}

func (r *Runner) space(ctx context.Context, userID int64, spaceID string) store.Space {
	if r.cfg.Store == nil || spaceID == "" {
		return store.Space{}
	}
	space, err := r.cfg.Store.GetSpace(ctx, userID, spaceID)
	if err != nil {
		r.cfg.Logger.Warn("loading the session's space", "space", spaceID, "error", err)
		return store.Space{}
	}
	return space
}

func spacePrompt(space store.Space) string {
	if space.Name == "" {
		return ""
	}
	return "You are working in the " + space.Name + " space. Its chats, jobs and default accounts are separate from the user's other spaces."
}
