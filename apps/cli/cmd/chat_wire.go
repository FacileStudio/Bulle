package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/FacileStudio/Bulle/core/assistant"
	"github.com/FacileStudio/Bulle/core/chat"
	"github.com/FacileStudio/Bulle/core/memory"
)

func newChatServices(adapter *chat.Matrix, rooms *roomRegistry, reactions *chat.ReactionRouter) (*chatServices, error) {
	mem, err := memory.Default()
	if err != nil {
		return nil, err
	}
	state, err := assistant.DefaultState()
	if err != nil {
		return nil, err
	}
	ledger, err := assistant.NewLedger(state)
	if err != nil {
		return nil, err
	}
	cfg, err := assistantSettings()
	if err != nil {
		return nil, err
	}
	ap := newApprovals(adapter, reactions)
	rooms.useAsker(ap)
	rooms.nudges = newNudgeLog(reactions)
	rooms.memory = mem
	return &chatServices{
		rooms: rooms, adapter: adapter, approvals: ap, ledger: ledger,
		memory: mem, assistant: cfg, state: state,
	}, nil
}

func startAntenneWatch(ctx context.Context) {
	deps, err := assistantDeps()
	if err != nil {
		return
	}
	go assistant.WatchAntenne(ctx, func() {
		if _, err := assistant.Watch(ctx, deps); err != nil {
			fmt.Fprintf(os.Stderr, "bulle chat: antenne watch: %v\n", err)
		}
	})
}
