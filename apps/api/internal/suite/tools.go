package suite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/nacelle"
)

const maxToolOutput = 24 << 10

type toolSet struct {
	userID   int64
	defaults Defaults
	tools    []nacelle.Tool
	names    []string
	err      error
}

func (s *toolSet) add(tool nacelle.Tool, err error) {
	if err != nil && s.err == nil {
		s.err = err
	}
	s.tools = append(s.tools, Aliased(tool))
	s.names = append(s.names, tool.Spec().Name)
}

func readTool[In any](set *toolSet, name, description string, run func(ctx context.Context, in In) (string, error)) {
	tool, err := nacelle.NewToolWithOptions(name, description, run, nacelle.ToolOptions{ReadOnly: true})
	set.add(tool, err)
}

// MemorySpace answers the Mycelium space this conversation's memory lives in.
type MemorySpace func(ctx context.Context) (string, error)

// Runtime is what the runner lends the tools for one conversation: the
// user's personal memory space, used when the Bulle space names no Mycelium
// space of its own, the user's Bulle spaces, and the id of the space this
// conversation is in. Any may be empty.
type Runtime struct {
	PersonalSpace MemorySpace
	Spaces        func(ctx context.Context) ([]store.Space, error)
	CurrentSpace  string
}

func Tools(gw *Gateway, userID int64, defaults Defaults, rt Runtime) ([]nacelle.Tool, string, error) {
	set := &toolSet{userID: userID, defaults: defaults}
	cfg := gw.Config()
	for _, app := range cfg.Enabled() {
		switch app {
		case AppMycelium:
			memoryTools(set, NewMycelium(gw), defaults.memorySpace(NewMycelium(gw), userID, rt.PersonalSpace))
			artifactTools(set, NewMycelium(gw), defaults.memorySpace(NewMycelium(gw), userID, NewMycelium(gw).personalArtifacts(userID, rt.PersonalSpace)), rt.CurrentSpace)
		}
	}
	if rt.Spaces != nil {
		spacesTools(set, rt)
	}
	if set.err != nil {
		return nil, "", set.err
	}
	note := Prompt(set.names)
	if cfg.Apps[AppMycelium].Enabled() {
		note += "\n\n" + memoryNote + "\n\n" + artifactNote
	} else {
		note = strings.TrimSpace(note + "\n\n" + NoMemoryNote)
	}
	return set.tools, note, nil
}

func Prompt(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return "You can work with the signed-in user's own Facile data (memory, artifacts, spaces) through " +
		strings.Join(names, ", ") + ". These tools only ever see this user's data. " +
		"When one says the account needs reconnecting, tell the user to sign out of Bulle and sign in again."
}

func jsonResult(value any, err error) (string, error) {
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return clip(string(data)), nil
}

func rawResult(raw []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return clip(string(raw)), nil
}

func clip(text string) string {
	if len(text) <= maxToolOutput {
		return text
	}
	return fmt.Sprintf("%s\n[output truncated at %d bytes, narrow the query]", text[:maxToolOutput], maxToolOutput)
}
