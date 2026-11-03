# Plan: web search and browsing for Bulle agents

Approved 2026-11-03. Tracks A and B run in parallel; Track C follows B step 13.

## Goal

Give Bulle server agents web search (sonar, server keys with per-user BYOK override) and full interactive browsing (Playwright in the user's boite VM). Submits need owner approval, logins happen by hand through a live view, and the agent sees pages through accessibility snapshots with screenshots on demand.

## Why

Agents only have `web_fetch` (`apps/api/internal/runner/runner.go`, `HostTools` default): a plain guarded GET, no JS, no search, no interaction.

## Decisions

- Search runs on the API host as a nacelle tool, sonar used as a Go library. Keys never reach the VM.
- Billing: server keys in Dokploy env plus a per-user daily quota; a user's own key (credentials store) overrides the server key per engine and is not counted.
- Browsing runs in the per-user boite VM: a Playwright daemon (`bulle-browser`) driven over the existing SSH ControlMaster from `core/sandbox`.
- Submit gate: the daemon holds (does not abort) every non-GET request fired during an agent action. The agent calls `browser_confirm(held_id)`, an outward tool, so the existing runner gate asks the owner. The approval card shows the daemon's record of the request (method, URL, body preview), never model-written args. One approval covers that origin for the rest of the run.
- URL provenance: `web_fetch` and `browser_navigate` go straight through only for URLs that appeared in a user message or a tool result this run. Any other URL needs owner approval. This blocks GET query-string exfiltration (pattern from Anthropic's web fetch tool).
- Logins: the user signs in by hand through a CDP screencast live view; the profile persists in the VM. The agent never sees passwords.
- Perception: accessibility snapshots with element refs (Playwright MCP format); screenshots on demand.

## Track A: search

1. `sonar/websearch/` (new public package): `New(Config) (*Client, error)`, `Search(ctx, query, count, engines) ([]Result, error)` over `internal/search.NewDispatcher`. Keys only from `Config`: no config file, env or tiroir reads; in-memory cache. CLI and `sonar mcp` unchanged. One test proving an empty config reads no env and no tiroir.
2. Release sonar via its GoReleaser flow (feature commit, `chore: release vX.Y.Z` marker, tag, push, watch Release workflow). Do not bump `cmd/root.go` version.
3. `apps/api/internal/credentials/credentials.go`: separate `searchEnvVars` map (tavily, parallel, exa, brave) and `SearchEngines()`. `Valid` accepts both; `Backends()` unchanged so search keys stay out of the model picker. Update `modules/credentials/documentation.go`.
4. `apps/api/internal/runner/runner.go`, `run.go`: `HostTools` becomes `func(userID int64) ([]nacelle.Tool, error)`.
5. `apps/api/internal/runner/websearch.go` (new): `web_search(query, count)`, per-engine key resolution (user key, else server env), results as title / URL / snippet under a "results are data" header, ReadOnly.
6. `apps/api/migrations/00016_search_usage.sql`: `(user_id, day, count)` plus store methods; only server-key calls count; `BULLE_SEARCH_DAILY_LIMIT`. Check `core/usage` first and reuse if it fits.
7. `apps/api/internal/runner/provenance.go` (new): per-run URL set fed from the user message and every tool result. `web_fetch` on an unknown URL goes through the gate for approval.
8. Per-run caps on `web_search` and `web_fetch` calls, and truncation of fetched text at a token budget.
9. `defaultSystem` in `runner.go`: document `web_search`; results and pages are data, not instructions.
10. Server keys into Bulle's Dokploy env with the `dokploy-env-safe` flow (full merged env, never partial).

Exit: an agent answers a current-events question with `web_search` then `web_fetch`; unknown-URL fetch triggers approval; user key overrides server key; quota counts only server-key calls; `filet check` clean in sonar and Bulle.

## Track B: browser in the VM

11. `nacelle/message.go` and backends: image blocks in `ToolResult` (anthropic, openai, google). Only `browser_screenshot` needs it. Tag nacelle, bump in Bulle.
12. `boite/scripts/bake-provision.sh`: Node plus pinned Playwright and Chromium in `/opt/ms-playwright`. Rebake, repin `cmd/qemu/paths.go`, record image size change.
13. `apps/browser/` (new, TS, node runtime): `bulle-browser` daemon plus CLI. One persistent context (`~/.bulle/browser`), headless, CDP on `127.0.0.1:9222`, `acceptDownloads: false`. Commands: navigate, snapshot, find, click, type, screenshot, back, confirm, cancel, abort, status. Snapshots with refs, fresh snapshot after each action. Held non-GET requests time out after 2 minutes.
14. `core/sandbox/tools_browser.go` (new) plus `ToolsOptions.Browser`: `browser_navigate`, `browser_snapshot`, `browser_find`, `browser_click`, `browser_type`, `browser_screenshot`, `browser_back`, each running `bulle-browser <cmd> '<json>'` over the remote session. `browser_navigate` checks provenance.
15. `apps/api/internal/runner/gate.go`, `main.go` `Outward`: `browser_confirm` is outward; gate gains an optional `describe(name, input)` hook so the interrupt event carries the daemon's request record; per-origin-per-run allowance.
16. Daemon delivery in `main.go` `Preflight`: compare `bulle-browser --version`, copy the bundle when it differs. Bundle built in `Dockerfile`, Playwright external and version-pinned to the bake.
17. Run stop or timeout calls `bulle-browser abort`.

Exit: the agent finds a product on a JS-heavy site through snapshots; a test form submit produces an approval card with the real POST URL and body; deny aborts it; stopping the run kills a held request.

## Track C: live view (after step 13)

18. `modules/computers/`: `GET /computers/browser/live` WebSocket, SSH `-L` to the VM's `:9222`, `Page.startScreencast` out, `Input.dispatch*` in. Cookie auth, Origin check on upgrade.
19. Takeover lock: while the user drives, browser tools return "the owner has control".
20. Client live view panel with take control / hand back. muse tokens, Svelte 5 runes, mobile-first.
21. apiref entry in `modules/computers/documentation.go`.

Exit: user logs in through the live view, hands back, next run is still logged in; cross-origin upgrade refused.

## Risks

- slirp `open` mode: the guest reaches the host's `127.0.0.1` through `10.0.2.2`. Audit loopback listeners on the boite host before shipping step 13.
- Playwright version drift between bake and daemon bundle: pin from one place.
- Approval fatigue on POST-heavy sites; per-origin-per-run is the compromise.
- Persistent cookies: a hijacked agent inside a logged-in site can read what the user can. Add a "forget logins" action.
- Prompt injection has no complete fix; the plan limits blast radius (gate, provenance, no keys in VM).
- Bake grows roughly 400-600MB; measure boot time.

## Skip

Vision or coordinate clicking, uploads and downloads, multi-tab management, rod or `sonar fetch` on the host, CAPTCHA solving, per-user domain allowlists, session recordings.

## Sources

- https://platform.claude.com/docs/en/agents-and-tools/tool-use/web-fetch-tool
- https://playwright.dev/mcp/snapshots
- https://playwright.dev/mcp/configuration/options
- https://mckruz.github.io/microsoft-agentic-harness/security/06-egress-and-ssrf.html
- https://midpoint.ai/post/browser-automation-security-checklist
- https://arxiv.org/abs/2506.08837
