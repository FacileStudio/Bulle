# Plan: make Bulle the best agent harness for humans

Status: approved by the user on 2026-11-07 ("write a big plan, then run parallel agents on the tasks, then ship"). That instruction replaces the usual review checkpoint before execution.

## Goal

Make Bulle reliable on long work (compaction everywhere), easier for models to drive (tool names and arguments that match what frontier models were post-trained on), and calmer and more coherent for humans across the TUI, web, desktop and mobile. Deepen nacelle so Bulle stops working around it.

## Why (evidence)

Six read-only reviews ran on 2026-11-07:
- **Server never compacts.** `apps/api/internal/runner/runner.go:225` `prepare` loads the whole history. A chat that outgrows the window fails on every later run, and each failure is saved into the history, so it grows further.
- **The server tool loop has no iteration cap.** `main.go:141` never sets `MaxIterations`.
- **Opus status slugs are wrong.** The real slugs are `to-do`, `in-progress`, `in-review` and `done`, but `suite/opus_write.go:43` tells the model `todo` and `in_progress`. Writes fail with a bare error, and the `opus_tasks` filter silently returns `[]`.
- **Core tool names drift from the famous harnesses.** Examples: `edit_file(old,new)` vs `old_string/new_string`, `search_content` vs `grep`, `find_files` vs `glob`, and `run_command` timeout in seconds vs Claude Code's milliseconds.
- **nacelle's Approve hook sees only name and input.** Bulle recovers call ids by fuzzy matching (`runner/gate.go:65-105`, with a 10 ms sleep loop) and rewrites inputs through decorators (`gate_edit.go`, `cli/agent/shell_tool.go`).
- **A tool cannot end the turn.** ask_user costs an extra model call and relies on the model obeying the prompt.
- **Web notifications arrive twice when the app is open.** The account live stream can also die silently on mobile.
- **Gate races:** `claim` falls back to `queue[0]` (`gate.go:81`), edits are keyed by input rather than call id, and approval wait time counts against `RunTimeout`.

Sources: the review reports summarised in this plan, plus https://github.com/mattpocock/skills/blob/main/skills/engineering/improve-codebase-architecture/SKILL.md, https://code.claude.com/docs/en/tools-reference, https://www.anthropic.com/engineering/writing-tools-for-agents, https://opencode.ai/docs/tools/ and https://raw.githubusercontent.com/google-gemini/gemini-cli/main/docs/tools/file-system.md.

## Approach

Work in two waves of parallel workers. Each worker gets its own git worktree, and each owns a disjoint set of packages.
- **Wave 1:** everything that needs no new nacelle API, plus nacelle v0.32 itself, in its own repo.
- **Wave 2:** Bulle adopts nacelle v0.32 once it is tagged. That deletes the gate's call-id matching and the input-rewriting decorators, ends the turn natively for ask_user, and renames the core tools.

Reuse what exists:
- `core/compaction`'s engine becomes the single entry point for the TUI, headless runs and the server.
- jev stays an advisory judge; it never replaces the summarizer.
- `core/askuser` stays the ask_user contract.
- muse tokens apply to every client.

Checked against: `standards/migrations.md` (goose, sequential), `conventions/multi-agent-git-hygiene.md` (explicit paths, check the branch), `conventions/parallel-agents-one-go-package.md`, `conventions/suite-rename-sweeps.md` (old names survive only in migrations, changelogs and alias maps), `conventions/facile-versioning.md`, porte/CSRF, muse (no borders, tokens, runes, 44px targets), and filet (`filet.yml` limits; never raise them).

## Wave 1: parallel workstreams

Each workstream is one worker in one worktree, branch `wave1/<letter>-<slug>`. Exit for every stream: `sh scripts/check.sh` green, `filet check` clean, one commit per logical change, and conventional commit subjects. A stream does not tag or release.

### A. Compaction everywhere (core/compaction, server runner, CLI callers)

Owns: `core/compaction/**`, `apps/api/internal/runner/{run.go,runner.go,hydrate.go}`, `apps/api/main.go` (only the `MaxIterations` line), `apps/cli/internal/tui/compact*.go`, `apps/cli/internal/agent/{policy.go,support.go}`.

