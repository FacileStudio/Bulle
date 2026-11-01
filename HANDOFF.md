# Bulle web UI: handoff

Written 2026-11-01. A cold-start brief for the next agent. Read this file, then `docs/plan-web-ui.md` (the original plan, partly out of date, see "Plan deviations") and `docs/deploy.md` (the deploy runbook).

## Goal

Ship `https://bulle.facile.studio`: a suite web app (Go API + SvelteKit client, one container). Registre SSO only. Each user gets one persistent boite VM as the agent's computer. The agent loop runs on the host and runs every tool call in that VM over SSH. Provider keys never enter the VM. Assistant replies can render OpenUI blocks.

Release 1 scope: sign in, a per-user computer, chat with streamed tool calls, OpenUI in replies. Out of scope for release 1: background jobs and moving the assistant brain into Postgres (see "Later phases").

## Rules from the user (global CLAUDE.md, still apply)

- No inline comments and no comments in function bodies, in any language.
- Conventional commits: `type(scope): summary`, imperative, lowercase, under 72 chars. No co-author lines, no "Generated with" lines. The `commit-msg` hook adds the gitmoji.
- `filet check` must pass before any commit that ships. Known noise: libolm lsp errors in the old CLI path can make it exit 1; compare to baseline.
- No em dashes, no emoji in committed prose. Bun for JS, mise when available.
- Check `mycelium flow list` before typing a multi-step shell sequence. Wiki lives in `~/.mycelium/memory`. A short entry exists in `projects/bulle.md`; update it when you finish.
- Ask before outward actions that were not requested. The user explicitly asked for: repo rename (done), release and deploy of a usable version. Pushing Bulle `main` is covered; creating accounts or secrets on the user's behalf is not.

## Current state

### Done and released

- **boite v0.10.0** is tagged and released (`FacileStudio/boite`). It adds `boite-worker`, a host daemon on a Unix socket. Release archives include `boite-worker_0.10.0_linux_{amd64,arm64}.tar.gz`. `main` is pushed.
  - Async create: `POST /boxes` returns 202 with status `creating`, then `running` or `error`. Creates run one at a time.
  - Socket access: `BOITE_WORKER_SOCK_MODE` (default 0600) and `BOITE_WORKER_SOCK_GROUP`.
  - Other env: `BOITE_WORKER_TOKEN` (required), `BOITE_WORKER_SOCK`, `BOITE_SSH_BIND`, `BOITE_WORKER_HOST`, `BOITE_CONTROL_PUBKEY`, `BOITE_DIR`.
- **GitHub repo renamed** to `FacileStudio/Bulle`. Local folder is `~/Code/Facile/Bulle`. `origin` points at the new URL.
- **DNS**: the user created `bulle.facile.studio` (Cloudflare, same addresses as `sonde.facile.studio`). It answers 404 until a Dokploy compose exists.

### Committed on Bulle `main`, NOT pushed

`git log` head is `95542d4`. Nothing from Bulle has been pushed since the rename. In order:

1. Phase 0: `core/` (shared Go module) + `apps/cli` (the old CLI, module `github.com/FacileStudio/Bulle/apps/cli`) + `go.work`.
2. Phase 2 skeleton: `apps/api` (Go, porte SSO-only, tronc, GORM, goose migration `00001_users.sql`) and `apps/client` (SvelteKit 2, Svelte 5, muse). Logo is Iconify `solar:smile-square-bold-duotone`.
3. `apps/api/internal/worker`, `internal/computers`, `modules/computers`, migration `00002_computers.sql`: the per-user computer. Tests pass (the DB test needs `TEST_DATABASE_URL`).
4. Client chat: sessions sidebar, `/s/[id]` streamed chat, computer chip, approval cards, stop button.
5. OpenUI: `apps/client/src/lib/openui/` (pinned `@openuidev/svelte-lang` 0.4.2, zod 4, vitest 5.0.5), wired into `Reply.svelte`.
6. Deploy fixes and `docs/deploy.md`: `go.sum`, `.dockerignore`, compose mounts the socket directory and adds `group_add`.

Verified at the end of the last session: `go build/vet/test` in `apps/api` pass; client `svelte-check` 0 errors, `vitest` 42 passed, `vite build` OK; `scripts/check.sh` passed after the Phase 0 merge; the Docker image built in the deploy agent's run.

### Partial: sessions and runner (the main missing piece)

Branch `worktree-agent-a45a30173daa49ed5`, worktree `.claude/worktrees/agent-a45a30173daa49ed5`, last commit `c2c4ff1 wip(api): partial sessions store and runner, unverified`. The agent was stopped mid-work. It has:

- `apps/api/migrations/00003_sessions.sql`
- `apps/api/internal/store/` (`store.go` interface, `postgres.go`, `memory.go`, `store_test.go`)
- `apps/api/internal/runner/` (`runner.go`, `run.go`, `events.go`, `codec.go`, `gate.go` approval gate, `tools.go`, `runner_test.go`)
- `apps/api/go.mod`, `go.sum`, `go.work.sum` edits

`go build ./...` in `apps/api` printed no errors on that branch. Tests were NOT run by the lead and the agent never reported. It does NOT have the HTTP layer: `apps/api/modules/sessions/` does not exist, and nothing is wired in `main.go` or `router.go`.

First move: merge that branch into `main` (expect `go.mod`/`go.sum` conflicts with the computers branch, resolve with `go mod tidy` in the workspace), run its tests, then finish.

## Remaining work for release 1, in order

1. **Merge and verify the runner branch.** `go test -race ./...` in `apps/api` and `core`. `TEST_DATABASE_URL=postgres://postgres:postgres@127.0.0.1:5573/bulle_test?sslmode=disable` with `mise run db-up` for the store test. Fix what fails.
2. **Write `apps/api/modules/sessions/`**: router, handler, types, documentation (the suite API reference registry; copy the pattern in `modules/computers/`). Exact contract the client already codes against:
   - `GET /api/sessions` returns `[{id,title,updated_at}]`
   - `POST /api/sessions {title?}` returns `{id,title}`
   - `GET /api/sessions/{id}` returns `{session, messages:[{seq,role,content}], active_run_id?}`. It must leave the in-flight assistant message out of `messages`.
   - `DELETE /api/sessions/{id}`
   - `POST /api/sessions/{id}/messages {text}` returns 202 `{run_id}`
   - `GET /api/runs/{id}/events`: `text/event-stream`, `id=<seq>`, `event=<TYPE>`, `data=<JSON>`. Replays from `run_events` honoring `Last-Event-ID` header AND the `?last_event_id=N` query param, then tails live. With no id, replay the whole run. Closes after `RUN_FINISHED` or `RUN_ERROR`.
   - `POST /api/runs/{id}/stop`
   - `POST /api/runs/{id}/approvals/{call} {approve:bool}`
   - Event types and payloads: `RUN_STARTED{run_id}`, `TEXT_MESSAGE_CONTENT{delta}`, `THINKING{delta}`, `TOOL_CALL_START{call_id,name,args}`, `TOOL_CALL_END{call_id,ok,output_preview,duration_ms}`, `INTERRUPT{call_id,tool,args}`, `RUN_FINISHED{stop,usage,cost_usd}`, `RUN_ERROR{message}`. `usage` should carry `total_tokens` or `input_tokens`+`output_tokens`.
   - Auth floor on every route: cookie read before `Authorization`, `X-Facile-CSRF` on cookie mutations, strict per-user ownership of sessions and runs.
   - The SSE route needs a longer write timeout than the server's 15 s `WriteTimeout`: use `http.ResponseController.SetWriteDeadline` per write. Snapshot/rollback in `modules/computers` can also outlast 15 s.
3. **Wire in `apps/api/main.go` and `router.go`**: build the runner with `computers.NewProvider(computerService)` as its `ComputerProvider` (`Ensure(ctx, userID string) (*sandbox.Target, error)` and `Touch(userID string)`); `userID` is the decimal user id. Call `core/sandbox.PreflightCheck` on the target before the first run. Graceful shutdown must mark live runs `interrupted`.
4. **Embed the OpenUI prompt** for the web channel only: copy `apps/client/src/lib/openui/prompt.generated.txt` into `apps/api` and `//go:embed` it; append to the system prompt. Regenerate with `bun run gen:openui-prompt` in `apps/client`. See `apps/client/src/lib/openui/README.md`. Models should answer with a fenced block tagged `openui`.
5. **Full gate**: `sh scripts/check.sh` (Go, all workspace modules), `sh scripts/check-web.sh` (API + client), `filet check`. Build the image: `docker build -t bulle:test .` and run it against Postgres with `SSO_ONLY=true`, a fake issuer is not enough: use the real one only after Registre is set up (step 7).
6. **Release Bulle** (use the `/ship` skill). Release specifics:
   - `.github/workflows/release.yml` triggers on `v*` tags; `.goreleaser.yml` builds `apps/cli` (`dir: apps/cli`).
   - For `go install` to work, also tag `core/vX.Y.Z` and `apps/cli/vX.Y.Z`, and pin `apps/cli/go.mod` to the `core` tag (`go.work` has a `replace` for local builds). Until `core` is tagged, builds with `GOWORK=off` fail. Decide whether to keep `v*` tags for the binary release and add the subdirectory tags beside them.
   - Version files: check `CHANGELOG.md` (add an entry; there is an Unreleased section from Phase 0), `apps/cli` version var, `apps/client/package.json`. Last released CLI tag is `v0.91.1`; the restructure plus the web app is a minor bump (`v0.92.0`).
   - The Homebrew cask (`homebrew-tap/Casks/bulle.rb`) downloads from `download/v#{version}`. If GoReleaser does not generate the cask, change the tag part to match.
   - Never push a tag before the commit is on `origin`. Pushing `main` is safe for the web app: no Dokploy compose exists yet, so nothing auto-deploys. After the compose exists, a push to `main` redeploys (deploy agent noted CI does not gate it; use a PR then).
