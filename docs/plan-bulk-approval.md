## Goal

Let the owner approve a batch of agent operations once, restyle the approval and mail-compose cards in muse style, and let `opus_project` read projects outside the default workspace.

## Why (evidence)

- `apps/api/internal/suite/tools_opus.go`: `opusProjectInput` has no `Workspace`, and the handler passes `set.defaults.workspace("")`, so a project in a second workspace cannot be fetched by id.
- `apps/api/internal/runner/gate.go`: every outward call creates one approval and blocks on it, so a 15-item bulk delete asks 15 times. `gate_policy.go` already has a run-scoped `grants` map, but only mail-style policies can set a scope.
- `apps/client/src/lib/components/ApprovalCard.svelte` and `MailApprovalCard.svelte`: bordered `rounded-fc-lg border bg-fc-component` boxes with a shield-icon header, stacked one per call.

## Research (batch approval practices)

- One decision for a whole batch needs one place that lists every gated call, so the reviewer sees the full set before deciding: https://www.ortussolutions.com/blog/boxlang-ai-34-blog-series-part-3-batched-approvals
- Approve, edit and reject stay per call in the data model; the batch is a UI and policy layer over them: https://docs.langchain.com/oss/concepts/human-in-the-loop
- Ask for every call and users stop reading. Separate classification (which tool, how risky) from enforcement (a deterministic policy, not the model): https://reddit.sentinel-team.org/posts/1uv8iul/snapshots/2026-07-17T19%3A45%3A55.60807Z (forum opinion, not a standard)

Decisions taken from this:
- The batch is a view over the same per-call approvals. A mismatched count cannot wedge a run.
- The grant is explicit opt-in, per tool, and lasts for the current run only.
- Mail sends never take a grant: each email stays reviewed and editable.

## Approach

Three tracks with disjoint files. Contract between tracks B and C: `Decision` gains `remember?: boolean`. When true on an approve, the server grants that tool for the rest of the run, so later calls of the same tool skip the prompt. The client groups concurrent pending approvals of the same tool into one batch card with "Approve all N" and "Deny all", which posts one decision per call.

## Steps (ordered)

Track A, opus workspace [module-path] (`apps/api/internal/suite/`)
1. `tools_opus.go` add `Workspace` to `opusProjectInput`, pass `set.defaults.workspace(in.Workspace)`. `opusProjectInput` is also used by `opus_delete_project`, so give `opus_project` its own input type with `Workspace`; delete keeps `ID` only.
2. `opus_write.go` `Project()` already takes the workspace ref. No change.
3. `write_test.go` or `tools_test.go` test: a project listed only in a non-default workspace resolves when `workspace` is passed.

Track B, bulk grant (`apps/api/internal/runner/`, `apps/api/modules/sessions/`)
4. `modules/sessions/handler.go` + `documentation.go` decideApproval body accepts `"remember": true`, passes it to `gate.decide`.
5. `gate.go` `decide` takes `remember bool`; on approve, when the tool has no mail policy, set `g.grants[name+"\x00*"] = true`.
6. `gate_policy.go` `granted` also returns true for `name+"\x00*"`. A call covered by the grant emits no interrupt and no approval row.
7. Mail tools (`mailPolicy`) ignore `remember`.
8. `gate_policy_internal_test.go` tests: remember grants later calls of the same tool, another tool still asks, mail still asks, grant does not outlive the gate.

Track C, client [muse] (`apps/client/src/lib/`)
9. `api/types.ts` add `remember?: boolean` to `Decision`.
10. `chat/reducer.ts` or a new `chat/batch.ts` helper: group an item's pending approvals by tool (non-mail only). Pure function with a test.
11. `components/ApprovalBatch.svelte` (new): one borderless block listing the N calls as compact rows (label plus key fields), buttons "Approve all N", "Deny all", and a quiet "Approve all and skip the rest of this run" using `remember`. Wire into `Message.svelte`.
12. `components/ApprovalCard.svelte`: remove border and box, use muse tokens, calm header (no shield icon block), danger tone only for destructive tools, mobile-first, targets at least 44px, no hover-only affordances.
13. `components/MailApprovalCard.svelte`: remove borders, plain underlined-field look for To/Cc/Bcc/Subject, larger message area, Send as the single primary action, Save as draft and Cancel as quiet text buttons, sticky action row on mobile.
14. `chat/chat.svelte.ts` add `decideAll(callIds, decision)` that sets all to `sending`, posts in parallel, and rolls back only the failures.

## Files to Modify / New

- `apps/api/internal/suite/tools_opus.go`, `write_test.go` or `tools_test.go`
- `apps/api/internal/runner/gate.go`, `gate_policy.go`, `gate_policy_internal_test.go`
- `apps/api/modules/sessions/handler.go`, `documentation.go`
- `apps/client/src/lib/api/types.ts`, `chat/chat.svelte.ts`, `chat/batch.ts` (new) + test, `components/ApprovalBatch.svelte` (new), `ApprovalCard.svelte`, `MailApprovalCard.svelte`, `Message.svelte`
- `CHANGELOG.md`, `docs/` touched only if the API doc page lists the decision body

## Exit criteria

- `scripts/check.sh` passes; `filet check` shows no new findings; client `bun run check` and tests pass.
- Fifteen concurrent `opus_delete_task` approvals show as one batch card; "Approve all 15" resolves all of them with 15 POSTs.
- After "skip the rest of this run", a later `opus_delete_task` in that run does not prompt; a new run prompts again.
- `courrier_send` always prompts and stays editable.
- `opus_project` with `workspace` set finds a project outside the default workspace.
- Neither card has a border, and both read at 360px width.

## Risks / unknown unknowns

- Tool calls inside one turn may run sequentially. Then no batch of pending cards ever exists at the same time, and only the run-scoped grant helps. Track C's grouping is still correct and costs nothing.
- A run-scoped grant on a destructive tool is a real blast radius. The button must say what it covers; the grant is per tool and ends with the run.
- The Android build embeds the client under `apps/client/android/.../assets`. Do not hand-edit those bundles.

## Skip (YAGNI)

- No cross-run or persisted "always allow" setting.
- No risk classifier or per-argument policy engine.
- No batch for mail sends.
- No change to the CLI approval flow (`apps/cli/internal/approval`).
