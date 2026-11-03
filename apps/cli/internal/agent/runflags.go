package agent

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/FacileStudio/nacelle"
	"github.com/FacileStudio/nacelle/tools"

	"github.com/FacileStudio/Bulle/apps/cli/internal/approval"
	"github.com/FacileStudio/Bulle/apps/cli/internal/tui"
	"github.com/FacileStudio/Bulle/core/settings"
	"github.com/FacileStudio/Bulle/core/skills"
)

// Run is the main entry point for the agent.
func Run(v string) error {
	if handled, err := checkVersionFlag(v); handled {
		return err
	}
	if handled, err := checkBenchFlag(); handled {
		return err
	}
	if handled, err := checkPrintFlag(); handled {
		return err
	}
	sess, cleanup, err := bootOrAsk(v)
	if err != nil {
		return err
	}
	defer cleanup()
	return tui.Launch(*sess)
}

type preparedTools struct {
	config settings.Config
	set    *tools.Set
	mcp    connected
	local  []nacelle.Tool
}

func setupAgentTools(noConfig bool) (preparedTools, error) {
	flags := settings.FromFlags(settings.Defaults(""))
	if noConfig {
		flags.NoConfig = &noConfig
	}
	config, err := settings.Settings(DefaultSystemPrompt(), flags)
	if err != nil {
		return preparedTools{}, err
	}
	set, local, err := localTools(config)
	if err != nil {
		return preparedTools{}, err
	}
	mcp, local, err := mcpTools(config, local)
	if err != nil {
		return preparedTools{}, closeOnErr(err, set)
	}
	return preparedTools{config: config, set: set, mcp: mcp, local: local}, nil
}

func setupAgentSession(p preparedTools, v string) (*tui.UISession, error) {
	found := augmentSystem(&p.config, p.mcp)
	approvalGate, approve := approval.Build(*p.config.ApproveTools)

	hooks, hookNotice, err := settings.SessionHooks(p.config)
	if err != nil {
		return nil, err
	}
	backend, err := backendFor(&p.config)
	if err != nil {
		return nil, err
	}
	publisher, err := startIDE(p.config, v, backend)
	if err != nil {
		return nil, err
	}
	get, err := assemble(p.config, backend, p.local, approve, mergeHooks(hooks, publisher.Hooks()))
	if err != nil {
		return nil, errors.Join(err, publisher.Close())
	}

	assistant := signedInAssistant(p.config)
	return &tui.UISession{
		Agent:             get.agent,
		Banner:            banner(get.backend, p.config, found, p.mcp, title(v, assistant)),
		Skills:            found.skills,
		HookNotice:        hookNotice,
		Gate:              approvalGate,
		DelegateConfig:    get.config,
		Mode:              *p.config.Mode,
		TransparentBlocks: *p.config.TransparentBlocks,
		BaseURL:           p.config.BaseURL,
		APIKey:            p.config.APIKey,
		Models:            p.config.Models,
		IDE:               ideSurface(publisher),
		SessionConfig:     sessionConfig(p, found, get.backend, assistant),
	}, nil
}

// sessionConfig folds the runtime settings an interactive session reads into
// one struct — the banner and the launch notes take theirs from the same
// snapshot, so the two cannot disagree about what was loaded.
func sessionConfig(p preparedTools, found loaded, backend nacelle.Backend, assistant string) tui.SessionConfig {
	budget := ResolveBudget(p.config.CompactAt, p.config.Compaction, backend)
	return tui.SessionConfig{
		Root:       p.config.Root,
		Model:      resolvedModel(backend, p.config),
		Backend:    p.config.Backend,
		Diffs:      *p.config.Diffs,
		GroupTools: p.config.GroupTools, ShowThinking: *p.config.ShowThinking,
		Compaction:     CompactionConfig(budget, p.config.Compaction),
		MaxConcurrency: resolveConcurrency(p.config.MaxConcurrency, p.config.MaxParallelAgents),
		Grind: tui.GrindConfig{
			Cost:          *p.config.GrindCost,
			Tokens:        *p.config.GrindTokens,
			Continuations: *p.config.GrindContinuations,
		},
		AutoResume:        *p.config.Continue,
		Resume:            *p.config.Resume,
		PromptPlaceholder: *p.config.PromptPlaceholder,
		StartMessage:      welcome(p.config, assistant),
		Editor: tui.EditorConfig{
			PromptEditKey: p.config.PromptEditKey,
			Editor:        p.config.Editor.Editor,
		},
		Hooks: tui.HookUIConfig{
			Show:       settings.DerefBool(p.config.ShowHooks),
			ShowOutput: settings.DerefBool(p.config.ShowHookOutput) || settings.DerefBool(p.config.HookOutput),
		},
	}
}

func buildUISession(v string, noConfig bool) (*tui.UISession, func(), error) {
	prep, err := setupAgentTools(noConfig)
	if err != nil {
		return nil, nil, err
	}
	sess, err := setupAgentSession(prep, v)
	if err != nil {
		return nil, nil, closeOnErr(err, prep.set, prep.mcp.set)
	}
	return sess, func() {
		if err := closeAll(prep.set, prep.mcp.set); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}, nil
}

type loaded struct {
	notice       string
	skills       []skills.Skill
	contextFiles int
}

func augmentSystem(config *settings.Config, mcp connected) loaded {
	var found loaded
	config.System += environment(*config, time.Now(), mcp)
	if *config.ProjectContext {
		text, files := projectContext(config.Root)
		config.System += text
		found.contextFiles = files
	}
	if *config.Skills {
		res := skills.LoadSkills(config.Root, *config.TrustSkills, config.SkillDirs)
		config.System += res.System
		found.notice = res.Notice
		found.skills = res.Skills
	}
	config.System += additionalPrompt(config.Additional)
	return found
}

// additionalPrompt is the session's own steering text, appended last so it
// sits closest to the turn and carries the most weight over the harness
// guidance it layers on. A base prompt replaced via system_prompt is appended
// to just the same: the two settings compose rather than compete.
func additionalPrompt(text string) string {
	if text == "" {
		return ""
	}
	return "\n\n" + text
}