1. `core/compaction`: add `Engine.Maintain(ctx, conv, size int, overflow bool) (conv, Report)`. It owns tier choice, the stub pass with the 8 KB cache guard, the smart pass, the stub fallback, the thrash guard and the forced overflow pass.
2. `core/compaction/runner.go:28`: a judge error degrades to `Fold{Ledger: blocks}` and reports a warning. It never aborts the pass.
3. `core/compaction/force.go:66`: `LandsUnder` takes `overhead = size - EstTokens(Bytes(conv))`. `PartBytes` gives images and documents a fixed estimate of about 1.6k tokens.
4. `core/compaction/macro.go:38`: the summarizer gets ledger blocks rendered to text (`renderBlock`) in one user message, never raw tool_use or tool_result parts.
5. `core/compaction/tombstone.go:99`: a stubbed ToolResult keeps nothing heavy but counts its images as freed. `droppable` considers images.
6. Add `Policy.Over(size)` and use it at every boundary comparison.
7. `apps/api/internal/runner/runner.go` `prepare`/`complete`:
   - Before `Submit`, count tokens with the agent's backend, build a `compaction.Policy` from `Capabilities().ContextWindow`, call `Maintain`, and save with `saveHistory`.
   - On `engine.DetectOverflow`, force one pass and retry once.
   - A failed run no longer appends "The run failed" text that grows the history past the window. Store the failure as run status only.
8. `apps/api/main.go`: set `MaxIterations` (50) and default it in `runner.New` when it is 0.
9. `apps/api/internal/runner/hydrate.go`:
   - Write the date stamp into the stored user message once, in `prepare`. No request-time rewrite, so the prompt cache prefix holds.
   - Memoize attachment loads per run.
10. CLI: the TUI (`compact*.go`) and headless (`support.go`) call `Maintain`. Delete `evictionCanLandUnder`, `applyMaskFallback`, `resolvedPolicy` and `agent.ChatCompactor`, plus the dead `AfterCompact` counter. `/resume` resets size and thrash state.

Exit:
- A server test where a long fake history triggers a pass before the run, and an overflow error triggers one forced pass and one retry.
- `core/compaction` tests for the judge-error fallback and for `LandsUnder` with overhead.
- The TUI behaves as before (existing compaction tests pass).

### B. Suite tool coherence (server tools, client labels)

Owns: `apps/api/internal/suite/**`, `apps/api/internal/runner/{job_tools.go,chats_search.go,ideas_propose.go}`, `apps/client/src/lib/chat/tools.ts` (+ test), approval and phrase keys in `apps/client/src/lib/i18n/messages/chat.ts`.

1. Opus status (P0):
   - Resolve the status against `GET /api/column/{projectId}`: normalise case, `_` and spaces to `-`, then match the slug, then the name. Map `completed` and `closed` to the final column.
   - Apply this in Create, Update and the list filter.
   - On a miss, return an error that lists the valid statuses. Never return `[]` for a status that matches nothing.
   - Add the columns to the `opus_get_project` output. Fix the `"todo"` fixtures to `to-do` and add the resolve and error test.
2. Rename the suite read tools to `<app>_<verb>_<noun>`:

   | Old | New |
   |---|---|
   | `opus_tasks`, `opus_task`, `opus_projects`, `opus_project` | `opus_list_tasks`, `opus_get_task`, `opus_list_projects`, `opus_get_project` |
   | `agenda_events`, `agenda_event`, `agenda_calendars` | `agenda_list_events`, `agenda_get_event`, `agenda_list_calendars` |
   | `courrier_inbox`, `courrier_read`, `courrier_accounts` | `courrier_list_inbox`, `courrier_get_email`, `courrier_list_accounts` |
   | `glouton_contacts`, `glouton_company`, `glouton_person`, `glouton_outreach`, `glouton_suggestions` | `glouton_search_contacts`, `glouton_get_company`, `glouton_get_person`, `glouton_list_outreach`, `glouton_list_suggestions` |
   | `files_*` | `nuage_*` |
   | the background job tools | `jobs_create`, `jobs_list`, `jobs_update`, `jobs_delete` |
3. Arguments:
   - Use snake_case everywhere: `person_id`, `company_id`, `start_date`, `due_date`, `first_name`, `last_name`.
   - Replace every bare `id` with `task_id`, `event_id`, `project_id`, `email_id`, `job_id`.
   - Decode the old names as aliases. The schema advertises only the new names.
