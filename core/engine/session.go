// Package engine drives a multi-turn agent conversation and streams its events.
package engine

import (
	"github.com/FacileStudio/nacelle"
)

// Session coordinates multi-turn agent execution over one conversation. The
// conversation is the messages the agent's runs commit, in the order nacelle
// commits them, so it keeps reasoning, tool images and provider Native parts.
type Session struct {
	agent    *nacelle.Agent
	messages []nacelle.Message
}

// NewSession constructs a Session with the given agent, starting from history
// once Repair has made it one a provider accepts.
func NewSession(agent *nacelle.Agent, history []nacelle.Message) *Session {
	return &Session{agent: agent, messages: Repair(history)}
}

// Messages returns a copy of the conversation. Read it once the channel a
// Submit returned is closed.
func (s *Session) Messages() []nacelle.Message {
	out := make([]nacelle.Message, len(s.messages))
	for i, m := range s.messages {
		out[i] = nacelle.Message{Role: m.Role, Parts: append([]nacelle.Part(nil), m.Parts...)}
	}
	return out
}

// Agent returns the session's nacelle Agent.
func (s *Session) Agent() *nacelle.Agent {
	return s.agent
}

// keep commits the text a turn streamed before it was cut off, when no
// assistant message holds it, so what the reader saw is what the model
// remembers.
func (s *Session) keep(text string) {
	if text == "" {
		return
	}
	if n := len(s.messages); n > 0 && s.messages[n-1].Role == nacelle.RoleAssistant {
		return
	}
	s.messages = append(s.messages, nacelle.AssistantText(text))
}