7. **Deploy** following `docs/deploy.md`. Only the user can do some steps; hand them exact commands:
   - **Done by the user**: DNS record for `bulle.facile.studio`.
   - **ruche, needs sudo**: `sudo apt install dosfstools` (`mkfs.vfat` missing), create group gid 2001 (`boite-sock`), install `boite-worker` from the v0.10.0 release archive (or build from `~/Code/Facile/boite`), install the systemd unit `scripts/boite-worker.service`, write `/etc/boite/worker.env` with `BOITE_WORKER_TOKEN`, `BOITE_WORKER_SOCK_MODE=0660`, `BOITE_WORKER_SOCK_GROUP=2001`, `BOITE_SSH_BIND=<docker bridge gateway, 10.128.0.1 on ruche>`, `BOITE_CONTROL_PUBKEY`. Then `systemctl enable --now boite-worker`. `/dev/kvm` exists on ruche. Verify with `curl --unix-socket`.
   - **casier**: the user must run `casier login` (browser). Then `casier secrets set -p bulle -e prod KEY VALUE` for every key in the table in `docs/deploy.md` (`DATABASE_URL`, `OIDC_*`, `ANTHROPIC_API_KEY`, `BOITE_WORKER_SOCK`/`TOKEN`, `BULLE_BOX_KEY` (private key, base64 on one line), `JOURNAL_*`). Only the user has the Anthropic key. Check with `casier check apps/api/.env.example -p bulle -e prod`.
   - **Registre client `bulle`**: add an entry to `~/Code/Facile/Registre/seed.yaml` with redirect `https://bulle.facile.studio/api/auth/oidc/callback` (and `http://localhost:5173/api/auth/oidc/callback`), add the `REGISTRE_SEED_BULLE_SECRET` passthrough in Registre's compose and the same value in Registre's Dokploy env (read-append-write; `casier push` would wipe its env). Push Registre and redeploy. The exact steps and the verification curl are in `docs/deploy.md`.
   - **Dokploy compose** at `https://gare.facile.studio` (project Facile, environment production): GitHub compose from `FacileStudio/Bulle`, path `./docker-compose.yml`, autoDeploy. Then `casier push dokploy <composeId> -p bulle -e prod` with `DOKPLOY_URL` and `DOKPLOY_API_KEY`.
   - Verify: `sh scripts/verify-deploy.sh bulle`, then log in at `https://bulle.facile.studio` and send a message.
   - Open risk: untested whether the container can reach the host SSH bind address. `docs/deploy.md` has a check and a fallback.

## Other repos with unpushed work

Each is a single small commit on branch `chore/bulle-rename`, unpushed, made by an agent that replaced `FacileStudio/bulle` with `FacileStudio/Bulle`:

- `~/Code/Facile/facile`: `internal/manifest/tools.yml` now has `repo: FacileStudio/Bulle` and `srcSubdir: apps/cli`. The `srcSubdir` change was the agent's call: confirm that `facile install bulle` really builds from `apps/cli`.
- `~/Code/Facile/homebrew-tap`: `Casks/bulle.rb` URLs.
- `~/Code/Facile/bulle.nvim`: one README link.

Push these after the Bulle release exists, so the URLs they point at work. GitHub redirects the old name meanwhile.

## Plan deviations (what differs from `docs/plan-web-ui.md`)

