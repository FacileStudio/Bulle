# Bulle web UI plan

Status: DRAFT, awaiting review. Not executed. Cold-start handoff: a fresh worker needs no prior conversation.

Date: 2026-11-01. Supersedes `~/Code/Facile/bulle-mono/ARCHITECTURE_PLAN.md` (Tauri, Capacitor and Mycelium session sync are out of scope here).

## Goal

Serve Bulle at `https://bulle.facile.studio` as a suite web app (Go API + SvelteKit client in one container). Each Registre user signs in by SSO only and gets one persistent boite VM as the agent's own computer. The agent loop runs on the host and executes every tool call inside that VM. The web is a new surface on the same assistant brain that Matrix and cron use today.

## Why (evidence)

Bulle is terminal-only (`cmd/`, `internal/tui`). Its assistant brain lives in files on one host (`~/.bulle`, `cmd/chat_room_store.go`) and is reachable only through Matrix. `bulle sandbox` already runs a host-side agent loop with tools over SSH in a boite VM (`cmd/sandbox.go`, `internal/sandbox/tools*.go`), so the "agent has its own computer" model exists, minus a server, auth and a browser.

## Decisions taken (2026-11-01)

| Topic | Decision |
|---|---|
| Repo | Rename GitHub repo `bulle` to `Bulle`. Monorepo: `core/` shared Go module, `apps/cli`, `apps/api`, `apps/client`, tied by `go.work` |
| Agent loop | Runs in the Bulle API process on the host. Tools run in the VM over SSH. Provider keys and credentials never enter the VM |
| Computer | One persistent boite VM per user, created on first use, stopped when idle, started on demand |
| Tenancy | Registre team users. Provider keys are server-side, in casier |
| Env | casier for every env var, dev and prod (`casier.yml`, `casier run`) |
| Auth | porte, `SSO_ONLY=true`, issuer `https://sso.facile.studio` |
| UI | muse for the app shell. OpenUI (`@openuidev/svelte-lang`) only inside agent replies |
| v1 surfaces | Chat + sessions, background jobs. No approvals inbox, no computer view |
| Assistant | Same brain: memory, room history and jobs move to Postgres; Matrix and web are both channels of it |
| Boite | Build `boite-worker` (Track A of `boite/boite-web-plan.md`) first; Bulle calls its Unix socket |
| Approvals | The VM is the boundary: VM tools auto-run. Tools that act outside the VM (mail, Matrix, connectors) show inline approve/deny cards in chat |
| Domain | The app takes `bulle.facile.studio`. The landing content moves into the SPA's logged-out root; the `gare.yml` static deploy goes |

## Research: practices adopted

Each line names the practice and where it lands in this plan. Sources are listed at the end.

- **Secrets stay outside the sandbox.** Meta Muse keeps credential storage, the state database and connectors outside the agent's container and talks to them over Unix sockets; the agent only holds surrogate tokens. Here: provider keys and connector tokens live in the API process; the VM receives tool calls and returns output, nothing else. The worker is reached over a Unix socket.
- **Approval lives outside the agent's workspace.** "A model-written preference is not the same thing as an enforced policy" (developersdigest on dots). Here: the approval gate sits in the API, keyed on tool class, not in the prompt.
- **Gate by risk class.** Muse's Sentinel auto-allows clean low-risk actions and asks for tainted or outward ones. Here: VM-local tools auto-run; outward tools ask.
- **Snapshots for rollback and incident review; CPU, memory, disk and time limits.** Here: boite `snapshot`/`rollback` exposed on the worker; per-run timeout and per-VM resource caps.
- **Runs continue unattended and leave receipts.** Dots and Muse keep working after the app closes. Here: runs are server-side; every event is persisted, so a closed tab loses nothing and job runs keep a full log.
- **Event protocol for agent UIs.** AG-UI defines lifecycle, text, tool-call and state events over SSE, with user input on plain HTTP and an interrupt event for approvals. Here: the SSE stream uses AG-UI event names and shapes (`RUN_STARTED`, `TEXT_MESSAGE_CONTENT`, `TOOL_CALL_START`, ...) without adopting the SDK, plus `Last-Event-ID` replay from Postgres for reconnects.
- **Generative UI from a whitelisted component library.** OpenUI's model emits OpenUI Lang, which renders only components you register; it streams line by line. Here: the library is built from muse components; unknown components or parse failures fall back to markdown.

