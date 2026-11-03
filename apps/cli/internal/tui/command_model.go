package tui

import (
	"fmt"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/menu"
	"github.com/FacileStudio/Bulle/core/provider"
	"github.com/FacileStudio/nacelle"
)

func (m *Model) modelCmd(arg string) tea.Cmd {
	trimmed := strings.TrimSpace(arg)
	if trimmed == "" {
		return m.openModelMenu()
	}
	return m.switchModel(trimmed)
}

// openModelMenu lists the models the server offers, the only ones a session
// may switch to.
func (m *Model) openModelMenu() tea.Cmd {
	if len(m.models) == 0 {
		m.say(fromClient, "the server offers no other model")
		return nil
	}
	items := make([]menu.Item, 0, len(m.models))
	for _, name := range m.models {
		description := "server model"
		if name == m.activeModel {
			description = "active"
		}
		items = append(items, menu.Item{Value: "/model " + name, Description: description})
	}
	m.menu = *menu.New(items)
	m.menu.Filtered = items
	m.menu.Selected = 0
	m.menu.Dismissed = false
	m.modelPicker = true
	return nil
}

// switchModel moves the session to another of the server's models, on the
// same backend, proxy and token.
func (m *Model) switchModel(target string) tea.Cmd {
	if m.run.busy {
		m.say(fromClient, "cannot switch model while a run is in progress")
		return nil
	}
	if !slices.Contains(m.models, target) {
		m.say(fromClient, fmt.Sprintf("%s is not offered by the server; /model lists what is", target))
		return nil
	}
	backend, err := provider.New(provider.Config{
		Backend: m.activeBackend,
		Model:   target,
		BaseURL: m.activeBaseURL,
		APIKey:  m.activeAPIKey,
	})
	if err != nil {
		m.say(fromClient, "failed to switch model: "+err.Error())
		return nil
	}
	m.delegate.Backend = nacelle.Retry(backend, nacelle.RetryOptions{})
	agent, err := nacelle.New(m.delegate)
	if err != nil {
		m.say(fromClient, "failed to create agent: "+err.Error())
		return nil
	}
	m.agent = agent
	m.activate(m.activeBackend, target, m.activeBaseURL, m.activeAPIKey)
	m.say(fromClient, "switched to "+target)
	return nil
}