- `internal/agent`, `approval`, `toolview`, `theme`, `tasks` stayed in `apps/cli/internal` because they import the TUI. `core/` has no charm dependencies. The runner builds its own nacelle agent from `core/engine`, `core/provider`, `core/sandbox` instead of reusing `internal/agent`. Check how much it duplicates; extract a shared builder only if it is large.
- `scripts/check-web.sh` and the `web-*` mise tasks exist beside the CLI's `scripts/check.sh` and tasks, because the names collided. `mise.toml` at the root is now a monorepo root.
- `scripts/verify-deploy.sh` takes an app name (`bulle`), not a URL.
- `BULLE_BOX_KEY` (not `BOITE_CONTROL_KEY`) holds the control private key the API uses to SSH into boxes. `.env.example` still lists `BOITE_CONTROL_KEY`; remove or reconcile.
- Computers agent assumptions to confirm against the real worker: memory in MB and disk in GB, snapshot list shape `[{tag, created_at?}]`.
- The worker returns 409 `not_ready` for start/stop/ssh-key on a `creating` or `error` box. The API's `Ensure` waits for `running`; make sure it surfaces `error` to the client instead of polling forever.

## Later phases (not in release 1)

- **Phase 6, background jobs**: `jobs` table, in-process scheduler with a Postgres advisory lock, client `/jobs` UI, delivery to `web` and `matrix:<room>`.
- **Phase 7, same brain**: move memory, room history and jobs from `~/.bulle` files into Postgres behind store interfaces in `core/` (file implementation for the CLI, Postgres for the API); `bulle assistant migrate`; point the Matrix chat daemon (`apps/cli/cmd/chat*.go`) at the API or fold it into the API process; retire crontab-installed assistant jobs. Matrix daemon and web must never both write files and Postgres during cutover.
- **Phase 5 follow-ups**: nobody looked at the OpenUI Table, Chart and Form in a browser in light and dark; only server-side render tests exist. `@openuidev/lang-core` resolves to 0.4.1 (not 1.0.2) because svelte-lang 0.4.2 depends on `^0.4.1`. Its telemetry postinstall is blocked by Bun; do not enable it in CI.
- **Computer view** (terminal, file browser, take over) and an approvals inbox were explicitly deferred by the user.
- Egress filtering of the VM, per-user provider keys, quotas beyond the run caps.

## Gotchas

- **Never run broad `pkill -f` or `kill` patterns.** An agent killed the user's QEMU VMs `pingu` and `pinug` that way. `pingu` was restarted (running on port 2226). `pinug` was an orphan and is gone. Kill only PIDs you started.
- ruche is both production and the user's workstation. `/dev/kvm` exists; boxes share its RAM (62 GB) with every suite app.
- mise refuses untrusted configs inside git worktrees. Use `MISE_TRUSTED_CONFIG_PATHS=<worktree>` or run `bun --bun x <tool>` directly in `apps/client`.
- In `apps/client`, run `bun --bun x svelte-kit sync` before `svelte-check`, or it reports a missing `.svelte-kit/tsconfig.json`.
- `GOWORK=off` cannot resolve `core` (`go.sum` lacks it); use the workspace for local builds. `apps/cli` needs `-tags goolm` (the Matrix imports build a pure-Go Olm only under that tag; libolm is not installed).
- Leftover worktrees under `.claude/worktrees/` (branches `worktree-agent-*`) are all merged into `main` except `a45a30173daa49ed5` (the runner). After merging that one, remove them: `git worktree remove --force <path>` then `git branch -D <branch>`. A local Docker image `bulle-deploy-recon:local` can be deleted.
- Editor diagnostics about `$lib`, `Promise`, `Map` in the client appeared while agents worked in worktrees with no SvelteKit sync; they were noise, not real errors.
- Registre issuer is `https://sso.facile.studio` exactly (go-oidc compares byte for byte). `OIDC_REDIRECT_URL` must carry the `/api` prefix. Render `?error=` outside the SSO gate.
- Keep the one-container shape: no `PathPrefix`, no `stripprefix`, `expose` only, distroless `:nonroot`.

## Definition of done for release 1

- `https://bulle.facile.studio` serves the app; only Registre SSO signs in; `scripts/verify-deploy.sh bulle` is green.
- A new user's first message boots their VM (status `creating` shows in the chip), tools run in it, and the reply streams. No provider key appears in any VM file or tool output.
- Closing the tab mid-run loses nothing; reopening replays the run.
- A reply containing an `openui` block renders as muse components.
- A web redeploy leaves running boxes up (the worker is a host service, not in compose).
- `filet check`, `scripts/check.sh`, `scripts/check-web.sh` are green; CHANGELOG updated; tags pushed; release assets verified with `gh release view`.