## Approach

Restructure first, without behavior change: the CLI moves to `apps/cli`, and the packages both apps need (`agent`, `engine`, `sandbox`, `settings`, `compaction`, `memory`, `assistant`, `chat` stores) move to `core/`. The API is forked from `GoSvelteBoilerplate` (`scripts/rename.sh`), the same shape as Sonde: tronc chassis, porte, GORM over Postgres, goose migrations, one distroless image serving `/api` and the SPA. A `runner` package inside the API owns live runs: it builds a nacelle agent with `core/agent`, points VM tools at the user's box through `core/sandbox`, writes each event to `run_events`, and fans events out to SSE subscribers. The seam that makes "same brain" possible is a store interface in `core/` with a file implementation (CLI, today's behavior) and a Postgres implementation (API). The Matrix daemon and cron become channels that read and write through the API instead of `~/.bulle`.

```text
browser ──HTTPS──> Traefik ──> bulle container (distroless)
                                 ├─ /api/*  chi + porte + runner + scheduler
                                 └─ /*      SvelteKit SPA (muse + OpenUI)
                                      │ Unix socket (token)        │ Postgres
                                      v                            v
                         host: boite-worker (systemd, root, /dev/kvm)
                                      │ QEMU, ssh bound to docker bridge only
                                      v
                         bulle-<user> VM  /workspace, tools run here
```

## Phases and steps

Phases run in order unless marked parallel. Every step names its files and its exit criterion.

### Phase 0: rename and restructure (no behavior change)

