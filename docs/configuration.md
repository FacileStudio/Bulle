# bulle — Configuration

Every field the core `Config` reads, then the client's own settings layer — a separate thing,
with its own precedence order — and the traps in each.

## Using a Bulle server

The CLI runs on a Bulle server. The server owns the provider, its keys and the models a session may
use. The CLI keeps only the server URL and a login token, so a laptop never sees a vendor key or
picks a backend. Without a login, `bulle` refuses to start and asks you to run `bulle login`.

```sh
bulle login https://bulle.example.org
bulle whoami
```

| Command | What it does |
|---|---|
| `bulle login [url]` | Opens your browser, signs you in, and stores a token. The URL is saved as `server.url`. The URL is optional: without one, login uses `BULLE_SERVER_URL`, `server.url`, the last login's server, or https://bulle.facile.studio. Over SSH, with `--device`, or when no browser opens, it prints an address to open in any browser, on any device, and asks you to paste the code that page shows |
| `bulle logout` | Deletes the stored token. `server.url` stays |
| `bulle whoami` | Prints the signed-in user and the server. Fails when the token is missing or no longer valid |

`bulle status` also reports whether the server is reachable, whether the token is valid, and which
backends have a key.

### Server settings

| Setting | Env | What |
|---|---|---|
| `server.url` | `BULLE_SERVER_URL` | Base URL of the Bulle server. The environment variable wins over the file |
| (none) | `BULLE_TOKEN` | Login token. Set it for CI and other machines where you cannot open a browser. It wins over a stored token |

```yaml
server:
  url: https://bulle.example.org
```

### Token storage

`bulle login` stores the token in the OS keyring. When no keyring is available, it writes
`$XDG_DATA_HOME/bulle/token` (default `~/.local/share/bulle/token`) with mode `0600`. The token
never goes into `config.yml`.

### How model calls work

The server exposes a proxy at `/api/llm/<backend>/` for `anthropic`, `openai`, `google` and
`openrouter`. With a server configured, the CLI points each backend at
`<server>/api/llm/<backend>` and sends the Bulle token as its credential. The server looks up the
real vendor key, swaps the auth header, and streams the response back. The CLI never holds the
vendor key, and the server never returns one to a client.

A backend with no key anywhere answers `409` with `{"error":"no key for <backend>"}`.

### Models

The CLI reads its backend and model from `GET /api/llm` at the start of every run:

```json
{"backend":"openrouter","model":"xiaomi/mimo-v2.6-flash","models":["xiaomi/mimo-v2.6-flash","anthropic/claude-sonnet-5.5"]}
```

The operator sets these in the API's environment:

| Variable | Default | What |
|---|---|---|
| `BULLE_LLM_BACKEND` | `openrouter` | The backend every call goes through |
| `BULLE_LLM_MODEL` | `xiaomi/mimo-v2.6-flash` | The default model |
| `BULLE_LLM_MODELS` | the default model | Comma-separated models a client may switch to. The default model is always included |

A session may pick another model only from that list: `/model` in the TUI, `--model` on the
command line. A model the server does not offer is refused.

The models a client may switch to are the **provider's whole catalogue**, not only
`BULLE_LLM_MODELS`: the server asks the vendor directly, caches the answer for ten minutes and
falls back to the operator's list when the vendor cannot be asked. `GET /models` returns that catalogue with
display names, and the web settings search it. `BULLE_LLM_MODELS` still leads the list.

A user can also save their own default model in the web settings, under **Model**, by searching
that catalogue. The server remembers it per user, and `GET /api/llm` returns it as `model`, so
the next `bulle` session starts on it without a flag. An empty choice means `BULLE_LLM_MODEL`, and
a saved model that has left the catalogue falls back to it rather than being handed to a client
that would be refused.

On one machine, `session.model` in `~/.config/bulle/config.yml` does the same without the web
page: it names the model that machine starts on, and `-model` still wins over it.

### Provider keys on the server

Each user can add their own key per backend in the web UI, under settings, providers. The page
shows the last four characters of each key and where it comes from: `user`, `instance` or `none`.
Key fields are write-only, and each row has a test button that makes one small call to the vendor.

The server resolves a key in this order:

1. The signed-in user's own key for that backend.
2. The instance key from the server's environment (`ANTHROPIC_API_KEY`, `OPENAI_API_KEY`,
   `GEMINI_API_KEY` or `GOOGLE_API_KEY`, `OPENROUTER_API_KEY`).
3. No key: the request fails.

### For operators

User keys are encrypted at rest with AES-256-GCM. Set `BULLE_CREDENTIALS_KEY` in the API's
environment to a base64-encoded 32-byte key:

```sh
openssl rand -base64 32
```

Keep it stable. Changing or losing it makes every stored user key unreadable, and users have to
enter them again. Instance keys do not use it. The example is in `apps/api/.env.example`.

### Time tracking in Sablier

When the Sablier suite app is configured (`SUITE_SABLIER_URL` and `SUITE_SABLIER_AUDIENCE`),
Bulle books its own work to Sablier as the user's agent. Each run of 5 seconds or more becomes one
entry through `POST /api/agents/sessions`, under the assistant's name and avatar. Time spent
waiting for an approval is not counted as active. A message in a web or Matrix session also posts
a heartbeat for the user (`POST /api/heartbeats`), at most one per session every 2 minutes. Both
calls run in the background with a 10 second limit: a Sablier failure is logged and never slows or
fails a run.

| Variable | Default | What |
|---|---|---|
| `BULLE_SABLIER_TIME_TRACKING` | `true` | Set to `false` to stop reporting runs and heartbeats to Sablier |

## `bulle.Config`

The core reads nothing from the environment or from disk. Every field is passed in the struct
literal a consumer builds.

