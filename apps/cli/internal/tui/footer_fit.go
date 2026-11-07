package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// barFooter is the row under the prompt: the model, the context left and the
// spend on the left, the keys that act right now on the right. The keys are
// placed first; the left side drops its items from the end until both fit.
// What the run is doing is the status line's job, not this row's.
func (m *Model) barFooter() string {
	muted := lipgloss.NewStyle().Foreground(m.theme.Bar.Muted)
	keys := m.barKeys()
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k[0]+" "+k[1])
	}
	right := strings.Join(pairs, "  ")
	items := append([]string{m.activeModel, m.contextLeft()}, m.spentParts()...)
	left := joinLeft(items)
	for len(items) > 0 && lipgloss.Width(left)+1+lipgloss.Width(right) > m.width {
		items = items[:len(items)-1]
		left = joinLeft(items)
	}
	if lipgloss.Width(right) > m.width {
		right = ""
	}
	gap := max(m.width-lipgloss.Width(left)-lipgloss.Width(right), 1)
	return muted.Render(left + strings.Repeat(" ", gap) + right)
}

// joinLeft joins the footer's left items: the model and the context with a
// space, every spend figure after them with a dot.
func joinLeft(items []string) string {
	var out string
	for i, item := range items {
		if item == "" {
			continue
		}
		sep := " "
		if i > 2 {
			sep = " · "
		}
		if out == "" {
			sep = ""
		}
		out += sep + item
	}
	return out
}

// spentParts is the session's price and tokens, and what is left of a grind
// budget, from the same figures as footer.
func (m *Model) spentParts() []string {
	parts := m.footer()
	if len(parts) > 0 && m.activeModel != "" {
		parts = parts[1:]
	}
	var keep []string
	for _, p := range parts {
		if !strings.HasPrefix(p, "↕") {
			keep = append(keep, p)
		}
	}
	return keep
}
