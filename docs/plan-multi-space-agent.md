# Plan: the Bulle agent acts as the user, across the user's spaces

Status: proposed 2026-11-06, supersedes the audience-rule draft of the same day.
Written against v3.20.3.

Checked against: `facile-spaces` convention, `store/spaces.go`, `suite/tools.go`,
`suite/defaults.go`, `runner/gate.go`, `runner/gate_policy.go`, `runner/gate_mail.go`,
`runner/browser.go`, `docs/plan-spaces-memory.md`.

## Goal

The agent in a Bulle chat can see and change everything the signed-in user can see and
change, in every space the user belongs to. Jobs and long-term memory stay in the
session's own space.

## Decision

The agent's authority mirrors the user's authority. Nothing narrower.

A Bulle space is where the agent looks first and how work is organised. It is not a
permission. This is not a preference, it is what the code already is:

- `store.Space` (`apps/api/internal/store/spaces.go:15-27`) holds five pointers and no
  rights: a Mycelium space, an Opus workspace, an Agenda calendar, a Courrier account,
  and the Registre id of a shared space.
- `suite.Defaults` (`suite/defaults.go:12-18`) is a fallback set. Every accessor
  (`account`, `calendar`, `workspace`) returns the caller's explicit choice when there is
  one and the space's only when there is not.
- The suite client resolves against the user id: `Opus.workspace` (`suite/opus.go:65`)
  and `Agenda.Calendars` (`suite/agenda.go:43`) return the user's own rows,
  `courrier_accounts` (`suite/tools_courrier.go:81`) lists every mailbox the user has.
- Bulle holds one set of the user's suite tokens. There is no narrower token to hold.

So a space boundary enforced inside Bulle would exist inside Bulle alone. Every other
app would still see one user with one set of data. That is the exact defect the
`facile-spaces` convention names: a space id used as a filter, not a boundary. Antenne,
Ardoise and Plume each shipped that mistake and each had to unwind it. Building one on
purpose is not worth doing.

Two consequences follow, and both are already true in the code:

- **The approval card is the control, not the space.** `gate.approve`
  (`runner/gate.go:104-135`) already asks the owner before every outward call, and the
  axis it uses is consequence: `ungrantable` (`gate_policy.go:42`) holds deletions,
  `files_save`, `browser_` and `web_fetch`; mail never takes a remembered grant. That is
  the right axis. Which workspace a task lands in is not in the same class as deleting a
  file.
- **The real sharing boundary is the guest channel.** A colleague talks to the owner's
  agent with no tools at all (`runner/run.go`, `guestAgent`). Someone else's
  conversation cannot read the owner's mail or act as the owner. That stays exactly as it
  is.

## What changes

Three things, and no new permission model.

1. **Discovery.** A read-only `spaces_list` tool, so the agent can name the workspace,
   calendar or account it wants instead of guessing an id from a prompt sentence.
2. **A truthful prompt.** `spacePrompt` (`runner/suite.go:47-52`) currently says the
   space's "chats, jobs and default accounts are separate from the user's other spaces".
   Two thirds of that is false today. Replace it with what is true.
3. **A narrower remembered grant.** Today "stop asking" grants a whole tool for the rest
   of the run, whatever its arguments. That is a pre-existing hole, and the blast radius
   is exactly the CRUD reach this plan makes useful. Fix it as part of this work, or the
   feature ships on top of it.

## Phase 1: discovery and an honest prompt

Ships alone. Changes no approval behaviour, widens nothing, because the reach is already
there and the agent simply does not know it.

### Steps

1. `apps/api/internal/suite/spaces.go` (new) — the `spaces_list` tool, a `readTool`,
   returning the user's spaces as
   `[{id, name, icon, shared, mycelium_space_id, opus_workspace, agenda_calendar, courrier_account}]`.
   `shared` is `RegistreSpaceID != ""`; `icon` is `Space.Icon`; `spaces_list` is the one
   tool that should also report which space the current session is in, so the agent can
   tell "here" from "there".
2. `apps/api/internal/suite/tools.go` — add a `Spaces func(ctx context.Context) ([]store.Space, error)`
   to `suite.Runtime` (`tools.go:62-65`), the same shape as the existing `PersonalSpace`.
   It stays nil-safe: when nil, the tool reports that spaces are unavailable.
3. `apps/api/internal/suite/tools.go` — register the tool in `Tools` (`tools.go:67-96`),
   outside the per-app switch, because spaces are Bulle's own concept and not a suite
   app. Add its name to the prompt note.
