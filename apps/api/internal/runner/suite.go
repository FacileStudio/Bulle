package runner

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

// noSuiteNote is the memory note of an instance with no suite at all, where
// there is no Mycelium to connect.
const noSuiteNote = "You have no long-term memory and no artifacts in this conversation: nothing you are told is kept after it ends."

// suiteTools are the run's suite tools, mounted beside others more tools, and
// the prompt section that introduces them. The built-in apps' writes wait on
// Bulle's approval card; an MCP app's tools run as its server decides, and a
// confirmation the server asks for becomes the card instead. Bulle's memory
// and artifact tools call Mycelium's MCP server, so Mycelium's confirmation of
// a delete becomes a card the same way.
func (r *Runner) suiteTools(ctx context.Context, lr *liveRun, space store.Space, others int) ([]nacelle.Tool, string) {
	userID := lr.userID
	note := spacePrompt(space)
	if r.cfg.Suite == nil {
		return nil, appendSection(note, noSuiteNote)
	}
	defaults := suite.SpaceDefaults(space)
	rt := r.suiteRuntime(userID)
	rt.CurrentSpace = space.ID
	tools, suiteNote, err := suite.Tools(r.cfg.Suite, userID, defaults, rt)
	if err != nil {
		r.cfg.Logger.Error("building suite tools", "error", err)
		return nil, note
	}
	for i, tool := range tools {
		tools[i] = lr.gate.confirming(tool)
	}
	remote, absent := r.remoteTools(ctx, lr)
	suiteNote = appendSection(suiteNote, mcpNote(names(remote), downApps(absent.apps), defaults))
	suiteNote = appendSection(suiteNote, absentNote(absent.apps))
	mounted, catalog := mountSuite(append(tools, remote...), others, absent)
	return mounted, appendSection(appendSection(note, suiteNote), catalog)
}

// remoteTools are the run's MCP tools, wrapped for the gate, and the apps
// that could not be listed, with how to list the ones that did not answer
// again during the run.
func (r *Runner) remoteTools(ctx context.Context, lr *liveRun) ([]nacelle.Tool, absentApps) {
	prepare := func(remote []nacelle.Tool) []nacelle.Tool {
		remote = capWeb(remote)
		lr.gate.addSources(openWorldSources(remote))
		lr.gate.addWrites(mountedWrites(remote))
		for i, tool := range remote {
			remote[i] = lr.gate.confirming(tool)
		}
		return remote
	}
	listing := r.mcpTools(ctx, lr.userID, nil)
	return prepare(listing.Tools), absentApps{apps: listing.Absent, relist: func(apps []string) suite.Listing {
		again := r.mcpTools(ctx, lr.userID, apps)
		again.Tools = prepare(again.Tools)
		return again
	}}
}

// mcpTools are the user's tools from the suite MCP servers, every one or the
// named apps only. A server that cannot be reached is logged and named in the
// listing's Absent.
func (r *Runner) mcpTools(ctx context.Context, userID int64, apps []string) suite.Listing {
	mcp := r.cfg.Suite.MCP()
	listing := mcp.Tools(ctx, userID)
	if apps != nil {
		listing = mcp.ToolsOf(ctx, userID, apps)
	}
	if listing.Err != nil {
		r.cfg.Logger.Warn("listing suite MCP tools", "error", listing.Err)
	}
	return listing
}

func names(tools []nacelle.Tool) []string {
	out := make([]string, 0, len(tools))
	for _, tool := range tools {
		out = append(out, tool.Spec().Name)
	}
	return out
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
