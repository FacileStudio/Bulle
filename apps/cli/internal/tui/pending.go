package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/approval"
	"github.com/FacileStudio/Bulle/apps/cli/internal/herdr"
)

// This file keeps every approval a tool is blocked on. The shown one is
// run.pending; the rest wait in run.backlog, oldest first. A request whose
// caller has gone away is answered and dropped, and every other request stays
// until somebody answers it: a detached subagent's call outlives the main run
// that was going when it asked, and dropping it unanswered would leave the
// gate locked for good.

// live reports whether the caller behind a request is still waiting. A request
// with no Done channel is treated as the main run's and is gone once the run
// is.
func live(done <-chan struct{}) bool {
	if done == nil {
		return false
	}
	select {
	case <-done:
		return false
	default:
		return true
	}
}

// deny answers a dropped approval without ever blocking the loop.
func deny(req *approval.Request) {
	select {
	case req.Decision <- approval.Denied:
	default:
	}
}

// queueApproval shows req, or holds it behind the one already shown.
func (m *Model) queueApproval(req approval.Request) bool {
	if m.run.pending != nil {
		m.run.backlog = append(m.run.backlog, req)
		return false
	}
	m.run.pending = &req
	herdr.Report(m.herdrClient, herdr.Blocked)
	return true
}

// answerApproval hands the shown approval its decision and brings up the next
// one in line, asking an attached editor about it too.
func (m *Model) answerApproval(decision approval.Decision) tea.Cmd {
	pending := m.run.pending
	m.run.pending = nil
	pending.Decision <- decision
	herdr.Report(m.herdrClient, herdr.Working)
	return m.promoteApproval()
}

// promoteApproval moves the oldest held approval into view.
func (m *Model) promoteApproval() tea.Cmd {
	if m.run.pending != nil || len(m.run.backlog) == 0 {
		return nil
	}
	next := m.run.backlog[0]
	m.run.backlog = m.run.backlog[1:]
	m.queueApproval(next)
	m.layout(m.windowHeight)
	return m.askEditor(next)
}

// dropDead answers and drops every approval and question whose caller has
// gone, and keeps the rest. A run that ends this way never touches a request
// a detached subagent is still waiting on.
func (m *Model) dropDead() tea.Cmd {
	if m.run.pending != nil && !live(m.run.pending.Done) {
		deny(m.run.pending)
		m.run.pending = nil
	}
	kept := m.run.backlog[:0]
	for i := range m.run.backlog {
		if live(m.run.backlog[i].Done) {
			kept = append(kept, m.run.backlog[i])
			continue
		}
		deny(&m.run.backlog[i])
	}
	m.run.backlog = kept
	m.dropDeadQuestion()
	return m.promoteApproval()
}

// dropDeadQuestion dismisses the shown question once its caller has gone.
func (m *Model) dropDeadQuestion() {
	q := m.run.question
	if q == nil || live(q.req.Done) {
		return
	}
	m.run.question = nil
	select {
	case q.req.Answers <- nil:
	default:
	}
}