4. `apps/api/internal/runner/suite_runtime.go` — implement the new `Runtime.Spaces`
   closure over `r.cfg.Store.ListSpaces`, which is already user-scoped
   (`handler.go:28`). Guard on `r.cfg.Store == nil` like the existing closures.
5. `apps/api/internal/runner/suite.go` — rewrite `spacePrompt`:

   > You are working in the NAME space. When a tool needs a mailbox, calendar or
   > workspace and you do not name one, it uses this space's. You can read and change
   > anything the user's account reaches, in any of their spaces; call spaces_list to see
   > them. Jobs and memory belong to the space this conversation is in.

   Keep the sentence short. It explains, it does not authorise; the tools and the gate
   are the controls.
6. `apps/api/internal/suite/tools_test.go` — assert the tool is registered, is read-only,
   and returns the session's own spaces only.

### Why this is safe on its own

`spaces_list` returns only what `ListSpaces(ctx, userID)` returns, which is already the
user's own rows behind `RequireAuth`. It answers nothing the user cannot read in the
Bulle UI's own space switcher. The prompt change removes a false claim; it grants nothing
the tools did not already do.

## Phase 2: the remembered grant

Ships alone, and independently of Phase 1. Do this one first if you prefer: it is a bug
today.

### The hole

`gate.decide` (`runner/gate.go:168-169`) sets `grants[name+"\x00*"]` when the owner ticks
"stop asking", and `gate.granted` (`gate_policy.go:33-40`) treats that wildcard as
covering every later call of that tool in the run, whatever its arguments. The same
branch (`gate.go:170-174`, added in v3.20.2) also releases every already-waiting call of
that tool. "Approve all" in `ApprovalBatch.svelte:75` passes `remember: true`.

So: approve one `opus_create_task`, tick stop-asking, and a later injected instruction in
the same conversation can create tasks anywhere, with no card. With one space that is
bad. With every space reachable it is the whole feature turned against the user.

`mailPolicy.describe` returns an empty scope (`gate_mail.go:38`), and tools with no
policy at all fall back to `rawArgs` with an empty scope
(`gate_policy.go:19-29`). The only tool that produces a scope today is `browserConfirm`,
which returns the held request's origin (`browser.go:58`). So the scope mechanism exists,
is used once, and every suite write tool bypasses it and lands on the wildcard.

### The fix

Remember means "do not ask again for this tool, on this target, in this run". Unknown
target means keep asking.

1. `apps/api/internal/suite/scope.go` (new) —
   `ScopeFor(name string, input []byte, d Defaults) (scope string, summary map[string]any, err error)`.
   It takes the run's `Defaults` because a call that omits its workspace, account or
   calendar is targeting the space's, and the scope must name the real target. One
   function, one table, checked against each tool's input struct:

   | Tool | Scope |
   |---|---|
   | `opus_create_task` | `workspace:<ws>/project:<proj>` |
   | `opus_update_task`, `opus_delete_task` | `task:<id>` |
   | `opus_create_project` | `workspace:<ws>` |
   | `opus_update_project`, `opus_delete_project` | `project:<id>` |
   | `agenda_create_event` | `calendar:<id>` |
   | `agenda_update_event`, `agenda_delete_event` | `event:<id>` |
   | `sablier_start` | `project:<id>/task:<id>` |
   | `sablier_stop` | `timer` |
   | `courrier_update` | `account:<id>/action:<action>` |
   | `courrier_move` | `account:<id>/move:<folder>` |
   | `glouton_create_company` | `companies` |
   | `glouton_create_person` | `company:<companyId>` when the input names one, else `people` |
   | `glouton_update_company` | `company:<id>` |
   | `glouton_update_person` | `person:<id>` |
   | `courrier_send`, `courrier_save_draft` | `mailPolicy` owns them, unchanged |
   | `courrier_delete_draft`, `files_save` | unchanged, ungrantable |

   A create has no id yet, so it scopes to the collection it lands in, plus the parent
   when the input names one (a task's project, a person's company). Two notes on why
   these rows are shaped the way they are:

   - `courrier_update` carries its action as an argument, not in the tool name, and
     `takesGrant` (`gate_policy.go:48-59`) only reads the name. So a remembered
     `mark_read` would otherwise cover a later `action=delete`. Putting the action in the
     scope closes that, and it is why the row is not just `account:<id>`.
   - An unparseable input returns `err`; a `""` scope with no error means "no target".
     Either way the call is asked every time rather than falling back to a wildcard. Fail
     closed.

