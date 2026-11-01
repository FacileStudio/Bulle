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

func (r *Runner) execute(ctx context.Context, lr *liveRun, session store.Session, history []nacelle.Message) {
	defer r.release(lr.userID)
	defer func() {
		r.mu.Lock()
		delete(r.live, lr.id)
		r.mu.Unlock()
	}()
	ctx, stop := context.WithTimeoutCause(ctx, r.cfg.RunTimeout, errTimeout)
	defer stop()
	defer lr.cancel(nil)

	lr.emitter.emit(EventRunStarted, map[string]any{"run_id": lr.id})
	conv := engine.NewConversation(history...)
	result := r.drive(ctx, lr, session, conv)
	result = r.classify(ctx, result)

	background := context.WithoutCancel(ctx)
	if err := r.saveHistory(background, session.ID, conv.Messages()); err != nil {
		r.cfg.Logger.Error("saving transcript", "run", lr.id, "error", err)
	}
	usage, _ := json.Marshal(usageJSON(result.usage))
	if err := r.cfg.Store.FinishRun(background, lr.id, result.status, result.stop, usage, result.usage.Cost); err != nil {
		r.cfg.Logger.Error("finishing run", "run", lr.id, "error", err)
	}
	if result.err != "" {
		lr.emitter.emit(EventRunError, map[string]any{"message": result.err})
		return
	}
	lr.emitter.emit(EventRunFinished, map[string]any{"stop": result.stop, "usage": usageJSON(result.usage), "cost_usd": result.usage.Cost})
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
			r.cfg.Computers.Touch(user)
			lr.emitter.emit(EventToolEnd, toolEnd(event.Tool, r.cfg.Redactor))
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
	backend, err := r.cfg.Backends(backendName, model)
	if err != nil {
		return nil, nil, err
	}
	target, err := r.cfg.Computers.Ensure(ctx, user)
	if err != nil {
		return nil, nil, fmt.Errorf("your computer is not available: %w", err)
	}
	vm, closer, err := r.cfg.VMTools(target)
	if err != nil {
		return nil, nil, fmt.Errorf("opening your computer: %w", err)
	}
	host, err := r.cfg.HostTools()
	if err != nil {
		_ = closer.Close()
		return nil, nil, err
	}
	agent, err := nacelle.New(nacelle.Config{
		Backend:       nacelle.Retry(backend, nacelle.RetryOptions{}),
		System:        r.cfg.System,
		Thinking:      nacelle.Thinking{Show: true},
		Tools:         append(vm, host...),
		MaxIterations: r.cfg.MaxIterations,
		Approve:       lr.gate.approve,
	})
	if err != nil {
		_ = closer.Close()
		return nil, nil, err
	}
	return agent, func() { _ = closer.Close() }, nil
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
