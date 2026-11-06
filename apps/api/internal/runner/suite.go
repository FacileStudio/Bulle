package runner

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

func (r *Runner) suiteTools(lr *liveRun, space store.Space) ([]nacelle.Tool, string) {
	userID := lr.userID
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
	if space.RegistreSpaceID != "" {
		defaults.SharedSpace = space.Name
	}
	rt := r.suiteRuntime(userID)
	rt.CurrentSpace = space.ID
	tools, suiteNote, err := suite.Tools(r.cfg.Suite, userID, defaults, rt)
	if err != nil {
		r.cfg.Logger.Error("building suite tools", "error", err)
		return nil, note
	}
	lr.gate.addPolicy(mailSendTool, mailPolicy{courrier: suite.NewCourrier(r.cfg.Suite), userID: userID, defaults: defaults})
	for _, name := range suite.WriteTools() {
		if name == mailSendTool || name == mailDraftTool {
			continue
		}
		lr.gate.addPolicy(name, suitePolicy{name: name, defaults: defaults, space: space, gate: lr.gate})
	}
	return lr.gate.revisableTools(tools), appendSection(note, suiteNote)
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
	return "You are working in the " + space.Name + " space. " +
		"When a tool needs a mailbox, calendar or workspace and you do not name one, it uses this space's. " +
		"You can read and change anything the user's account reaches, in any of the user's spaces; " +
		"call spaces_list to see them. Jobs and memory belong to the space this conversation is in."
}