4. Write descriptions of three to four sentences: what the tool does, when to use it, when not to, and what each parameter means. Give the job tools the cron timezone and a `delivery` enum. Every error names the bad value, the valid values, and the tool that fetches them.
5. `suite.WriteTools`, `SelfWrites`, `InboxTools`, `MemoryTools`, the gate policy names (`runner/gate_policy.go` `ungrantable`) and `suite/scope.go` all move to the new names.
6. Alias old names at dispatch, so stored transcripts and approvals that still carry an old tool name replay. This is a map in `suite`, used by the runner's codec when it loads history.
7. Client `tools.ts`:
   - Update the labels and the `toolKind` regex: `nuage_` and `jobs_` count as apps, and `memory_*` and `chats_search` count as other.
   - Fix `approvalCanRemember` for `courrier_move`.
   - Add phrase keys for the tools that fall back to raw names, in all four languages.

Exit:
- `suite` tests cover the new names, the aliases, and opus status resolution.
- Client tests cover the new labels.
- The old tool names appear only in the alias map and the changelog: run a substring grep.

### C. Client reliability and polish (web, desktop, Capacitor)

Owns: `apps/client/src/**` except `lib/chat/tools.ts` and the approval keys (stream B), plus `apps/desktop/**`.

1. `src/service-worker.ts`: skip the system notification when a focused client window is on that route. Android: no foreground banner while the app is visible.
2. `src/lib/live/stream.ts`:
   - Add a watchdog that reconnects when no data arrives for 2.5× the ping interval.
   - Restart on `visibilitychange` and Capacitor `appStateChange`.
   - A 401 signs out, like `apiFetch` does.
3. `src/lib/api/http.ts:34` and `chat.svelte.ts:126`: fatal stream loss shows a "Reconnecting" state and resyncs on reconnect. It never marks the run failed. The string goes through i18n.
4. `components/Reply.svelte` and `s/[id]/+page.svelte`:
   - Render once per animation frame and re-render only the last markdown block while streaming.
   - Do the auto-scroll in the same frame.
   - Move `aria-live` to a status region that announces only "finished" or "needs approval".
5. `components/AskCard.svelte`:
   - Add `ask.*` i18n keys in four languages.
   - `disabled={closed || disabled}` on the fieldsets.
   - Add an `isComposing` guard.
   - Add digit shortcuts 1-4, like the TUI.
6. Read `grantable` from the INTERRUPT payload when it is present (stream D adds it). Fall back to `approvalCanRemember`.
7. Android updates: on native, compare the app version with `/version` and the GitHub latest release, and link the APK. Never claim "up to date" without a real check (`updater.svelte.ts:34`).
8. Small fixes:
   - Hover-only delete becomes `[@media(hover:hover)]` (`SessionList.svelte:108`).
   - Desktop notification click navigates to its route (`desktop-notifications.ts:21`).
   - The desktop CSP comes from the API base (`apps/desktop/tauri.conf.json`).
   - Read the launch deep link (`getCurrent()` for Tauri, `App.getLaunchUrl()` for Capacitor).
   - The login verifier moves to `Preferences` on native (`native.ts:31`).
   - Add one `assistantName()` helper for the five copies of the fallback.
   - The artifacts list clears when the space changes.
   - Prettier-format `tools.ts` only after stream B merges. That one is left to B.

Exit: vitest and svelte-check are green, and a test covers the watchdog reconnect and the service-worker focus skip.

### D. Gate correctness, API security, system prompt

Owns: `apps/api/internal/runner/{gate.go,gate_policy.go,gate_edit.go,gate_mail.go,gate_suite.go,blocked.go,browser.go,personality.go,suite.go,live.go,live_test.go}`, `apps/api/internal/matrixbot/**`, `apps/api/internal/push/**`, plus any security finding outside the files streams A and B own.

