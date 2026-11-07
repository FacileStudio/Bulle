package tui

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/FacileStudio/Bulle/core/compaction"
	"github.com/FacileStudio/nacelle"
)

type compactFinished struct{}

func resolvedPolicy(base compaction.Policy) compaction.Policy {
	return base.Resolved()
}

func (m *Model) plan() []compaction.Span {
	return compaction.Plan(m.conversation, m.compactor.Policy)
}

func (m *Model) engine() *compaction.Engine {
	if m.compactor == nil {
		m.compactor = compaction.NewEngine(compaction.Policy{}, nil, nil)
	}
	if m.compactor.Builder == nil && m.agent != nil {
		m.compactor.Builder = func() (*nacelle.Agent, error) {
			if m.agent == nil {
				return nil, nil
			}
			return compaction.BackendAgent(m.agent.Backend())()
		}
	}
	return m.compactor
}

func (m *Model) beginCompaction(ctx context.Context, force bool) tea.Cmd {
	if m.compacting {
		return nil
	}
	engine := m.engine()
	outcome, pending := engine.Prepare(m.conversation, m.size)
	if !pending {
		m.takeStep(outcome)
		return nil
	}

	m.compacting = true
	m.compactBegan = time.Now()

	resultsChan := make(chan compaction.Outcome, 1)
	m.run.compactChan = resultsChan

	conv := compaction.Snapshot(m.conversation)
	size := m.size
	go func() {
		defer close(resultsChan)
		resultsChan <- engine.RunPassWithSize(ctx, conv, size, force)
	}()
	return tea.Batch(waitForCompact(resultsChan), m.spin.Tick)
}

func (m *Model) settleCompaction(outcome compaction.Outcome) tea.Cmd {
	m.compacting = false
	m.run.compactChan = nil

	conv, settled := m.engine().Settle(m.conversation, outcome)
	m.conversation = conv
	m.report(settled)
	m.sayFailure(settled)
	m.sayThrash(settled)

	if m.run.busy && m.agent != nil {
		return m.startRun(m.run.bgCtx)
	}
	return m.deliver()
}

func (m *Model) installFold(outcome compaction.Outcome) {
	start, end, _ := compaction.HistoryRange(outcome.Plan)
	m.size = outcome.After
	m.last = compacted{
		evictCut: end - start,
		turns:    outcome.Fold.LedgerSize(),
		pruned:   outcome.Fold.PrunedSize(),
		kept:     activeLen(outcome.Plan),
		tier:     outcome.Tier,
		replaced: outcome.Applied.LedgerReplaced,
		keptAsIs: outcome.Applied.LedgerKept,
	}
	m.say(fromCompact, compactReport(outcome, m.last))
}

func waitForCompact(results <-chan compaction.Outcome) tea.Cmd {
	return func() tea.Msg {
		next, open := <-results
		if !open {
			return compactFinished{}
		}
		return next
	}
}
