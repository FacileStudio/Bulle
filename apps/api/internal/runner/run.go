package runner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/core/engine"
	"github.com/FacileStudio/nacelle"
)

type outcome struct {
	status string
	stop   string
	usage  nacelle.Usage
	err    string
}

func (r *Runner) complete(ctx context.Context, lr *liveRun, session store.Session, history []nacelle.Message) string {
	ctx, stop := context.WithTimeoutCause(ctx, r.cfg.RunTimeout, errTimeout)
	defer stop()
	defer lr.cancel(nil)

	lr.emitter.emit(EventRunStarted, map[string]any{"run_id": lr.id})
	conv := engine.NewConversation(history...)
	lr.seen = newProvenance(history)
	result := r.drive(ctx, lr, session, conv)
	result = r.classify(ctx, result)

	background := context.WithoutCancel(ctx)
	if err := r.saveHistory(background, session.ID, r.transcript(conv, result)); err != nil {
		r.cfg.Logger.Error("saving transcript", "run", lr.id, "error", err)
	}
	usage, _ := json.Marshal(usageJSON(result.usage))
	if err := r.cfg.Store.FinishRun(background, lr.id, result.status, result.stop, usage, result.usage.Cost); err != nil {
		r.cfg.Logger.Error("finishing run", "run", lr.id, "error", err)
	}
	r.recordAgentTime(background, lr, session, result.usage)
	if result.err != "" {
		lr.emitter.emit(EventRunError, map[string]any{"message": result.err})
		return result.status
	}
	lr.emitter.emit(EventRunFinished, map[string]any{"stop": result.stop, "usage": usageJSON(result.usage), "cost_usd": result.usage.Cost})
	return result.status
}

// transcript is what the session's stored conversation becomes once a run ends.
//
// A turn that produces no assistant text — the box failed to open, the user
// stopped it before the model spoke, the model errored — would otherwise leave
// this turn's user message next to the previous one, and the engine merges two
// adjacent user messages into a single message on the next run. The stored
// transcript then loses the boundary for good: the client shows what was five
// messages as one bubble of glued sentences, and no reload can split it back.
// Recording the outcome as the assistant's turn keeps the roles alternating.
func (r *Runner) transcript(conv *engine.Conversation, result outcome) []nacelle.Message {
	messages := conv.Messages()
	if len(messages) > 0 && messages[len(messages)-1].Role == nacelle.RoleAssistant {
		return messages
	}
	conv.AppendAssistantText(r.unanswered(result))
	return conv.Messages()
}

func (r *Runner) unanswered(result outcome) string {
	switch {
	case result.err != "":
		return "The run failed: " + result.err
	case result.stop != "" && result.stop != "end":
		return "The run stopped (" + result.stop + ") before I answered."
	default:
		return "The run ended without an answer."
	}
}

func (r *Runner) classify(ctx context.Context, result outcome) outcome {
	cause := context.Cause(ctx)
	switch {
	case errors.Is(cause, errStopped):
		return outcome{status: store.RunStopped, stop: "stopped", usage: result.usage}
	case errors.Is(cause, errInterrupted):
		return outcome{status: store.RunInterrupted, stop: store.RunInterrupted, usage: result.usage}
	case errors.Is(cause, errTimeout):
		return outcome{status: store.RunError, stop: "timeout", usage: result.usage, err: fmt.Sprintf("run timed out after %s", r.cfg.RunTimeout)}
	}
	return result
}

func (r *Runner) drive(ctx context.Context, lr *liveRun, session store.Session, conv *engine.Conversation) outcome {
	user := strconv.FormatInt(lr.userID, 10)
	agent, closeTools, err := r.buildAgent(ctx, lr, session, user)
	if err != nil {
		return r.failure(err)
	}
	defer closeTools()
	defer r.cfg.Computers.Touch(user)

	events, err := engine.NewSession(agent, conv).Submit(ctx, "")
	if err != nil {
		return r.failure(err)
	}
	lr.engaged = true
	var result outcome
	for event := range events {
		switch event.Kind {
		case engine.EventTextDelta:
			lr.emitter.emit(EventText, map[string]any{"delta": r.cfg.Redactor.Apply(event.Text)})
		case engine.EventThinking:
			lr.emitter.emit(EventThinking, map[string]any{"delta": r.cfg.Redactor.Apply(event.Text)})
		case engine.EventToolCall:
			r.cfg.Computers.Touch(user)
			lr.gate.note(event.Tool.ID, event.Tool.Name, event.Tool.Input)
			lr.emitter.emit(EventToolStart, map[string]any{
				"call_id": event.Tool.ID, "name": event.Tool.Name, "args": rawArgs(r.cfg.Redactor.Apply(event.Tool.Input)),
			})
		case engine.EventToolResult:
			r.toolDone(lr, user, event.Tool)
		case engine.EventTurnDone:
			result = outcome{status: store.RunFinished, stop: string(event.Stop), usage: event.Usage}
		case engine.EventError:
			result = r.failure(event.Err)
		}
	}
	if result.status == "" {
		result = r.failure(errors.New("the run ended without a result"))
	}
	return result
}

