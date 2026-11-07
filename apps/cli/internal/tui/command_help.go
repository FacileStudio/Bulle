package tui

import (
	"strings"

	"github.com/FacileStudio/Bulle/apps/cli/internal/menu"
	"github.com/FacileStudio/Bulle/core/skills"
)

// commandDoc is one command as /help and the slash menu describe it: the name
// the menu inserts, the arguments help shows after it, and what it does.
type commandDoc struct {
	name string
	args string
	desc string
}

// commandDocs is the one list both /help and the slash menu read, so the two
// never disagree. Aliases such as /bg are left out: they still run, but a menu
// listing one command three times is noise.
var commandDocs = []commandDoc{
	{"/clear", "", "start a new session, same client"},
	{"/compact", "", "summarize older turns now to free context (auto-compacts on its own over the threshold)"},
	{"/cost", "", "what this session has spent so far"},
	{"/detach", "", "detach session to background (also /bg, /background)"},
	{"/help", "", "show this message"},
	{"/model", " [name]", "switch to another of the server's models, or open the picker"},
	{"/quit", "", "quit"},
	{"/resume", "", "resume the most recent session for this project"},
	{"/sessions", "", "list all available sessions for this project"},
	{"/status", "", "session summary: questions, answers, tools, cached tokens, context size, elapsed time, log size"},
	{"/skill:name", " [what to do]", "run a loaded skill directly, instead of waiting for the model to decide to"},
	{"/parallel", "", "delegate multiple independent tasks to run concurrently"},
	{"/parallel cancel", " [batch]", "stop a live parallel fan-out (or every one, with no batch)"},
}

// helpLines is the command half of /help.
func helpLines() []string {
	lines := make([]string, 0, len(commandDocs))
	for _, doc := range commandDocs {
		lines = append(lines, doc.name+doc.args+" — "+doc.desc)
	}
	return lines
}

// menuItems is the slash menu: every command a reader can pick, with its help
// description, then every loaded skill. The /skill:name placeholder and the
// /parallel cancel subcommand are help text, not something to insert.
func menuItems(sks map[string]skills.Skill) []menu.Item {
	skillNames := skills.SkillCommandNames(sks)
	items := make([]menu.Item, 0, len(commandDocs)+len(skillNames))
	for _, doc := range commandDocs {
		if strings.ContainsAny(doc.name, ": ") {
			continue
		}
		items = append(items, menu.Item{Value: doc.name, Description: doc.desc})
	}
	for _, name := range skillNames {
		items = append(items, menu.Item{Value: name, Description: sks[strings.TrimPrefix(name, "/skill:")].Description})
	}
	return items
}
