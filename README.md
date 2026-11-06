# bulle

Terminal coding agent — the human harness for the
[nacelle](https://github.com/FacileStudio/nacelle) agent SDK.

The binary is named `bulle`. The SDK is a Go library for building agents;
this program is its first consumer and lives to exercise every part of it from
a terminal, where someone is watching: text, reasoning, tools starting and
finishing, why a turn ended, what it cost. It is deliberately small — sessions,
profiles and panes are what a product grows, not what a contract test needs.

**Note**: This repo changed its name from `nacelle-tui` to `bulle` on 2026-09-14. The core SDK (`nacelle`) remains the same library.

## What it does

- Streams one model turn at a time in a full-screen Bubble Tea v2 interface
- Runs on your Bulle server, which picks the backend and model, holds the keys
  and gives the CLI the same assistant name and persona as the web app
- Collapses a long paste into a `[Pasted text #1 +133 lines]` placeholder and
  sends the full text when you submit
- Lets the model read and edit files under a root you choose, run commands when
  `-bash` is on, and call MCP server tools from files
  every other client already has (`-mcp ~/.claude/.mcp.json`)
- Lets the model lay a large job out as steps and keep them current while it
  works, drawn live above the prompt and scrolled to the step in flight
- Fans independent side tasks out to concurrent nested runs with `-subagents`,
  so a wide search or a log dump costs the conversation one answer instead of
  its whole output
- Discovers project context (CLAUDE.md, AGENTS.md) and skills into the system
  prompt, each behind its own flag
- Gates tool calls behind an approval prompt with `-approve-tools`, and trusts
  project hook files only after an explicit, remembered decision

## Stack

| Layer | Tech |
|---|---|
| TUI | Go 1.26.4, `charm.land/bubbletea/v2`, lipgloss v2, glamour v2 |
| Agent | [FacileStudio/nacelle](https://github.com/FacileStudio/nacelle), pinned by tag |
| Files | XDG: `~/.config/bulle/config.yml`, data in `~/.local/share/bulle`, trust and logs in `~/.local/state/bulle` |
| Release | GoReleaser, GitHub Actions on tag push, Homebrew tap `FacileStudio/tap` |

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/FacileStudio/Bulle/main/install.sh | bash
```

Installs to `~/.local/bin` via [facile](https://github.com/FacileStudio/facile), the suite
installer. Pass `--bin-dir <dir>` to change that, `--source` to build from source.

Already have `facile`:

```sh
facile install bulle
```

Or Homebrew:

```sh
brew install FacileStudio/tap/bulle
```

From source, inside the workspace:

```sh
go build -tags goolm -o bulle ./apps/cli
```

`go install github.com/FacileStudio/bulle@latest` no longer works. The module is now
`github.com/FacileStudio/Bulle/apps/cli`, and `go install` resolves it only from tags with the
`apps/cli/` prefix (`apps/cli/vX.Y.Z`), which in turn need a matching `core/vX.Y.Z` tag.

## Mobile app

Download the signed Android APK (Android 8+):
<https://github.com/FacileStudio/Bulle/releases/latest/download/bulle-android.apk>

A SHA-256 checksum sits next to it as `bulle-android.apk.sha256`. Setup and usage are in
[docs/mobile.md](docs/mobile.md).

## Usage

Sign in to your Bulle server once. The server holds the provider keys and the CLI
never sees them:

```sh
bulle login
```

With no address, `bulle login` signs in to https://bulle.facile.studio. Pass a url for another
server. It opens your browser, and a session you already have with another Facile tool completes
the sign-in with no code to copy; over SSH, or with `--device`, it prints an address and asks for
the code the page shows.

Then run it in the directory you want it to work in:

```sh
bulle
```

`bulle whoami` shows who you are, `bulle logout` signs you out. For CI, set
`BULLE_SERVER_URL` and `BULLE_TOKEN`. Details in
[docs/configuration.md](docs/configuration.md#using-a-bulle-server).

There is no standalone mode: without a login, `bulle` refuses to start. The server's
`/api/llm` names the backend, the default model, the models you may switch to and
your assistant's persona. In a session, `/model` lists those models and switches
between them. The operator picks them with `BULLE_LLM_BACKEND`, `BULLE_LLM_MODEL`
and `BULLE_LLM_MODELS`; see [docs/configuration.md](docs/configuration.md#models).

Your own default model is set in the web app, under **Settings → Model**, with a
search over everything the current provider offers. The server returns it from
`/api/llm`, so the next CLI session starts on it, and `/model` browses the same
catalogue. To pin a model on one machine instead, set `session.model` in
`~/.config/bulle/config.yml`; `-model` still wins over both.

A paste of 10 or more lines, or over 1000 characters, shows in the prompt as
`[Pasted text #N +X lines]`. Submitting, or opening the prompt in your editor
(`ctrl+g`), expands every placeholder, so the model always gets the full text.

Settings layer bottom-up: defaults, then `~/.config/bulle/config.yml`, then `BULLE_*`
environment variables, then flags. The useful ones:

| Flag | Env | What |
|---|---|---|
| `-model` | — | one of the server's models; the server's default otherwise |
| `-root` | `bulle_ROOT` | directory the file tools may reach |
| `-bash` | `bulle_BASH` | let the model run commands (off by default) |
| `-continue` | — | auto-resume the newest session for the current project |
| `-resume` | — | resume a specific session by id or file path |
| `-no-config` | — | start with default settings, ignoring ~/.config/bulle/config.yml |
| `-tasks` | `bulle_TASKS` | task planning tool (on by default) |
| `-approve-tools` | `bulle_APPROVE_TOOLS` | ask before every tool call runs |
| `-subagents` | `bulle_SUBAGENTS` | give the model the parallel delegate tool (on by default) |
| `-max-iterations` | `bulle_MAX_ITERATIONS` | how many times the model may be asked; `0`, the default, lifts the cap |
| `-mcp` | — | MCP servers file (repeatable) |
| `-skill-dir` | `bulle_SKILL_DIRS` | extra skills directory (repeatable) |
| `-ide` | `bulle_IDE` | publish the session to an editor over a unix socket |

`bulle -version` prints exactly `bulle <semver>`. See `-h` for the full set:
reasoning effort and budget, project-context and skill
discovery, hooks trust.

Full settings reference: [docs/configuration.md](docs/configuration.md).

## Configuration

Settings live in `~/.config/bulle/config.yml`, **written on first boot with every default
explicitly set** — delete it to regenerate. An existing file is never touched.
The example file with all defaults, and the full reference with the
precedence order and the traps in each setting:
[docs/configuration.md](docs/configuration.md).

`example.bulle.yml` in this repo is the same file the first boot writes — the whole
surface, abridged here. The abridgement is illustrative: the scaffold-parity test covers
`example.bulle.yml`, and nothing reads the block below, so a wrong key here fails no gate —
it fails at load, under `KnownFields`, when you paste it.

The provider, its keys and the models you may switch to come from your Bulle
server (`bulle login <url>`), not from this file.

```yaml
session:
  root: .

limits:
  max_iterations: 0
  max_concurrency: 16
  # compact_at: 75000
  compaction:
    judge:
      enabled: false
      provider: jev

security:
  deny_elevation: true
  env_isolation: false

editor:
  # editor: /usr/bin/vim
  # prompt_edit_key: ctrl+g

sources:
  skill_dirs: []
  mcp: {}
```

The block above is the full scaffold — the file bulle writes on first boot,
and the `TestExampleConfigIsTheScaffoldTemplate` test checks it is byte-identical
to `example.bulle.yml`. The rest of the surface (tools, reasoning, discovery,
ui, sandbox, remote, chat, assistant, gates, sources, hooks) is documented in
[docs/configuration.md](docs/configuration.md); a key missing from the block
above is not a broken setting, it is a setting kept to its default.

## Context compaction

Two tiers: soft (deterministic, no model call) and smart (a judge pass
that classifies each history block and keeps, prunes or folds it).
Only four keys are the user's decisions; the ratios and tail bounds
are internalised defaults — read when present, not documented or
scaffolded:

| Key | Default | What it does |
|---|---|---|
| `limits.compact_at` | unset (backend's window or 75000) | Absolute ceiling; `0` disables compaction |
| `limits.compaction.judge.enabled` | `false` | Opt-in. Sends history to the judge's provider |
| `limits.compaction.judge.provider` | `jev` | `jev` (TypeSafe System One) or `clef` (Cloudflare via OpenRouter) |
| `limits.compaction.window_tokens` | unset | Override for gateways that under-report the window |

The judge is **opt-in and off by default**: turning on
`limits.compaction.judge` sends conversation history — which can include source
code and secrets — to the judge's provider for classification.
`/status` reports the provider and model version that actually answered the
last call.

## Sandboxes & remote hosts

bulle runs on the host and sends every tool call to the target over SSH, so the
provider keys and the local filesystem never enter the environment the model
reaches. Two commands cover the two kinds of target:

```sh
# Local boite microVMs
bulle sandbox list              # configured sandbox.targets and running boite VMs
bulle sandbox pingu             # interactive session inside a VM
bulle sandbox pingu --snapshot  # snapshot the overlay disk when the session ends

# SSH hosts
bulle remote list               # hosts configured under remote.targets
bulle remote staging            # interactive session on a configured host
bulle remote deploy@build:2222 "run the tests"
```

`sandbox` targets resolve to local [boite](https://github.com/FacileStudio/boite)
microVMs: an entry in `sandbox.targets`, then a registered instance of that
name. `remote` targets resolve to a `remote.targets` entry, a direct
`user@host:port` address, or an `~/.ssh/config` host alias. Each reads its
defaults — user, port, identity, workspace — from its own group in `~/.config/bulle/config.yml`.

Scheduled jobs and the chat agent run on the server. The agent is created for you in
Prisme, joins your spaces' channels, and answers when mentioned, in a thread.

## herdr

Run inside [herdr](https://herdr.dev), `bulle` reports its live state and
session identity over herdr's socket API (`apps/cli/internal/herdr/`). A bulle pane
shows as an agent with an idle / working / blocked state, and herdr holds a
reference to the run's transcript. This needs no herdr binary update and works
on any machine, including stock herdr.

After a herdr **server restart**, herdr restores a bulle pane as a plain
shell in its saved directory — bulle is not in herdr's compiled-in resume
table, and no config or plugin adds it. Reopen the session in that directory
with `bulle` (auto-resumes the newest session by cwd) or
`bulle --resume <transcript-path>`. Real auto-restore awaits herdr adding
bulle to its resume table; `--resume` already accepts the exact absolute
transcript path the reporter reports.

## Editors

`bulle --ide` (or `BULLE_IDE=1`) publishes the session to an editor over a unix
socket, so a plugin can watch a run and drive it: mark the lines bulle changes,
put an approval in front of you, run a prompt with your cursor's place
attached, and stop a run. It is off by default, and a session that publishes to
nobody creates no socket, no file and no goroutine.

The contract is [docs/ide-protocol.md](docs/ide-protocol.md); the first client
is [bulle.nvim](https://github.com/FacileStudio/bulle.nvim).

## Structure

```
go.work                        Workspace: core and apps/cli
apps/cli/main.go               Entrypoint: calls cmd.Execute
apps/cli/cmd/                  CLI commands, Cobra tree, flag parsing, and Fang styling
apps/cli/internal/agent/       Agent lifecycle, tools wiring, headless mode, banner, flags
apps/cli/internal/approval/    Interactive and batch tool call approvals
apps/cli/internal/diagnostics/ Post-edit diagnostics
apps/cli/internal/diff/        Syntax-highlighted unified diff generator
apps/cli/internal/herdr/       Reports agent state and session identity to herdr over its socket API
apps/cli/internal/history/     Command history and navigation
apps/cli/internal/ide/         Publishes a session to an editor over a unix socket
apps/cli/internal/layout/      Terminal dimensions and line truncation
apps/cli/internal/menu/        Autocompletion menu for commands and skills
apps/cli/internal/queue/       Input queueing during active turns
apps/cli/internal/status/      Spinner and progress status indicator
apps/cli/internal/tasks/       Task planning tool, validation, step updates
apps/cli/internal/theme/       Terminal color palettes and syntax themes
apps/cli/internal/thinking/    Collapsible reasoning viewport
apps/cli/internal/toolview/    Compact and grouped tool rendering
apps/cli/internal/tui/         Bubble Tea v2 model, key handling, rendering, slash commands
core/assistant/                Assistant pulse, check-in and consolidation
core/chat/                     Matrix transport the server's bot runs on
core/compaction/               Zone/ledger context strategy: Plan, Apply, Tombstone, the judge
core/cost/                     Token usage and cost calculation
core/engine/                   Conversation pruning over the nacelle transcript
core/jev/                      TypeSafe System One client for the opt-in compaction judge
core/provider/                 Provider and backend selection
core/sandbox/                  Sandbox targets and tools over SSH
core/sessions/                 Session persistence, listing, rotation, resume
core/settings/                 CLI flags, ~/.config/bulle/config.yml, and environment configuration
core/skills/                   Agent skills discovery and execution
core/usage/                    Token accounting and context window headroom
core/xdgtest/                  Test helper that isolates XDG directories
```

`core/` is a separate Go module (`github.com/FacileStudio/Bulle/core`) of packages that
carry no terminal UI, so a server can import them. `apps/cli` is the `bulle` binary.

---

Part of the [Facile Suite](https://facile.studio) — self-hosted tools for creative studios
and freelancers. One login, zero cloud dependency.