1. `gate.go:81`: `claim` matches exactly. It takes `queue[0]` only when that is the only queued call.
2. `gate.go:95`: `note()` signals a per-name channel, and `callID` waits on it, bounded by ctx. Remove the sleep loop. This stays until wave 2 deletes the whole mechanism; keep the change small.
3. `gate_edit.go`: key edits by call id, not by name plus input.
4. Time spent waiting on approval extends the run deadline (the gate already measures blocked time).
5. A `CreateApproval` error logs and emits an error event instead of a silent denial.
6. Add `grantable` to the INTERRUPT payload (`takesGrant`).
7. `personality.go:29`: drop the duplicate identity line, and say "personal assistant" rather than "coding assistant". When `cfg.Suite` is nil, `NoMemoryNote` drops the "connect Mycelium in Settings" advice.
8. Security findings (review 2026-11-07):
   - **P0 job tools behind approval.** Add `create_background_job`, `update_background_job` and `delete_background_job` (jobs_* after stream B) to `Outward` in `apps/api/main.go:148`, with a policy card showing the schedule and the prompt. Enforce a 15-minute minimum interval and 20 jobs per user in `runner/job_tools.go`. Coordinate the names with B: D adds the gate policy keyed by the new `jobs_*` names, and B's alias map covers the old ones.
   - **P0 `run_command` after untrusted content.** Once a run has read untrusted content (mail, web, Nuage files, attachments), `run_command` asks for approval. Mark the taint in the gate when a tool from the untrusted-source set returns. VM default-deny egress is infra work and goes in the Risks section.
   - **P1 `memory_write` to `profile.md`** asks for approval once the run is tainted.
   - **P1 narrow `browser_navigate` check.** Ask only when the URL carries a query string or path token that the user's own text did not contain verbatim. Plain page reads stay unasked, as the user asked. Base it on `git show 199d1e9^:apps/api/internal/runner/provenance.go`.
   - **P1 stream deltas.** `events.go:48` inserts one row per token. Batch every 250 ms or 4 KB, and compact the events of finished runs (new migration `[migrations]`).
   - **P1 shutdown.** `main.go:253` drains the runner and the HTTP server at the same time, with separate budgets. `RegisterOnShutdown` cancels the SSE handlers, and a deadline counts as a clean exit.
   - **P2:**
     - The Matrix guest gets a per-sender rate limit and its own run slot.
     - `mail_seen` becomes one `INSERT … SELECT unnest(?::bigint[])`. Ideas and memory runs pass `SeeMail: nil`.
     - Delete orphaned attachments after 30 days.
     - Wire `Redactor: runner.EnvSecrets()` in `main.go`.
     - Cap live SSE subscribers at 10 per user.
   - **Flaky test root cause.** In `gate.decide` (`gate.go:194-199`), call `g.decided(...)` before `decision <- approve`. The reviewer confirmed this passes 4,500 iterations.
   - **Mycelium stored XSS.** `MarkdownMuse` has an `{@html}` path and mermaid runs at `securityLevel: loose`. That is the Mycelium repo, not Bulle; file it there.

Exit: gate tests for two identical concurrent calls (the right card, the right edit), a deadline that is not consumed by an approval wait, and the flaky test passing 200 runs with `-race -count=200`.

### E. TUI polish

Owns: `apps/cli/internal/{tui,toolview,theme,diff,menu,tasks,layout,approval,ask}/**` except `tui/compact*.go` (stream A).

Steps come from the TUI review (see "TUI findings" below). Exit: the existing TUI tests pass, plus a test per fix. Capture before and after in tmux at 100 and 60 columns, dark and light.

### N1. nacelle v0.32 (repo /home/yann/Code/Facile/nacelle, breaking)

Owns: the whole nacelle repo.

1. **P1-3:**
   - `nacelle.EndTurn(result string) error` and `StopYield Stop = "yield"` (not `Complete()`).
   - Both loops record the result as a successful ToolResult, let the other calls in the round finish, then end with `KindDone{Stop: StopYield}`.
2. **P1-1:**
   - `type Call struct{ID, Name string; Index int; Input json.RawMessage}`.
   - `type Decision struct{Deny, Name string; Input json.RawMessage; Note string}`.
   - `type Approve func(ctx, Call) Decision`.
   - `HookEvent` carries `Call`, and `HookResult` can return `Input`.
   - The loop reports the final Name and Input in `ToolEvent`.
   - Ship a `nacelle.ApproveBool(func(ctx, name, input) bool) Approve` adapter, so a consumer can migrate in one line.
3. Built-in tool renames, matching Claude Code, Gemini and OpenCode:
   - `edit_file(path, old_string, new_string, replace_all)`.
   - `search_content` becomes `grep(pattern, path, glob)`.
   - `find_files` becomes `glob(pattern, path)`.
   - `run_command` becomes `bash(command, timeout_ms, description)`.
   - `tasks` becomes `todo_write(todos[{content,status}])`.
   - `parallel_agents` items become `{description, prompt}`.
   - Decode aliases for `file_path`, the old argument names, and `multiSelect`.
   - Export `nacelle.LegacyToolNames map[string]string` (old to new) so consumers can replay old transcripts.
4. **P3-2/P3-3:** delete `CompactConversation`, the compaction hook points, and the untyped context keys for Anthropic compaction.
5. Update the CHANGELOG with a migration section. Tag `v0.32.0` once every test passes.

