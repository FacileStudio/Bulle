package tui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/apps/cli/internal/status"
	"github.com/FacileStudio/Bulle/core/cost"
	"github.com/FacileStudio/nacelle"
)

const abandoned nacelle.Stop = "abandoned"

// status is the run's state in words: ready, what the run is doing, a cut
// short run, a stop in progress, or an approval waiting.
func (m *Model) status() string {
	state, _ := m.state()
	return state
}

func (m *Model) state() (string, bool) {
	state, attention := "✓ ready", false
	if cut := cutShort(m.run.stop); cut != "" {
		state, attention = cut, true
	}
	if m.run.busy {
		state, attention = m.working(), false
		if time.Since(m.run.interrupted) < forceQuit {
			state, attention = "stopping · ctrl+c or ctrl+\\ to quit now", true
		}
	}
	if m.run.pending != nil {
		state = fmt.Sprintf("approve %s? y once · a always this session · n or esc deny", m.run.pending.Name)
		attention = true
	}
	if m.run.question != nil {
		state, attention = "? answer the question above", true
	}
	if m.session != nil && m.session.HasWriteError() {
		state, attention = "! could not write to session log · "+state, true
	}
	return state, attention
}

// footer is the stats line under the state line: the provider and model being
// billed, in the same words the launch banner uses and following a /model
// switch, then the price when the backend reported one, the input and output
// tokens, and the context size. The output tokens, the context and the price
// all carry the live estimate (liveOut) so they tick as the model writes — the
// price scaled by the last realised cost per token (rate) — and the real
// per-turn usage replaces the estimate the moment a turn ends. A session with a
// grind budget adds what is left of the current run's minimum, measured against
// that run alone, and a session without one looks exactly as it did before.
func (m *Model) footer() []string {
	var spent []string
	label := m.activeBackend
	if m.activeModel != "" {
		if label != "" {
			label += " · "
		}
		label += m.activeModel
	}
	if label != "" {
		spent = append(spent, label)
	}

	total := m.total()
	total.OutputTokens += m.run.liveOut
	total.Cost += m.rate * float64(m.run.liveOut)

	if total.Cost > 0 {
		spent = append(spent, cost.FormatCost(total.Cost))
	}
	if total.Total() > 0 {
		spent = append(spent, tokenTotals(total))
	}
	if m.size > 0 {
		spent = append(spent, contextLoad(m.size+m.run.liveOut, m.engine().Policy))
	}
	run := m.run.usage
	run.OutputTokens += m.run.liveOut
	run.Cost += m.rate * float64(m.run.liveOut)
	if left := m.grind.left(run); left != "" {
		spent = append(spent, "grind "+left)
	}
	return spent
}

func (m *Model) working() string {
	if m.compacting {
		return m.theme.Compacting.Render(m.spin.View() + " compacting session")
	}
	doing := status.WaitingVerb(time.Since(m.run.began))
	if m.run.pending != nil {
		return yellow().Render(m.spin.View() + " waiting for approval")
	}
	switch n := m.running(); n {
	case 0:
	case 1:
		name, ok := m.runningName()
		if ok {
			doing = "running " + name
		}
	default:
		doing = fmt.Sprintf("running %d tools", n)
	}
	if since := m.ongoing(); since != "" {
		doing += " · " + since
	}
	return yellow().Render(m.spin.View() + " " + doing)
}

func (m *Model) running() int {
	n := 0
	for _, g := range m.run.groups {
		if g.End.IsZero() {
			n++
		}
	}
	return n
}

func (m *Model) runningName() (string, bool) {
	for _, g := range m.run.groups {
		if g.End.IsZero() {
			return g.Tool.Name, true
		}
	}
	return "", false
}

func (m *Model) ongoing() string {
	if !m.run.busy || m.run.began.IsZero() {
		return ""
	}
	return status.Lasted(time.Since(m.run.began))
}

// spun advances the spinner. A compaction pass keeps it (and the running-row
// timer that redraws with it) ticking even though run.busy is false on the
// idle and /compact paths — without it that row freezes at the frame and
// elapsed time it had when the pass started.
func (m *Model) spun(message spinner.TickMsg) tea.Cmd {
	return m.spin.Spun(message, m.run.busy || m.hasLiveParallel() || m.compacting)
}
