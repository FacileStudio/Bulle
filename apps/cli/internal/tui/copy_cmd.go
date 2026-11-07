package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/nacelle"
)

// copyCmd puts the last answer on the system clipboard through OSC 52, which
// reaches the reader's clipboard over ssh and tmux too.
func (m *Model) copyCmd() tea.Cmd {
	answer := newestAnswer(m.conversation)
	if answer == "" {
		m.say(fromClient, "nothing to copy yet")
		return nil
	}
	m.say(fromClient, "copied the last answer")
	return tea.SetClipboard(answer)
}

// newestAnswer is the text of the newest assistant message that has any.
func newestAnswer(conversation []nacelle.Message) string {
	for i := len(conversation) - 1; i >= 0; i-- {
		if conversation[i].Role != nacelle.RoleAssistant {
			continue
		}
		if text := strings.TrimSpace(messageText(conversation[i])); text != "" {
			return text
		}
	}
	return ""
}