Exit: `go test ./...` green in nacelle, the migration notes are written, and the tag is pushed. Nothing in Bulle changes in wave 1.

### S. Sablier decision endpoint (owned by the Sablier session)

`POST /api/ai/decide` and the `RecordAgentSession` metadata. That work happens in another session and merges to main on its own. Wave 1 does not touch `runner/timetracking.go` or `suite/sablier.go`.

## Wave 2: adopt nacelle v0.32 (sequential after wave 1 merges)

### F. Bulle on nacelle v0.32

Owns: everything nacelle touches in Bulle (`go.mod` files, runner gate, `core/sandbox`, `core/settings`, `apps/cli/internal/{agent,diff,ide}`).

1. Bump nacelle to v0.32.0 in `apps/api`, `apps/cli` and `core`.
2. Gate: `approve(ctx, Call) Decision`.
   - Delete `note`, `claim`, `callID`, the orphan counter, the `asked` map, the `gate.note` call in `run.go`, `revisable` and `takeRevision`.
   - An edit or a draft redirect becomes `Decision{Name, Input, Note}`.
3. Server ask_user returns `nacelle.EndTurn`. Delete the "end your turn" prompt text, and map `StopYield` to a waiting session status.
4. CLI:
   - Delete `unwrapCallTool` and `commitCoAuthorTool` (rewrite through `Decision.Input`).
   - IDE hooks key on `Call.ID`.
   - The TUI ask_user stays blocking.
5. `core/sandbox` tools mirror nacelle's new names and reuse its input types. `core/settings` keeps the old YAML keys as aliases. `cli/internal/diff/diff.go` reads `old_string`/`new_string`, and the toolview icons follow the new names.
6. Replay: the runner codec and the CLI session loader map names through `nacelle.LegacyToolNames`. Test resume on Anthropic, OpenAI and Gemini.

Exit: the full gate is green, and a resumed pre-v0.32 session replays. On a live run, ask_user yields with no extra model request, and an edited `courrier_send` runs with the edited input.

## Wave 3: follow-ups (planned, not in this release)

- nacelle P2-2: one loop with a per-turn `Provider` seam. Fixes the Anthropic loop ignoring `ReadOnly` and `PlanCalls`.
- nacelle P1-2: the run returns its transcript. Deletes Bulle's two transcript rebuilds and keeps images and reasoning.
- nacelle P2-1: `Tool.Spec()` and `Wrap`, so decorators stop dropping the optional interfaces.
- nacelle P2-3: `Catalog` and `Router` for deferred MCP tools, with MCP images and `readOnlyHint`.
- nacelle P3-1: a message JSON codec. Bulle's codec drops `Image` and `Document`.
- Memory UI: list and delete what the assistant remembers. Needs a small API and is a user-control gap.
- Deferred tool loading per app: about 70 suite tools are mounted, and OpenAI suggests fewer than 20.

## Releases

1. After N1: nacelle `v0.32.0` (tag in nacelle, CHANGELOG with migration notes).
2. After wave 1 merges (A, B, C, D, E, plus S if it landed): Bulle `v3.28.0`.
3. After wave 2: Bulle `v3.29.0`.

The changelog, both version copies (`apps/cli/main.go`, `apps/client/package.json`), the gate, then tag, watch the Release and Desktop runs, and verify the assets.

## Risks

- **Renames:** stored sessions name old tools. Providers may reject a history that names a tool not declared in the request, hence the alias maps at load time. Test resume per provider before v3.29.0.
- **Wave 1 overlap:** A and D both touch `apps/api/internal/runner`, in disjoint files. B and C both touch `i18n/messages/chat.ts`; B adds approval and phrase keys and C adds `ask.*` keys, so the merge is textual.
- **Cache:** moving the date stamp changes stored messages. Older sessions keep working, because the stamp is plain text.
- **nacelle v0.32 is breaking** for `nacelle-tui` and other consumers. The `ApproveBool` adapter plus the migration notes keep their move to one line.
- **Other sessions are active:** the Sablier session, plus the session that shipped v3.26.0. Merge by rebasing onto origin/main. Never force-push, and commit explicit paths.

## Skip (YAGNI)

- No new provider, no new UI framework, no tool count reduction beyond the renames in this release (deferred loading is wave 3).
- No change to jev's model or endpoint: jev remains an advisory judge.
- No rename of browser_* tools (already close to Playwright MCP) or memory_* tools (different model than Anthropic's memory tool).