1. GitHub: `gh repo rename Bulle -R FacileStudio/bulle`. GitHub redirects the old URL. Rename the local folder to `~/Code/Facile/Bulle`. [exit: `git remote -v` shows `FacileStudio/Bulle`]
2. `apps/cli/` (move): today's `main.go`, `cmd/`, `internal/tui`, `internal/toolview`, `internal/layout`, `internal/menu`, `internal/theme`, `internal/ide`, `internal/herdr`, `main_test.go`. Module `github.com/FacileStudio/Bulle/apps/cli`. [module-path] [exit: `go build ./...` in `apps/cli` produces the same `bulle` binary]
3. `core/` (move): every `internal/*` package the API also needs, as public packages (Go's `internal/` rule would block `apps/api`). Module `github.com/FacileStudio/Bulle/core`. Start with `agent`, `engine`, `sandbox`, `settings`, `sessions`, `compaction`, `memory`, `assistant`, `chat`, `skills`, `cost`, `usage`, `approval`, `provider`. Anything only the TUI uses stays in `apps/cli/internal`. [exit: `go vet ./...` clean in both modules]
4. `go.work` at the root listing `core`, `apps/cli`, `apps/api`. [exit: `go work sync` clean]
5. Release plumbing: `.goreleaser.yml` (`dir: apps/cli`), `install.sh`, `.github/workflows/*.yml`, `lefthook.yml`, `filet.yml`, `mise.toml` (switch to `monorepo_root = true` with `config_roots`, as in Sonde), Homebrew tap formula URL, the `facile` installer manifest. Grep the suite for `FacileStudio/bulle` (`bulle.nvim`, `facile`, `homebrew-tap`, wiki) and update each. [exit: a snapshot GoReleaser build produces the binary; `facile install bulle` still works]
6. `README.md`, `docs/development.md`, `CHANGELOG.md`: new layout, new module paths. Note that `go install github.com/FacileStudio/bulle@latest` is replaced by `.../Bulle/apps/cli@latest` and that `go install` needs `apps/cli/vX.Y.Z` tags. [exit: docs match the tree]
7. Gate. [filet] [exit: `filet check` clean at the root, all tests pass, `bulle` TUI, `bulle chat`, `bulle cron` behave as before on ruche]

### Phase 1: boite-worker (boite repo, parallel with Phase 2)

Execute Track A of `~/Code/Facile/boite/boite-web-plan.md` (steps 1 to 3) with these additions Bulle needs:

8. `boite/cmd/worker.go`: also expose `POST /boxes/:id/snapshot`, `POST /boxes/:id/rollback`, `GET /boxes/:id/snapshots` (reuse `cmd/snapshot.go` logic). Accept `owner` and resource caps (`cpus`, `memory`, `disk`) on create. [exit: `curl --unix-socket` creates, snapshots, rolls back a box]
9. Control key: the worker reads one control public key from env (`BOITE_CONTROL_PUBKEY`) and injects it into every box's config ISO at create, beside user keys. Bulle holds the private half (casier). [exit: `ssh -i <control key> -p <port> boite@<bind>` works on a fresh box]
10. `BOITE_SSH_BIND` set to the Docker bridge gateway of the network the Bulle container sits on, never `0.0.0.0` and never public. [exit: SSH reachable from inside the Bulle container, refused from outside the host]
11. Idle stop: `POST /boxes/:id/stop` is enough; idle policy lives in Bulle, not the worker. [exit: n/a, a note in the worker README]
12. Confirm the target host has `/dev/kvm` (`ls -l /dev/kvm`, `kvm-ok`) and enough RAM for N users times the default box size. [exit: written down in `docs/deploy.md` of Bulle]

### Phase 2: API and client skeleton, deployed

13. Fork `GoSvelteBoilerplate` `apps/api`, `apps/client`, `Dockerfile`, `docker-compose.yml`, `scripts/{check,verify-deploy,postgres-up,postgres-down}.sh` into the repo; run `scripts/rename.sh` with slug `bulle` and module `github.com/FacileStudio/Bulle/apps/api`. [module-path] [exit: `mise run build` builds both apps]
14. `casier.yml` (`project.slug: bulle`, `environment: dev`); `mise.toml` dev tasks wrap `casier run -- ...` and keep an `--offline` variant, as Sablier does. `apps/api/.env.example` documents every key with no values. Production values are set in casier for the prod environment. Keys: `DATABASE_URL`, `SSO_ONLY=true`, `OIDC_ISSUER=https://sso.facile.studio`, `OIDC_CLIENT_ID=bulle`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URL=https://bulle.facile.studio/api/auth/oidc/callback`, `OIDC_SUCCESS_URL`, `OIDC_CLI_AUDIENCE=facile-cli`, `ANTHROPIC_API_KEY` (and the other backends), `BOITE_WORKER_SOCK`, `BOITE_WORKER_TOKEN`, `BOITE_CONTROL_KEY`, `JOURNAL_URL`, `JOURNAL_TOKEN`, `ANTENNE_*`. [exit: `mise run dev` boots with no `.env` file]
15. Registre: register client `bulle` with redirect `https://bulle.facile.studio/api/auth/oidc/callback` and `http://localhost:5173/api/auth/oidc/callback` for dev. [auth/porte] [exit: discovery and login round-trip against `https://sso.facile.studio`]
16. `apps/api/main.go`, `modules/auth`: porte kit with `SSO_ONLY=true` (no password routes), device exchange on `OIDC_CLI_AUDIENCE` so the CLI can sign in later. Cookie read before `Authorization`, `X-Facile-CSRF` on every cookie mutation, refuse `email_verified:false`. [auth/porte] [exit: `auth_test.go` covers CSRF refusal and unverified email refusal]
17. `apps/client/src/routes/login/+page.svelte`: one "Sign in with Facile" button; render `?error=` outside the SSO gate (the only failure feedback under `SSO_ONLY`). `(app)/+layout.svelte`: muse shell, sidebar of sessions, user menu. Logged-out `/`: landing content ported from `site/index.html`. [muse] [exit: login works end to end in dev]
18. `docker-compose.yml`: one `bulle` service, no `PathPrefix`, no `stripprefix`, `expose` only. Mount `/run/boite/worker.sock`. The image must reach the worker socket and SSH to the bridge gateway; keep distroless `:nonroot` unless a writable volume is mounted (none planned). Remove `gare.yml` and `site/` once the SPA landing ships. [exit: `scripts/verify-deploy.sh https://bulle.facile.studio` green, SSO login works in production]

### Phase 3: computers (needs Phases 1 and 2)

19. `apps/api/migrations/00002_computers.sql`: `computers(id, user_id unique, box_id, status, ssh_host, ssh_port, cpus, memory_mb, disk_gb, last_active_at, created_at)`. [migrations] [exit: `goose` applies on a fresh database; failed migration exits 1]
20. `apps/api/internal/worker/`: typed client for the worker socket, errors mapped to the suite error envelope, reconcile when Postgres and the worker disagree. [exit: tests against a stub socket server]
21. `apps/api/internal/computers/`: `Ensure(user)` creates the box on first use, starts a stopped one, waits for SSH, runs `core/sandbox.PreflightCheck`. Idle sweeper stops boxes after `BULLE_IDLE_MINUTES` (default 30) with no run. [exit: first chat of a new user boots a VM; it stops after the idle window; the next chat starts it again]
22. `apps/api/modules/computers/`: `GET /api/computer` (status, resources), `POST /api/computer/{start,stop}`, `POST /api/computer/snapshots`, `POST /api/computer/rollback`. Snapshot and rollback stop the box first, since qcow2 snapshots are taken offline. [auth/porte] [exit: curl through porte works; status shows in the client header]

### Phase 4: chat and sessions (needs Phase 3)

23. `core/` store seam: define `SessionStore`, `MemoryStore`, `RoomStore`, `JobStore` interfaces next to the current file code. The CLI keeps the file implementations. [exit: CLI tests unchanged and green]
24. `apps/api/migrations/00003_sessions.sql`: `sessions(id, user_id, channel web|matrix|cron, title, backend, model, created_at, updated_at)`, `messages(session_id, seq, payload jsonb)` holding the nacelle transcript, `runs(id, session_id, status, stop_reason, usage jsonb, cost_usd, started_at, finished_at)`, `run_events(run_id, seq, type, payload jsonb)`, `approvals(run_id, tool_call_id, status, decided_by, decided_at)`. Postgres implementations of the stores in `apps/api/internal/store/`. [migrations] [exit: round-trip test on `tronc/testdb`]
25. `apps/api/internal/runner/`: one goroutine per run. Builds the agent with `core/agent` and the user's settings profile, binds VM tools (`core/sandbox` command, file, edit, glob, search) to the user's computer, keeps host tools (web fetch, memory, connectors) on the host. Maps `core/engine` events to AG-UI event types, appends them to `run_events`, and publishes to subscribers. One active run per session, a per-user concurrency cap, a per-run timeout. On SIGTERM: stop accepting runs, wait a grace period, mark the rest `interrupted` so the UI offers "continue". [exit: unit test drives a fake backend through a full turn and asserts the persisted event sequence]
26. Approval gate in the runner: tools tagged `outward` pause the run, emit an `INTERRUPT`-style event, and resume on `POST /api/runs/:id/approvals/:call` (approve or deny). VM tools carry no gate. [exit: a test proves an outward call does not execute before approval and does execute after]
27. `apps/api/modules/sessions/`: `GET/POST /api/sessions`, `GET /api/sessions/:id`, `POST /api/sessions/:id/messages` (starts a run), `POST /api/runs/:id/stop`, `GET /api/runs/:id/events` as SSE with `Last-Event-ID` replay from `run_events`. [auth/porte] [exit: kill the connection mid-run, reconnect, receive the missing events once, in order]
28. `apps/client/src/routes/(app)/s/[id]/`: streamed chat, collapsed thinking (as in the TUI), tool calls grouped by kind, approval cards, stop button, token and cost footer. `$lib/backend.ts` stays relative `/api`; `$lib/stream.svelte.ts` wraps `EventSource` with runes. [muse] [exit: close the tab mid-run, reopen, see the run finished with the full log]

### Phase 5: OpenUI in replies (needs Phase 4)

29. `apps/client`: add `@openuidev/svelte-lang` (MIT, Svelte 5 peer). `$lib/openui/library.ts` registers muse-backed components with `defineComponent` and `createLibrary`: `Table`, `Card`, `Stat`, `List`, `Form`, `Field`, `Button`, `Badge`, `Chart`, `Callout`. Nothing else renders. [muse] [exit: a fixture of OpenUI Lang renders with muse tokens in light and dark]
30. API: generate the library's prompt section once at build time and add it to the system prompt for the web channel only (Matrix and cron keep plain text). Fence OpenUI blocks in replies so the client knows what to parse. [exit: the model produces a table reply that renders as a muse table]
31. Client: render OpenUI blocks while they stream; on a parse failure show the raw block as markdown. Button and form actions post back as a structured user message. [exit: a test feeds broken OpenUI Lang and gets markdown, not a crash]

### Phase 6: background jobs (needs Phase 4)

32. `apps/api/migrations/00004_jobs.sql`: `jobs(id, user_id, name, schedule, prompt, profile, delivery, enabled)`, `job_runs(job_id, run_id)`. [migrations]
33. `apps/api/internal/scheduler/`: an in-process cron over `jobs`, one tick per minute, a Postgres advisory lock so two replicas never double-fire. Each fire starts a normal run on channel `cron`, so it lands in the same event log. Keep the CLI's unattended defaults (shell off unless the job enables it, disabled until a test run). Delivery targets: `web` (notification in the app), `matrix:<room>`. [exit: a job set to every minute produces runs with logs; disabling it stops them]
34. `apps/client/src/routes/(app)/jobs/`: list, create, edit, test-run, enable, and a run history with the full log. [muse] [exit: create a job in the UI, test-run it, enable it, see the next scheduled run]

### Phase 7: same brain (needs Phases 4 and 6)

35. `apps/cli` command `bulle assistant migrate`: reads `~/.bulle` (memory, room history, jobs, profile) and uploads it to the API as the signed-in user, through porte device exchange (`OIDC_CLI_AUDIENCE`). Idempotent. [exit: run twice on ruche, data appears once in the web UI]
36. Matrix channel: the chat daemon switches from file stores to API-backed stores (the `core/` interfaces with an HTTP implementation), or moves into the API process as a channel adapter. Pick in review; the API-adapter option removes a process. [exit: a message in Matrix shows up in the web session list, and the reply in either surface uses the same memory]
37. Retire crontab-installed assistant jobs (`bulle assistant install`) in favor of the scheduler; keep the CLI path for local, non-assistant cron. [exit: `crontab -l` on ruche has no Bulle assistant entries; pulse and check-in run from the scheduler]
38. Connectors (courrier, agenda, opus): run on the host side with per-user tokens obtained through each app's porte device exchange, never inside the VM. Every sending action is `outward`. [auth/porte] [exit: the assistant reads inbox and calendar for the signed-in user only]

## Files to modify or create

- Moved: `main.go`, `cmd/**`, TUI-only `internal/**` to `apps/cli/`; shared `internal/**` to `core/`.
- Modified: `.goreleaser.yml`, `install.sh`, `.github/workflows/*.yml`, `mise.toml`, `filet.yml`, `lefthook.yml`, `README.md`, `docs/development.md`, `docs/configuration.md`, `CHANGELOG.md`.
- New: `go.work`, `casier.yml`, `Dockerfile`, `docker-compose.yml`, `scripts/*.sh` from the boilerplate, `apps/api/**` (`main.go`, `router.go`, `migrations/0000{2..4}_*.sql`, `internal/{worker,computers,runner,scheduler,store}`, `modules/{auth,users,computers,sessions,jobs}`), `apps/client/**` (`routes/login`, `routes/(app)/{s/[id],jobs}`, `lib/{backend.ts,stream.svelte.ts,openui/}`), `docs/deploy.md`.
- Removed after Phase 2: `gare.yml`, `site/`.
- Other repos: `boite` (`cmd/worker.go`, `cmd/qemu/instance.go`, `scripts/boite-worker.service`), Registre (client `bulle`), `homebrew-tap`, `facile`, `bulle.nvim`.

## Exit criteria (whole plan)

- `https://bulle.facile.studio` serves the app; only Registre SSO signs in; `scripts/verify-deploy.sh` is green.
- A new user's first message boots their VM; tools run inside it; `ps` on the VM shows no provider key, and no key appears in any VM file.
- Closing the tab mid-run loses nothing; reconnecting replays missed events once.
- Outward tools never run before approval.
- Jobs run on schedule with full logs.
- The same memory answers in Matrix and on the web.
- `filet check` clean at the root; `scripts/check.sh` green; the CLI behaves as before.
- A web redeploy does not stop any VM.

## Risks

- **Rename breaks Go import paths.** `go install` users and any repo importing `github.com/FacileStudio/bulle` break. Grep the suite in step 5. Go module tags in subdirectories need the `apps/cli/` prefix.
- **Runs die on redeploy.** The runner lives in the API process, so a redeploy interrupts live runs. Step 25 marks them `interrupted` and resumable. If that hurts in practice, split the runner into its own long-lived process (same binary, `bulle-api runner`), like the boite worker.
- **Docker to host SSH.** The container must reach the bridge gateway port; Dokploy's network layout decides the address. Verify on the real host in step 10 before Phase 3.
- **Host capacity.** One VM per user costs RAM even when idle stop works. Size the default box and the idle window from step 12's numbers.
- **Snapshots are offline.** qcow2 snapshots need a stopped VM, so no per-run automatic snapshot in v1.
- **VM egress is open.** slirp networking lets the VM reach the internet. Acceptable while the VM holds no secrets; revisit before connectors or personal data ever enter the VM.
- **Brain migration with two writers.** During Phase 7 the Matrix daemon and the web must not both write files and Postgres. Cut over in one step, after `migrate`.
- **OpenUI churn.** `@openuidev/svelte-lang` is 0.4.x. Pin the exact version; the markdown fallback keeps replies readable if it breaks.
- **Cross-repo order.** Phase 3 depends on the boite worker contract; pin boite by tag when it stabilizes. [distribute]

## Skip (YAGNI)

- Approvals inbox, computer view (web terminal, file browser, take over): not v1, by decision.
- Tauri, Capacitor, mobile apps, Mycelium session sync from the old mono plan.
- Per-user provider keys, billing, quotas beyond a concurrency cap and a per-run timeout.
- Muse-style egress filtering (Sentinel, eBPF taint): revisit when the VM handles personal data.
- Whole-app generative UI: OpenUI stays inside replies.
- `porte/spaces` or teams sharing a computer.
- The boite-web dashboard: Bulle manages its own VMs through the worker.

## Convention flags checked

Checked against: `[module-path]` (`github.com/FacileStudio/Bulle/apps/api`, matches Sonde, Capsule, Opus casing), `[auth/porte]` (`SSO_ONLY`, issuer `https://sso.facile.studio`, `/api` in redirect URL, `?error` outside the SSO gate, CSRF header, `email_verified`, device exchange from Sonde's `.env.example`), `[migrations]` (goose sequential, embedded, exit 1 on failure), `[muse]` (tokens, Svelte 5 runes, mobile first), `[filet]`, `[distribute]` (boite worker by tag), casier (`casier.yml` + `casier run` as in Sablier), deployment shape (one container, one router, no stripprefix, from Sonde's `CLAUDE.md`). Not applicable in v1: `[events]` beyond the existing Antenne publishing.

## Sources

- OpenUI home: https://www.openui.com/
- OpenUI repository: https://github.com/thesysdev/openui
- OpenUI Svelte renderer issue #302: https://github.com/thesysdev/openui/issues/302
- `@openuidev/svelte-lang` 0.4.2, MIT, peer `svelte >=5.0.0`: direct observation via `npm view`
- Meta Muse architecture (sandbox, surrogate tokens, Sentinel, Unix sockets; single paywalled source, unverified): https://www.uncoveralpha.com/p/meta-just-launched-its-most-important
- Meta Muse launch coverage: https://axios.com/2026/09/08/meta-debuts-muse-personal-ai-agent
- Meta Muse isolation statement: https://www.ibtimes.com/meta-launches-muse-delve-ai-agent-market-its-stock-climbing-3807271
- OpenAI dots overview (third party): https://www.eesel.ai/blog/openai-dots
- Dots control-plane lessons: https://www.developersdigest.tech/blog/openai-dots-agent-workflow
- Agent sandbox architecture guide: https://developersdigest.tech/blog/agent-sandbox-architecture-guide
- AG-UI concepts: https://docs.ag-ui.com/concepts
- AG-UI architecture: https://docs.ag-ui.com/concepts/architecture
- Not reached: OpenAI's own dots announcement and Meta's own Muse announcement; claims from them above are secondary.