2. `apps/api/internal/runner/gate_suite.go` (new) — `suitePolicy`, implementing
   `approvalPolicy` over `suite.ScopeFor`. `describe` returns the readable record plus
   the scope. It resolves the workspace, calendar or account id to its name from the run's
   space list, so the card reads "in the Work space", not "workspaceId 7". `denied` is a
   no-op.
3. `apps/api/internal/runner/suite.go` — in `suiteTools` (`suite.go:31`), register
   `suitePolicy` for every name in `suite.WriteTools()` that is not already owned by
   `mailPolicy`.
4. `apps/api/internal/runner/gate_policy.go` — `granted` and `settle` lose the wildcard
   path. `settle` (`gate_policy.go:63-73`) already records `name+"\x00"+scope` when the
   scope is non-empty; keep that and delete the rest.
5. `apps/api/internal/runner/gate.go:168-174` — "stop asking" records `name+"\x00"+scope`
   instead of `name+"\x00*"`, and the release loop at `gate.go:170-174` matches on equal
   scope rather than on equal tool name. `takesGrant` stays as the coarse filter it is,
   next to the scope check.
6. `apps/api/internal/runner/gate_edit.go` — `editableFields` (`gate_edit.go:24`) covers
   mail only. Leave it. An edit that retargets a suite write is a second rewrite path
   through the gate, and v1 does not need it.
7. `apps/api/internal/runner/gate_policy_internal_test.go` — the tests below.
8. `apps/api/internal/runner/lazy_vm.go:36`, `apps/api/internal/runner/gate_mail.go:59` —
   unchanged; confirm no policy is registered under a name twice.

### Why the card matters here

With a narrower grant, the card is shown more often, so it has to be worth reading. A
suite write with no policy currently renders `rawArgs` (`gate_policy.go:22`), which is
the model's JSON. Naming the target resource and the space is what makes "stop asking"
a decision the owner can take safely. That is step 2, and it is the only client-visible
work in this phase; no client code changes, because `ApprovalCard` renders whatever
record the policy returns.

## Files to modify or add

Phase 1

- `apps/api/internal/suite/spaces.go` (new) — the `spaces_list` tool
- `apps/api/internal/suite/tools.go` — `Runtime.Spaces`, register the tool, prompt note
- `apps/api/internal/suite/tools_test.go` — registration and scoping test
- `apps/api/internal/runner/suite_runtime.go` — the `Spaces` closure over `ListSpaces`
- `apps/api/internal/runner/suite.go` — `spacePrompt`

Phase 2

- `apps/api/internal/suite/scope.go` (new) — `ScopeFor`, the target table
- `apps/api/internal/suite/scope_test.go` (new) — one case per table row, plus the failures
- `apps/api/internal/runner/gate_suite.go` (new) — `suitePolicy`
- `apps/api/internal/runner/gate_policy.go` — drop the wildcard from `granted` and `settle`
- `apps/api/internal/runner/gate.go` — scope-keyed remember and release
- `apps/api/internal/runner/suite.go` — register `suitePolicy` for the write tools
- `apps/api/internal/runner/gate_policy_internal_test.go` — the grant tests

Both

- `CHANGELOG.md` — under the next minor, one entry per phase

## Tests

Phase 1

- `spaces_list` returns exactly `ListSpaces(userID)`, and nil when the runtime has no
  store, rather than panicking.
- The tool is `ReadOnly` and therefore never reaches the gate.
- `spacePrompt` names the session's space and no longer claims the spaces are separate.

Phase 2, all in `gate_policy_internal_test.go` against the existing fixture

- A remembered `opus_update_task` on task 1 does not cover task 2: the second call asks.
- A remembered `opus_update_task` on task 1 does cover a second update of task 1: it
  does not ask twice.
- A remembered write does not release a waiting call with a different scope.
- A remembered `courrier_update` with `action=mark_read` does not cover a later
  `action=delete` on the same account.
- Deletes stay ungrantable by name: `opus_delete_task`, `agenda_delete_event`,
  `opus_delete_project`, `courrier_delete_draft`, and `files_save` never take a grant,
  whatever scope they compute (`takesGrant`, `gate_policy.go:48-59`).
- An input `ScopeFor` cannot parse makes the call ask every time, and never grants.
- `mailPolicy` and `files_save` still take no grant, unchanged.
- `browserConfirm`'s origin scope is unaffected.
- `suite.ScopeFor` returns the documented scope for every name in `suite.WriteTools()`,
  so a new write tool added later fails the test rather than silently becoming
  wildcard-grantable.

## Exit criteria

