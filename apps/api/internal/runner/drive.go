package runner

import (
	"context"
	"errors"
	"strconv"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/core/engine"
	"github.com/FacileStudio/nacelle"
)

// attempt is one pass of the model over the conversation, and whether the
// provider refused it as too long.
type attempt struct {
	result   outcome
	overflow bool
}

// drive runs the turn. The conversation is kept inside the model's window
// before the first request, and a refusal for length gets one forced
// compaction and one retry.
func (r *Runner) drive(ctx context.Context, lr *liveRun, session store.Session, conv *engine.Conversation) outcome {
	user := strconv.FormatInt(lr.userID, 10)
	agent, closeTools, err := r.buildAgent(ctx, lr, session, user)
	if err != nil {
		return r.failure(err)
	}
	defer closeTools()
	defer r.cfg.Computers.Touch(user)

	r.maintain(ctx, lr, agent, conv, false)
	got := r.stream(ctx, lr, user, agent, conv)
	if got.overflow && ctx.Err() == nil {
		r.maintain(ctx, lr, agent, conv, true)
		got = r.stream(ctx, lr, user, agent, conv)
	}
	return got.result
}

func (r *Runner) stream(ctx context.Context, lr *liveRun, user string, agent *nacelle.Agent, conv *engine.Conversation) attempt {
	events, err := engine.NewSession(agent, conv).Submit(ctx, "")
	if err != nil {
		return attempt{result: r.failure(err)}
	}
	lr.engaged = true
	var got attempt
	for event := range events {
		r.relay(lr, user, event, &got)
	}
	if got.result.status == "" {
		got.result = r.failure(errors.New("the run ended without a result"))
	}
	return got
}

func (r *Runner) relay(lr *liveRun, user string, event engine.Event, got *attempt) {
	switch event.Kind {
	case engine.EventTextDelta:
		lr.emitter.emit(EventText, map[string]any{"delta": r.cfg.Redactor.Apply(event.Text)})
	case engine.EventThinking:
		lr.emitter.emit(EventThinking, map[string]any{"delta": r.cfg.Redactor.Apply(event.Text)})
	case engine.EventToolCall:
		r.cfg.Computers.Touch(user)
		lr.emitter.emit(EventToolStart, map[string]any{
			"call_id": event.Tool.ID, "name": event.Tool.Name, "args": rawArgs(r.cfg.Redactor.Apply(event.Tool.Input)),
		})
	case engine.EventToolResult:
		r.toolDone(lr, user, event.Tool)
	case engine.EventTurnDone:
		got.result = outcome{status: store.RunFinished, stop: string(event.Stop), usage: event.Usage}
	case engine.EventError:
		got.result = r.failure(event.Err)
		got.overflow = engine.DetectOverflow(event.Err)
	}
}

func (r *Runner) toolDone(lr *liveRun, user string, tool *nacelle.ToolEvent) {
	r.cfg.Computers.Touch(user)
	lr.emitter.emit(EventToolEnd, toolEnd(tool, r.cfg.Redactor))
}