func (r *Runner) buildAgent(ctx context.Context, lr *liveRun, session store.Session, user string) (*nacelle.Agent, func(), error) {
	backendName := session.Backend
	if backendName == "" {
		backendName = r.cfg.Backend
	}
	model := session.Model
	if model == "" {
		model = r.cfg.Model
	}
	backend, err := r.backend(ctx, lr.userID, backendName, model)
	if err != nil && (backendName != r.cfg.Backend || model != r.cfg.Model) {
		backendName = r.cfg.Backend
		model = r.cfg.Model
		backend, err = r.backend(ctx, lr.userID, backendName, model)
	}
	if err != nil {
		return nil, nil, err
	}
	switch session.Channel {
	case store.ChannelGuest:
		return r.guestAgent(ctx, lr, backend, backendName)
	case store.ChannelIdeas:
		return r.ideasAgent(ctx, lr, backend, backendName)
	}
	vmTools, vm, err := r.lazyVMTools(user, lr.gate)
	if err != nil {
		return nil, nil, err
	}
	host, err := r.hostTools(lr)
	if err != nil {
		return nil, nil, err
	}
	suiteTools, suiteNote := r.suiteTools(lr, r.space(ctx, lr.userID, session.SpaceID))
	jobTools := buildJobTools(r.cfg.Store, r, lr.userID, session.SpaceID, session.Channel)
	system := r.buildSystemPrompt(ctx, lr.userID, session.Channel, suiteNote)
	agent, err := nacelle.New(nacelle.Config{
		Backend:       r.hydrating(nacelle.Retry(backend, nacelle.RetryOptions{}), lr.userID, backendName),
		System:        system,
		Thinking:      nacelle.Thinking{Show: true},
		Tools:         append(append(append(vmTools, host...), jobTools...), suiteTools...),
		MaxIterations: r.cfg.MaxIterations,
		Approve:       lr.gate.approve,
	})
	if err != nil {
		return nil, nil, err
	}
	return agent, vm.close, nil
}

// guestAgent is the agent a colleague of the owner talks to: the owner's model
// and persona with no tool of any kind, so nothing a guest writes can read the
// owner's mail, files or memory, or act for them.
func (r *Runner) guestAgent(ctx context.Context, lr *liveRun, backend nacelle.Backend, backendName string) (*nacelle.Agent, func(), error) {
	agent, err := nacelle.New(nacelle.Config{
		Backend:       r.hydrating(nacelle.Retry(backend, nacelle.RetryOptions{}), lr.userID, backendName),
		System:        r.guestSystemPrompt(ctx, lr.userID),
		MaxIterations: 1,
	})
	return agent, func() {}, err
}

func (r *Runner) toolDone(lr *liveRun, user string, tool *nacelle.ToolEvent) {
	r.cfg.Computers.Touch(user)
	lr.seen.add(tool.Result)
	lr.emitter.emit(EventToolEnd, toolEnd(tool, r.cfg.Redactor))
}

func (r *Runner) failure(err error) outcome {
	return outcome{status: store.RunError, stop: "error", err: r.cfg.Redactor.Apply(err.Error())}
}

func toolEnd(tool *nacelle.ToolEvent, redact Redactor) map[string]any {
	output := tool.Result
	if output == "" && tool.Err != nil {
		output = tool.Err.Error()
	}
	return map[string]any{
		"call_id":        tool.ID,
		"ok":             tool.Err == nil,
		"output_preview": preview(redact.Apply(output)),
		"duration_ms":    tool.Duration.Milliseconds(),
	}
}

func usageJSON(u nacelle.Usage) map[string]any {
	return map[string]any{
		"input_tokens":          u.InputTokens,
		"output_tokens":         u.OutputTokens,
		"cache_read_tokens":     u.CacheReadTokens,
		"cache_creation_tokens": u.CacheCreationTokens,
	}
}
