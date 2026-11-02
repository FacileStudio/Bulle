package suite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FacileStudio/nacelle"
)

const maxToolOutput = 24 << 10

type toolSet struct {
	userID int64
	tools  []nacelle.Tool
	names  []string
	err    error
}

func (s *toolSet) add(tool nacelle.Tool, err error) {
	if err != nil && s.err == nil {
		s.err = err
	}
	s.tools = append(s.tools, tool)
	s.names = append(s.names, tool.Name())
}

func readTool[In any](set *toolSet, name, description string, run func(ctx context.Context, in In) (string, error)) {
	tool, err := nacelle.NewToolWithOptions(name, description, run, nacelle.ToolOptions{ReadOnly: true})
	set.add(tool, err)
}

func Tools(gw *Gateway, userID int64) ([]nacelle.Tool, string, error) {
	set := &toolSet{userID: userID}
	cfg := gw.Config()
	for _, app := range cfg.Enabled() {
		switch app {
		case AppCourrier:
			courrierTools(set, NewCourrier(gw))
		case AppAgenda:
			agendaTools(set, NewAgenda(gw))
		case AppOpus:
			opusTools(set, NewOpus(gw))
		case AppSablier:
			sablierTools(set, NewSablier(gw))
		}
	}
	if set.err != nil {
		return nil, "", set.err
	}
	return set.tools, Prompt(set.names), nil
}

func Prompt(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return "You can read the signed-in user's own Facile data with " + strings.Join(names, ", ") +
		". These tools are read-only and only ever see this user's data. " +
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
