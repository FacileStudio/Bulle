package engine

import (
	"github.com/FacileStudio/nacelle"
)

// Repair makes a stored conversation one a provider accepts again. A run that
// was stopped mid-tool, or capped by StopIterations, ends on tool calls no
// result answers, and every provider rejects a conversation carrying one: the
// calls are dropped, and an assistant turn left with nothing to say goes with
// them. Two adjacent messages of one role are then merged into one.
func Repair(messages []nacelle.Message) []nacelle.Message {
	return alternate(pruneOrphanedCalls(messages))
}

func toolResultIDs(msg nacelle.Message) map[string]bool {
	if msg.Role != nacelle.RoleUser {
		return nil
	}
	ids := make(map[string]bool)
	for _, part := range msg.Parts {
		if tr, ok := part.(nacelle.ToolResult); ok {
			ids[tr.ID] = true
		}
	}
	return ids
}

// filterAssistantParts keeps the calls a result answers and every other
// part, and reports whether what is left says anything.
func filterAssistantParts(parts []nacelle.Part, results map[string]bool) ([]nacelle.Part, bool) {
	kept := make([]nacelle.Part, 0, len(parts))
	said := false
	for _, part := range parts {
		switch p := part.(type) {
		case nacelle.ToolCall:
			if !results[p.ID] {
				continue
			}
			said = true
		case nacelle.Text:
			said = said || p.Text != ""
		}
		kept = append(kept, part)
	}
	return kept, said
}

func pruneOrphanedCalls(msgs []nacelle.Message) []nacelle.Message {
	out := make([]nacelle.Message, 0, len(msgs))
	for i, msg := range msgs {
		if msg.Role != nacelle.RoleAssistant {
			out = append(out, msg)
			continue
		}
		var results map[string]bool
		if i+1 < len(msgs) {
			results = toolResultIDs(msgs[i+1])
		}
		if kept, said := filterAssistantParts(msg.Parts, results); said {
			out = append(out, nacelle.Message{Role: nacelle.RoleAssistant, Parts: kept})
		}
	}
	return out
}

func alternate(msgs []nacelle.Message) []nacelle.Message {
	out := make([]nacelle.Message, 0, len(msgs))
	for _, msg := range msgs {
		if len(msg.Parts) == 0 {
			continue
		}
		if len(out) > 0 && out[len(out)-1].Role == msg.Role {
			last := &out[len(out)-1]
			last.Parts = append(append([]nacelle.Part(nil), last.Parts...), msg.Parts...)
			continue
		}
		out = append(out, msg)
	}
	return out
}