| Field | Default | What it does |
|---|---|---|
| `Backend` | — required | The model this agent runs on. No default: a package that picks one for you is a package that hides the most consequential decision in the configuration |
| `System` | — required | The system prompt. Empty is refused rather than defaulted — an agent with no system prompt is a general-purpose assistant wearing a product's name |
| `Thinking.Effort` | the backend's own default | `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, `max`. `none` is the only value that asks for no reasoning, and a model whose reasoning is mandatory answers it with a 400 rather than ignoring it (measured on `stealth/ox-alpha`, 2026-08-23). Empty hands the decision to the provider, which on a model whose reasoning is mandatory means that model's own maximum. Not validated against the model: OpenRouter clamps a level a model does not advertise rather than refusing it |
| `Thinking.Budget` | `0` (no ceiling from here) | Reasoning-token ceiling for one turn. The other spelling of `Effort`. Anthropic takes both at once; OpenRouter answers the pair with a 400 (`Only one of reasoning.effort and reasoning.max_tokens can be specified`, measured 2026-08-23) and its backend resolves that by letting the budget win, since the gateway defines each level as a percentage of the budget anyway. Refused at construction when it cannot work: at or above `MaxTokens` it leaves the turn nothing to answer with, and below a backend's `Capabilities.MinBudget` (1024 on `anthropic`) the API rejects it |
| `Thinking.Show` | `false` | Stream the model's reasoning as `KindThinking`. Off by default, matching the APIs: the raw chain of thought is never returned and a readable summary is opt-in. It changes what a consumer is shown and nothing else. The reasoning always travels back over the wire, because that is what the tool loop replays to keep the model's train of thought intact across a tool call |
| `MaxTokens` | `DefaultMaxTokens` (32000) | Per-turn output ceiling. Generous on purpose — every request is streamed, so a large ceiling costs nothing in latency, while a small one truncates an answer mid-sentence and buys a retry |
| `MaxIterations` | `0` (no cap) | How many times the model may be asked. A cap is only safe when every tool is read-only and cheap; reaching it ends the run with `Stop: StopIterations`, not a failure |
| `Tools` | `nil` | Tools the model may call in this process |
| `MCP` | `nil` | MCP servers the model may call tools on. Only `anthropic` can reach them — see [architecture.md](architecture.md#capabilities) |
| `Logger` | `slog.Default()` | Receives the few things worth recording that are not events — retry attempts, mainly |

`Backend`, `Thinking` and `MCP` are all checked against the chosen backend's
`Capabilities` at construction; a config asking for something the backend cannot do is refused
with an `*Unsupported` error rather than silently running with less.

## `anthropic.Config`

| Field | Default | What it does |
|---|---|---|
| `Client` | built from the environment | Set to share a client, point at a proxy, or hand a test a stub transport |
| `Model` | `DefaultModel` (`claude-opus-5`) | — |

## `openrouter.Config`

| Field | Default | What it does |
|---|---|---|
| `APIKey` | `OPENROUTER_API_KEY` | — |
| `Model` | — required | An OpenRouter model slug, e.g. `anthropic/claude-opus-5`. Required: OpenRouter fronts hundreds of models and a default would be this package choosing the one thing the caller came here to choose |
| `BaseURL` | `DefaultBaseURL` | Point at a compatible gateway or a recording proxy in tests |
| `Referer`, `Title` | — | Attribution only — `HTTP-Referer` and `X-OpenRouter-Title`. Neither affects the answer; without `Referer` the usage simply does not appear against an app |
| `Provider` | `nil` | OpenRouter's provider-routing object, passed through untouched. Worth setting `require_parameters: true` when tool calling matters — it keeps the request away from providers that would drop the tool schema |
| `Options` | `nil` | Extra request options for the underlying client |

## `google.Config`

| Field | Default | What it does |
|---|---|---|
| `APIKey` | `GEMINI_API_KEY` or `GOOGLE_API_KEY` | — |
| `Model` | `DefaultModel` (`gemini-3.7-flash`) | The Gemini model to run |
| `BaseURL` | `DefaultBaseURL` (`https://generativelanguage.googleapis.com/v1beta/openai`) | Point at a compatible proxy or gateway |
| `Options` | `nil` | Extra request options for the underlying client |

## `openai.Config`

| Field | Default | What it does |
|---|---|---|
| `APIKey` | `OPENAI_API_KEY` | — |
| `Model` | `DefaultModel` (`gpt-5.4`) | The OpenAI model to run |
| `BaseURL` | `DefaultBaseURL` (`https://api.openai.com/v1`) | Point at a compatible gateway or proxy in tests |
| `Organization` | `""` | Optional `OpenAI-Organization` header |
| `Project` | `""` | Optional `OpenAI-Project` header |
| `Options` | `nil` | Extra request options for the underlying client |

## `tools.Config`

| Field | Default | What it does |
|---|---|---|
| `Root` | — required | The directory every tool is confined to, through `os.Root` |
| `AllowBash` | `false` | Mounts the command tool. Not a limit that can be tuned into safety — running commands is either something this agent may do or it is not |
| `CommandEnv` | `PATH` + `HOME` only | The process environment is never inherited — it is where a service keeps the credentials a model must not read |
| `CommandTimeout` | `DefaultCommandTimeout` (2 min) | A command outliving it is killed with its children |
| `MaxOutputBytes` | `DefaultMaxOutputBytes` (64 KiB) | Caps what one tool call may return — output stays in the context window for the rest of the conversation |
| `MaxReadBytes` | `DefaultMaxReadBytes` (48 KiB) | Caps a single file read, below the output cap so a read that hit the limit still leaves room for the notice saying so |

## The TUI's settings layer

The client's settings layer is a separate settings problem from the core, and deliberately does not share a
mechanism with it: **the library must never read configuration from disk or environment.** A
`nacelle` package that read `~/.config/bulle/config.yml` would let a file on a *different* machine's disk
change a headless consumer's behaviour, the way Bubble Tea's own config never reaches past the
program that imports it.

### Precedence

**Flag beats environment beats file beats default**, resolved in one function
(`settings.Load()` in `core/settings/settings.go`) and nowhere else. The suite has already paid for the
alternative once: a CLI that read its environment inside one code branch and its file inside
another turned what its README called "overrides" into two mutually exclusive modes nobody
could tell apart.

| Layer | Source | Notes |
|---|---|---|
| Flags | `-model`, `-effort`, `-root`, `-system-prompt`, `-additional-prompt`, `-bash`, `-thinking`, `-project-context`, `-skills`, `-trust-skills`, `-skill-dir`, `-mcp`, `-approve-tools`, `-diffs`, `-show-hooks`, `-show-hook-output`, `-max-iterations`, `-compact-at`, `-max-concurrency`, `-max-parallel-agents`, `-tasks`, `-continue`, `-resume`, `-gates-file`, `-no-config` | Only flags actually **typed** are collected, via `flag.Visit` — Go's `flag` package cannot otherwise tell a flag left alone from one passed its own default value. `-skill-dir` and `-mcp` are repeatable (`-mcp a.json -mcp b.json`); every other flag keeps only its last occurrence. `-resume` names one session by id or file path and, when given, beats `-continue`. `-no-config` skips `~/.config/bulle/config.yml` entirely: defaults plus environment plus flags. An invalid file gets a coloured report and one prompt — yes boots with defaults, no exits with the documentation link |
| Environment | `BULLE_SERVER_URL`, `BULLE_TOKEN`, `BULLE_EFFORT`, `BULLE_ROOT`, `BULLE_SYSTEM_PROMPT`, `BULLE_ADDITIONAL_PROMPT`, `BULLE_BASH`, `BULLE_THINKING`, `BULLE_PROJECT_CONTEXT`, `BULLE_SKILLS`, `BULLE_TRUST_SKILLS`, `BULLE_SKILL_DIRS`, `BULLE_APPROVE_TOOLS`, `BULLE_DIFFS`, `BULLE_SHOW_HOOKS`, `BULLE_SHOW_HOOK_OUTPUT`, `BULLE_MAX_ITERATIONS`, `BULLE_COMPACT_AT`, `BULLE_MAX_CONCURRENCY`, `BULLE_MAX_PARALLEL_AGENTS`, `BULLE_COMPACTION_JUDGE`, `BULLE_COMPACTION_JUDGE_PROVIDER`, `BULLE_TASKS`, `TYPESAFE_API_KEY` | A misspelt boolean (`BULLE_BASH=yez`) is treated as unmentioned, not as `false`, and falls through to the layer below. `BULLE_SKILL_DIRS` is colon-separated, the same convention `PATH` itself uses for a list of directories. `TYPESAFE_API_KEY` is the compaction judge's key and beats `BULLE_COMPACTION_JUDGE_API_KEY`; it is the one credential to keep in the environment rather than the file. The compaction ladder's own ratios and tail bounds are sunstopped — they are read when present and still take effect, but they are no longer documented or scaffolded. `BULLE_COMPACTION_JUDGE_PROVIDER` names the judge's decision model (`jev` or `clef`) and is the one new judge key. |
| File | `~/.config/bulle/config.yml` | Preferences only, **no provider settings and no provider credentials**: those live on the server. The compaction judge can name the program that prints its key with `limits.compaction.judge.api_key_command` — see [Keeping the judge key out of the file](#keeping-the-judge-key-out-of-the-file). A `provider:`, `profile:` or `tools.web_fetch` key is refused with a message saying where the setting went. `KnownFields(true)`: an unrecognised key (`max_iteration:`, one letter short) is refused rather than silently ignored |
| Server | `server.url` and the stored token | The backend, model, model list, proxy URL and credential, always. See [Models](#models) |
| Defaults | — | `root: .`, `tools.bash: true`, `reasoning.thinking: true`, `discovery.project_context: true`, `discovery.skills: true`, `discovery.trust_skills: false`, `discovery.trust_hooks: false`, `sources.skill_dirs: []`, `sources.mcp: {}`, `security.approve_tools: false`, `security.deny_elevation: true`, `ui.diffs: true`, `ui.show_hooks: true`, `ui.show_hook_output: true`, `limits.max_iterations: 0` (no cap), `limits.compact_at: unset` (an absolute override when set; unset derives the ceiling from `soft_ratio` × the context window, `0` disables), `limits.max_concurrency: 16`, `limits.max_parallel_agents: 16`, `limits.compaction.judge.enabled: false`, `limits.compaction.judge.provider: jev` (`limits.compaction.window_tokens` and `reserve_tokens` are unset: the first is the backend's own window, the second a fifth of it), `tools.todo_write: true`, `tools.parallel_agents: true`, `ui.rendering_mode: tui`, `ui.group_tools: true`, `ui.show_thinking: true` |

`project_context` and `skills` default **on**, unlike `bash`: each fails soft to nothing when
there is nothing to find — no `AGENTS.md`/`CLAUDE.md` anywhere above `root`, no
`~/.agents/skills/` — so a machine without either is no worse off for asking. `trust_skills` and `approve_tools` both default **off**, and for the same
underlying reason: a project's own `.agents/skills/` can carry instructions to run arbitrary
scripts, and asking before every tool call is a decision that changes how the client feels to
use — blanket-trusting a directory, or blanket-interrupting every call, are both decisions only
the person running this should opt into, not defaults sprung on them. See
[Context and skills](#context-and-skills) and
[Tool approval](#tool-approval) below.

### Checking that everything runs

`bulle status` prints one line per check (`ok`, `warn` or `FAIL`): the config
loads, the server answers and the token is valid, and which backends have a key. It exits 1 when
any line says `FAIL`.

### Where bulle keeps its files

Bulle follows the XDG base directory spec. A relative `XDG_*` value is ignored.

| Directory | Default | Holds |
|---|---|---|
| Config | `$XDG_CONFIG_HOME/bulle` or `~/.config/bulle` | `config.yml` |
| Data | `$XDG_DATA_HOME/bulle` or `~/.local/share/bulle` | `memory/`, `sessions/`, `token`, `server` |
| State | `$XDG_STATE_HOME/bulle` or `~/.local/state/bulle` | `trust.json`, `ledger.json`, `logs/` |
| Runtime | `$XDG_RUNTIME_DIR/bulle`, else `/run/user/<uid>/bulle`, else `$TMPDIR/bulle-<uid>` | `ide/`, IDE sockets |

On first run, bulle moves an older `~/.bulle/` and `~/.bulle.yml` into these
directories. It never overwrites: an entry whose destination already exists
stays where it was.

### `~/.config/bulle/config.yml`

The file is grouped: each family of settings lives under its own key, the same
families the code groups them into. The old flat layout (every key at the top
level) was replaced by these groups in 0.44.0 — a flat key is now a parse
error, because `KnownFields(true)` refuses anything unknown.

```yaml
server:
  url: https://bulle.example.org
reasoning:
  effort: high
  thinking: true
  budget: 8192
limits:
  max_iterations: 0
  max_concurrency: 16
  max_parallel_agents: 16
  # compact_at is an absolute token ceiling. Left unset (the default) the
  # ceiling comes from the backend's context window; 0 turns compaction
  # off outright.
  # compact_at: 75000
  compaction:
    judge:
      enabled: false
      provider: jev
session:
  root: .
  # the model this machine starts on, from the server's catalogue; -model wins
  model: anthropic/claude-sonnet-5.5
  system_prompt: You are a terminal coding assistant.
  # appended after the base prompt and everything layered onto it; leave it out
  # to keep the harness prompt alone, or set it to layer a persona on top.
  additional_prompt: Keep answers under 200 words; always name the file you changed.
  continue: false
tools:
  bash: true
  todo_write: true
  parallel_agents: true
  # every commit the agent creates ends with the Co-authored-by trailer;
  # false turns that off
  git_coauthor: true

security:
  approve_tools: false
  path_isolation: false
  deny_elevation: true
  env_isolation: false

discovery:
  project_context: true
  skills: true
  trust_skills: false
  trust_hooks: false
ui:
  prompt_placeholder: 'Ask something. Esc stops a run, ctrl+c stops or quits, ctrl+\ forces it.'
  rendering_mode: tui
  group_tools: true
  show_thinking: true
  diffs: true
  transparent_blocks: true
  show_hooks: true
  show_hook_output: true
  # cron_list_json: false
  # start_message: |-
  #   Welcome to bulle.
  #     _   _
  #    | | | |
  #    | |_| |
  #    |_____|
sources:
  skill_dirs:
    - ~/.claude/skills
  mcp:
    mycelium:
      command: mycelium
      args: [mcp]
hooks: []
# Gates are deterministic checks run after every edit: the edited path is
# appended to the command's arguments, findings are injected into the edit
# result. No gates key keeps the built-in filet check.
gates: []
# gates:
#   - name: filet
#     command: [filet, check]
#     timeout_secs: 30
#   - name: prettier
#     command: [prettier, --write]
#     format: true
```

Old flat key → new home, for migrating a pre-0.44 file:

| Old key | New location |
|---|---|
| `backend`, `model`, `base_url`, `api_key` | removed: the server owns the provider |
| `max_iterations`, `compact_at`, `max_concurrency`, `max_parallel_agents` | `limits:` |
| `bash` | `tools.bash` |
| `subagents` | `tools.parallel_agents` |
| `fetch` | removed: there is no fetch tool |
| `tasks` | `tools.todo_write` |
| `approve_tools`, `path_isolation`, `deny_elevation`, `env_isolation` | `security:` |
| `effort`, `thinking`, `reasoning_budget` (now `budget`) | `reasoning:` |
| `project_context`, `skills`, `trust_skills`, `trust_hooks` | `discovery:` |
| `continue` | `session:` |
| `resume` | dropped from the file — `-resume` on the command line only |
| `mode` (now `rendering_mode`), `json` (now `cron_list_json`), `group_tools`, `show_thinking`, `diffs`, `prompt_placeholder`, `start_message`, `transparent_blocks` | `ui:` |
| `skill_dirs`, `mcp` | `sources:` |

The tool keys renamed with nacelle v0.32 still read under their old names, and the new name wins when a file carries both: `tools.run_command` is `tools.bash`, `tools.tasks` is `tools.todo_write`, `tools.search_content` is `tools.grep`, and `tools.find_files` is `tools.glob`.

**Environment placeholders.** A value may contain `${NAME}`, which is replaced by that environment
variable, or by nothing when it is unset. This keeps personal values out of a config you track in a
public repo:

```yaml
server:
  url: ${BULLE_SERVER_URL}
editor:
  editor: ${EDITOR}
```

Only the braced form is replaced; a bare `$NAME`, the form shell commands use, is left alone. The
replacement happens after the YAML is parsed, so a multi-line value cannot break the file, and the
file is checked as written first, so an error names the line you wrote.

Every field is optional. A missing file is not an error — on first boot
bulle writes this file itself with every default explicit, and you edit from
there; deleting it regenerates it. But an unreadable or malformed one is: a
unreadable or malformed one is: a config silently ignored is worse than no config, because the
setting carefully written is simply not in effect and nothing says so.

Scheduled jobs and the Matrix bot live on the server; see [Matrix](#matrix).

`prompt_placeholder` is the ghost text shown while the prompt is empty. The prompt has no prefix:
the first row opens at a single margin space, and wrapped rows of a long question hang under it.

`start_message` is printed as the first thing on launch, above the version banner. It is a plain
string that may span lines, so use a `|`-literal (or `|-` to drop the trailing newline) to put a
welcome block or an ascii banner there. Left unset, a signed-in session opens on your assistant's
blob avatar, with its name after the version in the banner, fetched from the server's `/api/auth/me` and drawn with
[`github.com/FacileStudio/blob`](https://github.com/FacileStudio/blob), the same library the web app
uses. With no login, with `NO_COLOR` set, or when the server does not answer within 1.5 seconds, the
default text logo shows instead. An empty string prints nothing.

**No per-project `./.bulle.yml` yet.** A second precedence layer before the first has real
users is a layer nobody has asked for the shape of.

### Keeping the judge key out of the file

`limits.compaction.judge.api_key_command` replaces the judge's literal key with the command that
produces it, so the file carries no secret:

```yaml
limits:
  compaction:
    judge:
      enabled: true
      api_key_command: tiroir get TYPESAFE_API_KEY
```

The command runs through `sh -c`, so a whole program with its arguments, a pipeline, or
`tiroir unlock && tiroir get K` all work. It prints the key on stdout; trailing whitespace is
trimmed. stderr is read for the error message and never for the key, so a command that warns on
stderr still works. Its stdin is closed rather than inherited, so a command that stops to prompt
fails against end-of-file instead of seizing your terminal, and it is killed after 10 seconds.

**It is a source, not an override.** A key any layer already supplied wins — the environment or a
literal `api_key` — and the command only fills what is still empty. The
practical consequence is worth stating plainly: **if `TYPESAFE_API_KEY` is exported on your machine,
a judge `api_key_command` never runs**, because the environment already answered. Unset the
variable when you want the command to take over. Nothing else changes: a key that arrives from the
environment keeps working with no config edit at all.

Two more properties, both deliberate:

- The command runs only where the key is actually used — when a session builds the judge — and
  never for inspection. `bulle list` and `bulle sessions` do not run it,
  so a locked secret store cannot break them.
- A command that fails, times out or prints nothing is **refused at startup**, naming the field,
  rather than left as an empty key. An empty key would reach the backend as "no credential" and read
  exactly like a config that forgot one. A locked tiroir fails loudly instead.

The environment-variable route remains the shortest one when the key is already exported; the
command route is for when it is not, and for the files you keep in git.

### Switching models

`/model` opens a picker of the models the server offers; `/model <name>` switches to one of them.
The backend, proxy and token stay the same, and the conversation history stays intact. A model the
server does not list is refused. See [Models](#models).

### Compaction

The session measures itself against the backend's context window and compacts
at the ratio it has crossed, instead of at one absolute token count. Two tiers:

| Tier | Crossed at | What it does | Model calls |
|---|---|---|---|
| soft | `soft_ratio` × usable window | Tombstones oversized old tool results — deterministic, no model call | 0 |
| smart | `smart_ratio` × usable window | Classifies each history block and folds it into one `[state ledger]` message, keeping, pruning or folding; forces the fold when the gentle one would not land | 1 judge + 1 ledger |

Only four keys are the user's decisions. The ratios and tail bounds below are
internalised defaults — they are read when present and still take effect from old
files, but they are no longer scaffolded or documented; delete them from anyone's
file once they know why the ladder works.

`limits.compact_at` (the absolute ceiling; `0` disables; unset derives the
ceiling from the ratios and the backend's reported window) is the override for
gateways that under-report the window and the one escape hatch the ladder does
not own. `limits.compaction.window_tokens` pins the figure the ladder is
measured against when the backend reports nothing, and the reserve the top rung
leaves for the model's own answer is a fifth of that window within shipped
bounds.

| Key | Default | What it does |
|---|---|---|
| `limits.compact_at` | unset (backend's window or 75000) | Absolute ceiling; `0` disables compaction |
| `limits.compaction.judge.enabled` | `false` | Opt-in. Sends history to the judge's provider — the one setting that leaves the machine |
| `limits.compaction.judge.provider` | `jev` | Decision model: `jev` (TypeSafe System One) or `clef` (Cloudflare via OpenRouter) |
| `limits.compaction.window_tokens` | unset | Override for gateways that under-report the context window |

**The judge is opt-in and off by default.** Turning on
`limits.compaction.judge` sends conversation history — which can include
source code and secrets — to the judge's provider for classification, so it is
never a shipped default and has to be asked for. Its key prefers its vendor's
own environment variable — `TYPESAFE_API_KEY` for `jev`, `OPENROUTER_API_KEY`
for `clef` — over `limits.compaction.judge.api_key`.
With the judge off, compaction behaves exactly as it did before this ladder
existed.

Enabling it is one key in `~/.config/bulle/config.yml`. No flag and no environment variable are
required:

```yaml
limits:
  compaction:
    judge:
      enabled: true          # sends history to the judge's provider on every summarizing pass
      provider: jev
      # model, base_url, api_key, prune_threshold, max_blocks_per_call are
      # internalised defaults — read from the file when present but not
      # documented or scaffolded. Leave them out unless you know you need them.
      # api_key_command: tiroir get TYPESAFE_API_KEY
```

`/status` reports the provider and model version that actually answered the last
call.

**Calibrating the threshold.** The prune threshold and the confidence floor are the two numbers
that decide whether a block is deleted. The calibration harness classifies a corpus of labeled
history blocks in one call, re-scores the answers across a sweep of thresholds, and prints where
the model's confidence and its accuracy diverge. It talks to the network and it bills, so it is
opt-in with `BULLE_CALIBRATION=1`, and each provider runs it against its own vendor key:

```sh
BULLE_CALIBRATION=1 TYPESAFE_API_KEY=... go test ./core/compaction -run JudgeCalibration -v
BULLE_CALIBRATION=1 BULLE_CALIBRATION_PROVIDER=clef OPENROUTER_API_KEY=... go test ./core/compaction -run JudgeCalibration -v
```

Every provider ships with its own measured numbers: on the current corpus `jev-1.13.0` agrees with
the labels 76% of the time at the shipped threshold, and `cloudflare/clef` 65% with zero false
prunes across the whole sweep (measured 2026-10-07, see `core/compaction/calibration_test.go`).
Re-run the harness against a provider before moving its numbers.

## Context and skills

Five things the TUI adds beyond flags and the model's own tools, all client-side, none of them
in the core `nacelle` package — the same "the library must never read configuration from disk"
rule above applies to these too, not only to settings.

**This session's own facts** (always on, `environment.go`). Prepended to whatever `System`
says, before context and skills, because it is the most general layer and the only one with no
switch: the absolute working directory, that the file and search tools are confined to it and
strip a leading `/` rather than refusing it, today's date, that the answer is rendered as
markdown, and whether anyone reviews a tool call before it runs. Two clauses are conditional the
way the banner's own lines are — the `bash` path exception and the paragraph naming
irreversible commands appear only under `-bash`, because explaining a tool the model was not
given teaches it to skim the rest. These are facts the model cannot recover from any tool schema.
Advice about *which* tool to reach for is deliberately not here; it lives in the tool
descriptions, where it travels with the tool that needs it and costs nothing when that tool is
not mounted.

The persona above it (`defaultSystem`, `apps/cli/main.go`) is the one layer `-system-prompt` replaces
outright, and it stays deliberately short. Codex ships two prompts for one harness — 6.6&nbsp;KB
for the models post-trained on it, 24&nbsp;KB for general GPT-5 — so prompt size mostly measures
how much the model was *not* trained on your harness, and Claude is well inside the trained case
for this shape of tool. What it carries is only what a terminal changes about answering: brevity,
paths rather than pasted source, and not claiming something works unchecked.

**Your own steering text** (`-additional-prompt`, `session.additional_prompt`) is appended last
of all, after the skills catalog. It is the one prompt setting that composes rather than replaces:
in a transformer prompt later text carries more weight, so a directive placed at the end exerts
the most influence over the guidance it sits on — and it mirrors how `CLAUDE.md` and `AGENTS.md`
already work, appended at the end as user instructions layered over the base. Use it for a persona
that must steer every turn, which a skill cannot do: skills use progressive disclosure, so only
their name and description reach the prompt and the body is read on demand. Leaving it empty is
the ordinary case.

**Project and global context** (`-project-context`, `context.go`). Every `CLAUDE.md` and
`AGENTS.md` found walking up from `-root` to the filesystem root, plus `~/.agents/AGENTS.md` —
the [AGENTS.md standard](https://agentsstandard.com)'s own global-base path, also read by
Codex, Cursor, Copilot, Gemini and pi (at its own equivalent, `~/.pi/agent/AGENTS.md`) — all
appended to the system prompt, most general first. A user's own `~/.claude/CLAUDE.md` is
deliberately **not** read: it assumes Claude Code's own tools, hooks and slash commands, where
an `AGENTS.md` at the standard's own path is written, by the convention's premise, to make
sense to whichever agent finds it. None of this is trust-gated — see
[architecture.md](architecture.md) for why a plain instruction file and a skill are not the
same risk.

**Skills** (`-skills`, `-trust-skills`, `-skill-dir`, `skills.go`), following the
[Agent Skills specification](https://agentskills.io/specification): every `SKILL.md` under
`~/.agents/skills/` (no trust needed, the user's own machine), every one under a **trusted**
`.agents/skills/` found the same way the context walk works, and every one under a directory
named by `-skill-dir` (repeatable), `BULLE_SKILL_DIRS` (colon-separated) or `skill_dirs` in
`~/.config/bulle/config.yml`. Only a skill's `name` and `description` ever reach the system prompt; the
model reads the rest with `read_file` once it decides a skill applies. `-trust-skills` trusts
every project-local `.agents/skills/` found on that run and remembers the decision in
`~/.local/state/bulle/trust.json`, keyed by canonical directory — run it once per project, not on every
launch.

**MCP servers** (`mcp.go`). A server is written inline under `sources.mcp:` in `~/.config/bulle/config.yml` — nothing
else is needed to start one:
```yaml
sources:
  mcp:
    mycelium:
      command: mycelium
      args: [mcp]
```
The key is the server name;the definition uses the keys the `mcpServers` format spells — `command` + `args`
(stdio), `"type": "http"` + `url` (remote), `env`, `cwd`, `headers`, `disabled`. Both stdio and HTTP
servers work, on **either backend**:the tools are bridged to `nacelle.Tool`, so
`-approve-tools` gates them like every other tool and `-backend openrouter` gets them too. `${VAR}`和
`${VAR:-default}` expand,, so a token stays in the environment rather than in `~/.config/bulle/config.yml`. Writing
the definition here means there is no second file to keep in step and no pointer to remember to point tor

For the times a server is already configured for another client, `-mcp <file>` (repeatable) names
a `.mcp.json` in the `mcpServers` format and merges its servers in — by server name,, with the flag's
servers winning over the inline ones,, so `-mcp ~/.claude/.mcp.json` reads what Claude Code already
has without copying it. Replacing would mean naming one server silently switching off nine already
configured; merging is how a personal inline list plus another client's file stack the way every client
in this ecosystem layers its own scopes.

A server that will not start **ends the run**, which is the opposite of how skills and project
context fail. Those are discovered, so finding nothing is indistinguishable from there being
nothing to find; this was asked for by name, and a tool that is quietly missing reads as a model
refusing to work rather than as a server that is down.

Nothing is discovered. A `.mcp.json` sitting in the working directory is **not** read, because it
names executables to run — strictly worse than the project-local `.agents/skills/` this client
already gates behind `~/.local/state/bulle/trust.json`, since a skill is text the model may decline to act
on and this is a subprocess started before the model is asked anything. Every server bulle
opens was named on the command line or in your own `~/.config/bulle/config.yml`.

The spec defines the `SKILL.md` file, not where it has to live on disk — every tool picks its
own directory (Claude Code reads `~/.claude/skills/`, pi reads `~/.agents/skills/` and its own
`~/.pi/agent/skills/`), so `~/.agents/skills/` is bulle's own choice, not a location any other
tool already reads. `-skill-dir` is how bulle sees a directory that belongs to one of those,
without moving or copying anything into its own — the same problem pi itself solves with a
`skills` array in `settings.json` pointed at `~/.claude/skills` or `~/.codex/skills`. No trust
gate applies to it, same reasoning as `~/.agents/skills/` above: naming a directory here is
something only the person running bulle, on their own machine, can do in the first place.

**The banner** (`apps/cli/main.go`'s `banner`) is how much of this is actually visible before typing
anything: line one names the backend and model, line two the resolved `-root`, how many skills
loaded (from every source above, combined), and how many `CLAUDE.md`/`AGENTS.md` files did.
Line one also carries the version and, when signed in, the assistant's name.

## Tool approval

`-approve-tools` (`BULLE_APPROVE_TOOLS`, `approve_tools`) asks before every tool call runs.
**Off by default**: a nil `Config.Approve` means the SDK itself runs every call unasked, and
the TUI's own default matches — nobody gets a behaviour change without opting in.

When on, a call blocks on a question above the input, with `y` (allow this one call), `a` (allow this tool
for the rest of the session — remembered in memory only, never written to disk, unlike
`-trust-skills`) or `n` (deny; asked again next time). Asks are serialized to one question at
a time, because a backend can run several tool calls concurrently within a turn and this
client has exactly one place to show a question. A pending question is answered or abandoned
by the same ctrl+c/ctrl+\ that already gets a wedged run's way out — both wait on the run's
own cancellable context, so cancelling one unblocks the other with no separate mechanism.

A denial is reported on `ToolEvent.Refused`, not as a tool failure — the SDK converts it into
a normal tool-result error block for the model to see, the same as any other failed call.

## Deny elevation

`security.deny_elevation` (`BULLE_DENY_ELEVATION`) refuses `bash` calls that try to
elevate privileges: `sudo`, `su`, `doas`, `pkexec`, checked after every chain separator and
inside subshells, not only at the start of the line. **On by default.** The refusal is an
ordinary tool error the model reads, so it adapts instead of retrying.

It is a policy guard against accidents and injected instructions, not a sandbox. A string
check cannot stop a setuid binary that is not sudo, `docker run -u root`, or anything already
running as root; the OS decides who may elevate, and an account outside the `sudo` group is
the actual wall. Set it to `false` to allow elevation. Phase 2 adds a pty password flow, where
sudo prompts on a terminal the harness allocates but never reads, so the agent never sees the
password. [docs/sudo.md](sudo.md) carries the full analysis.

## Matrix

The chat agent runs on the Bulle server, not in the CLI, and needs no setup. The server asks
[Prisme](https://prisme.facile.studio), the suite's chat, for each user's agent account (named
after the assistant), and the agent joins every channel of that user's Prisme spaces. A channel
works in the Bulle space with the same name, or the default space when none matches. In a channel
the agent answers only when it is mentioned, replies in a thread, and keeps one Bulle session per
thread. The web app shows its account, rooms and status under settings, Chat agent.

The operator sets `PRISME_URL` (default `https://prisme.facile.studio`) and `PRISME_SERVICE_KEY`
in the API's environment, plus `BULLE_CREDENTIALS_KEY`, which seals each agent's token. Without
them there is no chat agent. Scheduled jobs also run on the server; manage them in the web app.
See [plan-matrix-spaces.md](plan-matrix-spaces.md).

The CLI no longer has `bulle chat`, `bulle assistant` or `bulle cron`, and a `chat:` or
`assistant:` block in `config.yml` is refused with a message pointing here.

## Slash commands

Typing `/` at the start of a line names one of the client's own commands instead of a
question:

| Command | Does |
|---|---|
| `/clear` | Reset the transcript, the conversation sent to the model, and the running cost total. Same client, new session. Never starts a run. |
| `/cost` | Display cumulative token usage, cost, and tools executed so far. Never starts a run. |
| `/help` | List the commands above and the keybindings (esc, ctrl+c/ctrl+\, ctrl+t, ctrl+g). Never starts a run. |
| `/model [name]` | Switch to another of the server's models, or open the picker. Never starts a run. |
| `/quit` | Quit. Never starts a run. |
| `/resume` | Resume the most recent recorded session for the current project. Never starts a run. |
| `/sessions` | List recorded sessions for the current project with timestamps and models. Never starts a run. |
| `/status` | Show runtime status, context usage, cost, and active configuration. Never starts a run. |
| `/skill:name [what to do]` | Run a loaded skill directly. Starts a run, unlike the seven above. |

An unrecognised command or skill (a typo like `/clera`, or a `/skill:name` that names nothing
loaded) is reported back rather than sent to the model as a literal question — the same
trade-off every peer client with slash commands makes, on the same reasoning: a real question is
far less likely to start with a slash than a mistyped command is.

`/skill:name` sends that skill's own `SKILL.md` content as the question, instead of waiting for
the model to decide on its own that the skill applies and read it with `read_file` — the same
shortcut pi's own `/skill:name` offers, for the same reason: "models don't always" make that
call themselves. Anything typed after the name is appended as `User: <that text>`, the same
convention pi uses, so `/skill:pdf-tools extract the tables` tells the skill what to do with
itself rather than only that it applies. Every loaded skill counts, from `~/.agents/skills/`,
a trusted project `.agents/skills/`, or a `-skill-dir` — one namespace, `skill:`, so a skill
named `clear` or `help` can never collide with the three commands above.

Typing `/` alone opens a dropdown listing every command and skill, each skill with its own
description; more typed narrows it, ranked best match first — a prefix beats a plain substring
beats a fuzzy, order-preserving scatter of the same letters, so `/skill:rev` finds `review`,
`facile-review` **and** `hunk-review`, not only a name that starts with what was typed. While
it's open, `up`/`down` move the selection instead of scrolling the transcript, `tab`/`enter` fill
the prompt with the highlighted entry (plus a trailing space, and without submitting — most
useful with `/skill:name`'s own trailing argument) and `esc` closes it, leaving what was typed
alone. A second, ordinary `enter` is what actually sends it.

### The prompt

The prompt wraps and grows as you type, up to ten rows, then scrolls inside itself. Every row it
gains is a row the transcript gives up, and gets back the moment the question is sent. `alt+enter`
(or `ctrl+j`) starts a new line without sending; `enter` always sends.

It used to be a single-line input, which scrolled sideways instead of wrapping — a question wider
than the terminal slid out of view a character at a time and read as though it were typing over
itself.

`up`/`down` belong to the prompt at every height — there is no transcript here to scroll, so
nothing competes for them and a wrapped question is editable whatever size it has grown to.

A paste of 10 or more lines, or over 1000 characters, goes in as one placeholder,
`[Pasted text #N +X lines]`. Chunks the terminal splits a paste into merge back into one
placeholder. Sending the question, or opening it in the external editor, expands every
placeholder, so the model and the history get the full text. The transcript echoes a long
question as its first three lines plus `… +N lines`.

### Scrolling

The client renders inline, on the terminal's own screen, so scrolling is the
terminal's job — wheel, scrollbar, `tmux` copy-mode, your terminal's search, all of it, on the
whole session rather than on a window this client owns. There are no scroll keys here because
there is nothing here to scroll.

That is a deliberate reversal. It ran on the alternate screen once, which hands a program a
blank page with no scrollback of its own, so the client had to own scrolling: capture the wheel,
which takes click-drag selection with it, leave copy-mode reaching nothing, and un-draw the whole
session on quit. Those arrived as four separate complaints and were one decision. Giving the page
back answers all of them, and costs one thing — a printed line belongs to the terminal, so a
resize reflows it the terminal's way rather than this client's, and a theme change applies from
then on rather than retroactively.

### While a run is going

The input bar follows grok-cli's layout: a flat panel with the label `Bulle` on the left, and one
line under it with the model, the context left and the session's spend on the left and the keys
that work right now on the right. Nothing is drawn above the input while ready or running; a
question that needs you (an approval, a stop in progress, a run that ended short) is the only
line that appears there.

While a run is going the label turns into a three-box loader, and the line under the input names
what the run is waiting on, `waiting for a response` before the model answers, `running read_file`
between a tool's call and its result, `running 3 tools` when the backend runs a turn's tools at
once, next to `enter queue` and `esc interrupt`. It stays visible while scrolled back through
the transcript.

This replaced a spinner that only covered the gap before the first event. Every gap after it — a
tool running, the model called again with its result — left the screen still, and a client that has
stopped moving reads exactly like one that has stopped working.

**Enter still works while a run is going**: the line is queued, shown above the prompt, and sent
once the run finishes. A queued `/help` is still a command.

**`esc` stops the run**, and that is all it ever does. `ctrl+c` stops one too, but it is also the
key that quits an idle client, so the press that abandons an answer is a press that has to be
thought about first; `esc` is the one that never is. Either way the queue goes with the run rather
than being started by it, and the client says how many it dropped — the alternative is a message
that silently never gets asked. Idle, `esc` does nothing at all, and with the `/` dropdown open it
closes that first: one visible thing per press.

The question is on the screen before either of them is worth pressing. It used to appear only when
the answer did: the client printed what it had said only after running what that message started,
and what a question starts is a wait on the model. The prompt emptied and nothing else happened,
which reads as a client that swallowed it. The `⏺` line naming a tool had the same fault and was
worse for it — it waited on the tool it was announcing, so a six-second command was announced six
seconds late. The CHANGELOG carries the measurements.

### What a tool call reads as

A tool that worked is **one line**, carrying its own duration:

```
⏺ read_file(view.go) · 12ms
⏺ bash(go test ./...) · 2.34s
```

The name, the single argument that says which thing it acted on, and how long it took. The
argument is picked by key (`path`, `file_path`, `file`, `command`, `pattern`, `query`, `url`,
`name`), falling back to the sole key when a call has exactly one and to `name()` when nothing
there is worth printing. That beats a table of tool-name to argument-name, because this client
does not know which tools an agent was built with and half of them arrive over MCP under a
server's own names.

The line is held until the result arrives, which is the only way the duration can be on it: a
printed line belongs to the terminal's scrollback and can never be rewritten. Nothing is hidden
while it waits, because the line under the input names the running tool for exactly as long as the line is
held. A run that ends with calls still in flight (the iteration limit, or `esc`) still prints
them, without a duration.

**A failure stays two lines and stays loud**, because it is the thing a reader must not scroll
past:

```
⏺ edit_file(transcript.go)
  ⤷ edit_file failed after 3ms: no such file
```

What this replaced printed the raw JSON input and then a second `done in 12ms` line under it,
so every transcript was twice as tall in order to report that nothing had gone wrong.

**A file edit also draws its diff.** When `edit_file` or `write_file` changes a file, the
one-line report is followed by what changed, git-style — removals in the terminal's red,
additions in its green, with three unchanged lines around each block so the eye can find where
in the file it is looking:

```
⏺ edit_file(toolline.go) · 4ms
   1   func (m *model) finished(tool *nacelle.ToolEvent) {
   2 -   line := m.run.running[tool.ID]
   2 +   line, held := m.run.running[tool.ID]
   3       delete(m.run.running, tool.ID)
```

Line numbers run down the left gutter — green on additions, red on removals, muted on context.
Added and removed lines each sit on a faint wash (the pure green or red blended to ~20% onto the
pane's backdrop) rather than a solid block, so a long change stays readable without shouting.
The diff is cut to the window width so nothing wraps in scrollback, and capped at four hundred
lines with a `… more` marker saying it was cut rather than pretending it wasn't. A wholesale
rewrite too large to align line by line reads as one block of removals against additions.

`edit_file` carries both sides of the change in its own input, so nothing is read from disk to
draw it; `write_file` carries only the new contents, so what the file held when the call was seen
becomes the before side — empty for a file being created, which shows as all additions. Only
these two tools get a diff: search, commands and MCP tools render exactly as they did before.
Turn it off with `diffs: false` (`BULLE_DIFFS`, `-diffs`) to restore the bare one-line report.

**Input with a repeated key is refused rather than summarised.** `encoding/json` keeps the last
value of a duplicate key silently, so `{"command":"ls","command":"rm -rf /"}` can be shown as
`bash(ls)` over a call that ran something else. The line says `input has a duplicate key`
instead, and with `-approve-tools` on the gate denies the call outright without ever drawing a
prompt: an approval prompt that offers `y` for a call nobody can render is the same lie with a
warning label.

### Reasoning

With `thinking: true`, a turn's reasoning collapses to one quiet line:

```
· thought for 4.2s · ctrl+t to expand
```

The hint appears once per session, not under every answer. **`ctrl+t` expands**: it prints the
retained reasoning below, and sticks, so later turns print theirs in full until it is pressed
again. Expanding cannot replace the collapsed line already on screen for the same reason the
tool line carries its own duration, so a second copy underneath is what expanding means here.
Only the most recent turn's reasoning is retained. `ctrl+t` takes the press whatever is on
screen, which costs the prompt its `transpose-character-backward` binding.

### Quitting

`ctrl+c` when nothing is running, `ctrl+\` at any time, and `/quit` all end the session, and the
session simply stays where it is. Nothing is printed on the way out and nothing needs to be: what
was said went into the terminal's scrollback as it happened, so quitting leaves it exactly as
running left it.

That used to need a deliberate dump at exit, because the alternate screen hands the old page back
untouched and quitting un-drew the whole conversation. Inline rendering removed the problem rather
than the workaround.

## Sandboxes & remote hosts

bulle always runs **on the host**, with the session's tools executing in the
target over SSH. The model reaches the target's workspace, but no provider key,
session recording or host file ever crosses the boundary — the tools are the
only thing that does. Two groups configure the two kinds of target:

```yaml
sandbox:                          # local boite microVMs, used by 'bulle sandbox'
  default: ""                     # VM entered when 'bulle sandbox' has no argument
  user: boite                     # default SSH user
  port: 2226                      # default port
  ssh_key_path: ~/.ssh/id_ed25519 # default identity file
  root: ""                        # empty uses the VM's own workspace
  auto_snapshot: false            # snapshot the overlay disk on exit
  targets:
    dev-vm:
      vm_name: dev-box            # boite instance behind the key (default: the key)
      port: 2226
      user: boite
      ssh_key_path: ~/.ssh/id_ed25519
      root: ""
      auto_snapshot: false

remote:                           # SSH hosts, used by 'bulle remote'
  default: ""                     # host entered when 'bulle remote' has no argument
  user: ""                        # default SSH user (ssh's own config wins when empty)
  port: 0                         # 0 lets ssh choose: ssh_config's Port, then ssh's default (22)
  ssh_key_path: ""                # empty lets ssh choose: ssh_config IdentityFile, then the agent
  root: ""                        # empty starts in the SSH login directory (home)
  targets:
    staging:
      host: staging.example.com
      port: 22
      user: deploy
      ssh_key_path: ""
      root: /srv/app
```

Host-key verification is deliberately not relaxed for `remote`: a host on the
network checks against your own `known_hosts` and `~/.ssh/config`, exactly as a
hand-typed `ssh` would, and `BatchMode=yes` refuses an unknown host rather than
trusting it. Because `BatchMode` also removes the yes/no prompt a hand-typed
ssh shows, the first connection to an unknown host fails; that failure names the
host and the `ssh-keyscan` command that trusts it, so the step stays explicit
rather than automatic. Boite microVMs are the exception — their key is generated
per instance, never in `known_hosts`, and the connection is loopback — so only
the `sandbox` path turns host-key checking off. A `remote` target that happens
to be a local boite VM is better run through `bulle sandbox`, which trusts the
VM's key without touching `known_hosts`.

Commands:
- `bulle sandbox list`: lists configured `sandbox.targets` and local boite VMs.
- `bulle sandbox <vm> [prompt]`: interactive or headless session inside a boite VM.
- `bulle sandbox <vm> --snapshot`: snapshot the VM's overlay disk when the session ends.
- `bulle remote list`: lists the hosts configured under `remote.targets`.
- `bulle remote <host> [prompt]`: interactive or headless session on an SSH host.

`remote` accepts a `remote.targets` key, a direct `user@host:port` address, or
an `~/.ssh/config` host alias — ssh itself resolves the last two. Everything
else (port, user, identity, workspace) comes from the `remote` group unless the
target or a flag overrides it. Overrides common to both commands:
`--workdir/-w`, `--user/-u`, `--print/-p`; `remote` adds `--port` and `--key`.

`sandbox.targets` is boite-only. An SSH host belongs under `remote.targets`;
a key there that names a `backend:` or `host:` is refused at load, because a
strict decoder only accepts the fields each group actually has.

Preflight checks who the target answered as, and refuses a mismatch. It does not
refuse `root`: `bulle remote root@host` works, because connecting as root to a
machine you own is your call. A boite VM still resolves to `boite` unless you
override `--user`, so an instance that unexpectedly lands you as root is caught
by the user check rather than by a name.


