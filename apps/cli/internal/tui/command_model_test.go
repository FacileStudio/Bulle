package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/FacileStudio/nacelle"
)

func serverModel() *Model {
	m := sized()
	m.activeBackend = "openrouter"
	m.activeModel = "a/default"
	m.activeBaseURL = "https://s.example/api/llm/openrouter"
	m.activeAPIKey = "tok"
	m.models = []string{"a/default", "b/other"}
	m.delegate = nacelle.Config{System: "test"}
	return m
}

func TestModelCommandRefusesWhenBusy(t *testing.T) {
	m := &Model{}
	m.run.busy = true

	if cmd := m.modelCmd("b/other"); cmd != nil {
		t.Error("expected nil cmd")
	}
	if len(m.unprinted) == 0 || !strings.Contains(m.unprinted[len(m.unprinted)-1], "cannot switch model while a run is in progress") {
		t.Errorf("expected busy message, got %+v", m.unprinted)
	}
}

func TestModelCommandListsTheServerModels(t *testing.T) {
	m := serverModel()
	if cmd := m.modelCmd(""); cmd != nil {
		t.Errorf("expected nil cmd from open menu, got %v", cmd)
	}
	if !m.menu.Open() || len(m.menu.Items) != 2 || !m.modelPicker {
		t.Fatal("expected the model menu open with the two server models")
	}
	if !m.navigateMenu(tea.KeyPressMsg{Code: tea.KeyEnter}) {
		t.Fatal("expected enter to be handled by model menu")
	}
	if m.menu.Open() || m.modelPicker {
		t.Fatal("expected menu to be dismissed and picker deactivated")
	}
}

func TestModelCommandSwitchesOnTheSameProxy(t *testing.T) {
	m := serverModel()
	m.modelCmd("b/other")
	if m.activeModel != "b/other" || m.activeBackend != "openrouter" || m.activeAPIKey != "tok" {
		t.Errorf("got %s/%s key %q, want openrouter/b/other on the same token (%v)", m.activeBackend, m.activeModel, m.activeAPIKey, m.unprinted)
	}
}

func TestModelCommandRefusesAModelTheServerDoesNotOffer(t *testing.T) {
	m := serverModel()
	m.modelCmd("c/unknown")
	if m.activeModel != "a/default" {
		t.Errorf("switched to %s, want the refusal", m.activeModel)
	}
}

func TestStatusCmdOutputsActiveModel(t *testing.T) {
	m := &Model{}
	m.activeBackend = "anthropic"
	m.activeModel = "claude-opus-5"
	m.statusCmd()
	if len(m.unprinted) == 0 {
		t.Fatal("expected status output, got none")
	}
	last := m.unprinted[len(m.unprinted)-1]
	if !strings.Contains(last, "model · anthropic/claude-opus-5") {
		t.Fatalf("expected model output, got %q", last)
	}
}