- Both phases pass `scripts/check.sh` and `filet check`. Run them as
  `env -u BULLE_ADDITIONAL_PROMPT XDG_CONFIG_HOME=<scratch> scripts/check.sh`: this shell
  exports `BULLE_ADDITIONAL_PROMPT`, and three `core/settings` tests fail on it
  otherwise. That failure is environmental and predates this work.
- `memory_search`, `memory_read` and `memory_write` in a session in space A never reach
  space B's Mycelium space (`suite/tools_memory.go:34-43` via `Defaults.memorySpace`),
  whatever the args. Pinned by test.
- `create_background_job` in space A always writes `space_id` A
  (`runner/job_tools.go:48`). Pinned by test.
- A `ChannelGuest` session is still built with zero tools (`runner/run.go`, `guestAgent`).
  Pinned by test.
- Manual pass: in a personal chat, "list my spaces", then "add a task in the Work
  space"; the card names the task's target and the space, and approving once does not
  silently cover a different task.

## Risks

- **Scope collisions.** Two different resources must not produce the same scope string,
  or a grant widens by accident. Namespace every scope by resource type, as the table
  does, and test that the ids are included.
- **A new write tool with no scope row.** `suite.ScopeFor` fails closed (asks every
  time) and the table-completeness test fails, so this is caught at build time rather
  than in production.
- **The card is now load-bearing.** More asks means more chances for the owner to click
  through. Keep the record specific: the resource, the change, and the space.
- **The agent may work in the wrong space by omission.** It is the same defaults
  fallback as today, but with discovery it can now do it on purpose. The prompt and the
  card both name the space; that is the mitigation, and it is a product choice, not a
  guard.
- **Concurrent sessions in this repo.** On 2026-11-03 two agent sessions committed
  overlapping work in Bulle. Stage explicit paths, never `git add apps/api`, and rebase
  on `origin/main` right before any tag.

## Skip, and why

- **A space permission model.** A boundary that exists only inside Bulle is a filter
  pretending to be a boundary, and the `facile-spaces` convention already lists three
  apps that shipped one. The approval card is the control; the space is context.
- **A `space` argument on the write tools.** `Opus.workspace` resolves a workspace by
  name, `courrier_accounts` and `agenda_calendars` return names with ids, and
  `spaces_list` supplies the mapping. A second way to name the same target is a second
  way to get it wrong.
- **Space CRUD by the agent.** Creating, renaming and deleting spaces is the user's own
  reach, so the mirror principle would include it. It is out of v1 on purpose: deleting a
  space deletes its sessions and jobs (`store/spaces.go:29-30`), which is the most
  destructive single call in the app, and it is a settings action a person should take in
  the UI. Revisit as its own change with its own card.
- **Cross-space memory or jobs.** Out of scope by decision, pinned by test.
- **Separate personas per space, space members, a memory editor in Bulle.** Unchanged
  from `docs/plan-spaces-memory.md`.
- **A Nuage pointer on a Bulle space.** `store.Space` has none, so `files_*` already sees
  the user's whole Nuage. Adding a folder pointer is a separate change nobody asked for.
- **New suite scopes or cross-repo work.** The user's delegated tokens already cover
  every pointer, which is the whole reason the mirror model is the cheap one.

## Sources

Consulted 2026-11-06.

- OWASP, AI Agent Security Cheat Sheet:
  https://cheatsheetseries.owasp.org/cheatsheets/AI_Agent_Security_Cheat_Sheet.html
  (authorize in the user's own context, least privilege per tool, fail closed, bind an
  approval to the exact action rather than to the tool)
- OWASP LLM06:2025 Excessive Agency: https://genai.owasp.org/llmrisk/llm062025-excessive-agency/
  (enforce authorization downstream in the user's context; minimize permissions)
- Meta, Agents Rule of Two: https://ai.meta.com/blog/practical-ai-agent-security/
  (a session may hold two of untrusted input, private data, state change, unless a person
  approves; the approval gate is the control here)
- AWS Prescriptive Guidance, Enforcing tenant isolation:
  https://docs.aws.amazon.com/prescriptive-guidance/latest/agentic-ai-multitenant/enforcing-tenant-isolation.html
  (tenant context is carried to the server and enforced server-side with tenant-scoped
  credentials)
- Simon Willison, the lethal trifecta: https://simonwillison.net/2025/Jun/16/the-lethal-trifecta/
- Not fetched, cited from search results only, unverified:
  https://www.varonis.com/blog/least-privilege-for-ai-agents,
  https://www.noma.security/blog/mcp-servers-agentic-risk-and-the-framework-that-protects-it
