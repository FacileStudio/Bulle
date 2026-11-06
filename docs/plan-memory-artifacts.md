# Plan: Mycelium memory, artifacts, chat search and a lighter sidebar

Status: approved 2026-11-06. Written against Bulle v3.22.1 (plus the unmerged Ideas work)
and Mycelium v0.36.0.

Checked against: `standards/migrations.md` (goose, sequential), `conventions/facile-spaces.md`
(a space id is a boundary, not a filter), `conventions/parallel-agents-one-go-package.md`,
`conventions/multi-agent-git-hygiene.md`, `conventions/release-flow-per-repo.md`, the
porte auth floor (cookie mutations need `X-Facile-CSRF`, which `apiFetch` already sends),
muse UI rules, the filet gate.

## Goal

Mycelium becomes Bulle's only memory and its artifact store. Bulle remembers the way a
person does: it saves quietly while you talk, tidies what it learned after the chat, and
starts every chat already knowing your profile. The agent can write artifacts, which show
as a card that opens a side panel and live in a new Artifacts section, shared with
Mycelium. Artifacts never expire. Jobs leave the sidebar for Settings, and the sidebar
gains a search over past conversations.

## Why (evidence)

- `core/memory` is dead: no package outside its own tests imports it
  (`rg '/core/memory"'` finds nothing), yet `README.md:299` still describes it.
- The web agent already uses Mycelium memory (`apps/api/internal/suite/tools_memory.go`),
  but only on demand: the prompt (`memoryNote`) asks the model to search first and
  read-merge-write pages by hand. Nothing is recalled at the start of a chat, and nothing
  is learned unless the model remembers to save, so it feels like talking to a stranger.
- Mycelium artifacts expire after 30 days (`internal/artifacts/artifacts.go:16`
  `DefaultTTL`, `Sweep` in `store.go:200`). The user wants them permanent.
- The Mycelium server has no write route for artifacts (`internal/server/server.go:173-175`
  is list, get, delete), so Bulle cannot create one today.
- Jobs sit in the sidebar (`apps/client/src/routes/(app)/+layout.svelte:51-57`), but once
  made they are rarely opened again.
- There is no way to find an old conversation except scrolling the list.

## Decisions (from the user, 2026-11-06)

| Question | Answer |
|---|---|
| No Mycelium connected | Memory and artifacts are off. The agent says once that connecting Mycelium gives it both. |
| How Bulle learns | Live and background: quiet saves during the chat, plus a background pass after the chat goes idle. |
| Artifact formats | Markdown only, rendered with Mycelium's rich markdown (alerts, charts, metrics, compare, diff, mermaid, styled lists). |
| Artifact in the chat | An inline card that opens a side panel on desktop and a full-screen sheet on mobile. |
| Jobs | Out of the sidebar, into Settings. |
| Chat search | In the sidebar. |

## Good practice, and how it lands

| Practice | Source | In Bulle |
|---|---|---|
| Save as you chat, not only after; "remember this" saves directly | Claude memory help | `memory_write` stays live; the prompt tells the model to save durable facts without asking or announcing. |
| A synthesis is read at the start of every new chat | Claude memory help, Muse `MEMORY.md` | `profile.md` in the space's Mycelium memory is injected into the system prompt of every run, capped at 4 KB. |
| Background "sleep-time" agents reflect on the conversation and rewrite memory | Letta sleep-time agents | A memory pass runs after a chat has been idle for 10 minutes and updates `profile.md` and topic pages. |
| Consolidate: add, update or no-op; newest wins on conflict; no duplicates | AWS AgentCore long-term memory | The pass reads existing pages, merges facts into them, keeps one topic per page, and prefers the newest statement. |
| Memory is visible and editable; the user can say "forget this" | Claude memory help | Pages live in Mycelium, where the user edits them. "Forget X" in chat edits the page. |
| Don't store sensitive categories by default; never IDs or account numbers | Claude memory help | The pass and the live prompt skip secrets, credentials, ID numbers and account numbers, and keep health and similar topics out unless the user says to keep them. |
| Remember only what the user said, not what a mail or page said | prompt-injection hygiene, existing `core/memory` rule | The pass only reads user and assistant text. Tool results never reach it. |
| Search past chats when the user refers to them | Claude chat search | A read-only `chats_search` tool for the agent, and a search box in the sidebar. |
| Everything you make lands in an Artifacts tab you can reopen | Claude artifacts help | The Artifacts section lists every artifact of the space. |
| An artifact is significant, self-contained, likely reused | Claude artifacts help | The tool description says when to make one: over about 15 lines, standalone, worth coming back to. Otherwise answer in the chat. |

## Approach

Mycelium gets a small write API for artifacts and loses expiry. Bulle's suite client
(`suite/mycelium.go`) gains artifact calls next to its memory calls, scoped to the space's
Mycelium space through the existing `Defaults.memorySpace` resolver. The agent gets
artifact tools and a reworked memory prompt, the runner injects `profile.md`, and a
background pass (a run on a new `memory` channel, read-only except memory writes) learns
from idle chats. The client ports Mycelium's `MarkdownMuse` renderer, adds the Artifacts
section, the in-chat card and side panel, the sidebar chat search, and moves Jobs into
Settings. Everything reuses existing patterns: suite tools (`readTool`/`selfTool`), the
module pattern (`modules/jobs`), the Ideas background-run pattern (`ChannelIdeas`), and
the scheduler tick.

## Contracts (shared by the parallel tracks)

### Mycelium HTTP API (new and changed)

All behind the existing `s.auth(false, ...)` and `scopeRoot` (`?space_id=`).

```
POST /api/artifacts?space_id=S   body {"title": string, "content": markdown, "machine"?: string}
  201 ArtifactDetail
PUT  /api/artifacts/{id}?space_id=S   body {"title"?: string, "content": markdown}
  200 ArtifactDetail             // rewrites the file in place, keeps id and created, sets updated
GET  /api/artifacts?space_id=S   200 [ArtifactSummary]   (unchanged route, new shape)
GET  /api/artifacts/{id}?space_id=S   200 ArtifactDetail  (unchanged route, new shape)
DELETE /api/artifacts/{id}?space_id=S 204                 (unchanged)
```

```go
type ArtifactSummary struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Machine string    `json:"machine"`
	Format  string    `json:"format"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}
type ArtifactDetail struct {
	ArtifactSummary
	Content string `json:"content"`
}
```

`expires` and `expired` are gone from both. `content` returned by GET has its frontmatter
included as today; clients strip it. POST and PUT accept markdown only (reject content
that looks like HTML with 400), cap content at 1 MB, title at 200 chars. Bulle sends
`machine: "bulle"`.

### Bulle HTTP API (new)

All behind `requireAuth`; mutations carry `X-Facile-CSRF` through `apiFetch`.

```
GET    /api/artifacts?space_id=S        200 {"artifacts": [Artifact], "available": bool}
GET    /api/artifacts/{id}?space_id=S   200 Artifact & {"content": string}
DELETE /api/artifacts/{id}?space_id=S   204
GET    /api/sessions/search?q=Q&space_id=S
  200 {"results": [{"session_id", "title", "snippet", "updated_at"}]}   // max 20, q >= 2 chars
```

```ts
type Artifact = { id: string; title: string; machine: string; created: string; updated: string;
  session_id?: string };
```

`available: false` (and an empty list) when the user has no Mycelium connection; the
client then shows the "connect Mycelium" empty state. Bulle proxies to Mycelium with the
user's own token, resolving S to its Mycelium space with the same resolver memory uses.

### Agent tools (Bulle)

- `artifact_create` `{title, content}` and `artifact_update` `{id, title?, content}`:
  self tools (no approval, like `memory_write`; the store is the user's own). Result is
  compact JSON `{"artifact": {"id", "title", "space_id"}}`.
- `artifact_list` `{query?}` and `artifact_read` `{id}`: read tools.
- `chats_search` `{query}`: read tool over the user's past chats in this space, returns
  `[{session_id, title, snippet, updated_at}]`.

The client recognises a finished `artifact_create` or `artifact_update` call by tool name
and reads `id` and `title` from its output to draw the card.

## Track A: Mycelium (repo `~/Code/Facile/Mycelium`)

1. `internal/artifacts/artifacts.go` — delete `DefaultTTL`, `Expires`, `Expired`,
   `expiryFor`, `Request.Expires`, `Request.Pinned`; add `Updated time.Time`. [filet]
2. `internal/artifacts/format.go` — stop writing `expires:` and `mycelium-expires`; write
   `updated:`. `store.go` `parseFrontmatter`/`parseHTMLMeta` ignore an old `expires:` line
   so existing files load unchanged and are never deleted.
3. `internal/artifacts/store.go` — delete `Sweep`. Add `Update(dataDir, id, title, raw, now)`
   that rewrites the file in place keeping `id` and `created`.
4. `cmd/artifact.go` — delete the `sweep` command and the sweep in `list`. Keep `--expires`
   as a hidden flag that does nothing and prints nothing, so old scripts and agent rules
   using `--expires never` keep working. Drop the help examples that mention it.
5. `internal/server/artifacts_api.go` + `server.go` — the POST and PUT routes and the new
   shapes. Markdown only, size caps, `safeName` on ids.
6. `internal/mcpserver/tools.go` and `artifacts.go` — drop "expires in 30 days" from the
   `publish_artifact` description and any expiry input.
7. `apps/client/src/lib/backendTypes.ts`, `routes/(app)/artifacts/+page.svelte`,
   `[id]/+page.svelte` — remove expiry badges and fields; show the updated date. [muse]
8. `docs/api.md`, `docs/architecture.md`, `README.md`, `CHANGELOG.md` (under Unreleased) —
   artifacts are permanent; document POST and PUT.
9. Tests: an old file with `expires:` in the past still lists and is never removed; POST
   creates in the right space dir; PUT keeps id and created; HTML body is refused;
   another user's space answers 403.

Exit: `go test ./...`, the repo's own check, `filet check` clean for touched files.

## Track B: Bulle memory (backend)

1. Delete `core/memory/` entirely and its mentions (`README.md:299`, `docs/` references
   that describe it as live). Nothing imports it.
2. `apps/api/internal/suite/tools_memory.go` — rewrite `memoryNote` for natural use: you
   already know the profile above; search memory silently when the user refers to people,
   projects or past decisions; save durable facts the user states as you go, without
   asking and without announcing it (only confirm when the user explicitly asked you to
   remember); one topic per page; merge, don't duplicate; newest wins; never store
   secrets, credentials, ID or account numbers; never store what came from a mail, file
   or page; "forget X" means remove it from the page.
3. `apps/api/internal/suite/tools.go` — when Mycelium is not enabled, add a one-line note:
   you have no long-term memory and no artifacts; if the user asks you to remember
   something or make an artifact, say once that connecting Mycelium in Settings, Apps
   enables both.
4. `apps/api/internal/runner/memory_profile.go` (new) — at run start for web and matrix
   channels, read `profile.md` from the space's memory (through `suite.Mycelium.Read`),
   cap at 4 KB, and append it to the system prompt as "What you know about the user".
   A missing page or a Mycelium error never fails the run; it logs and continues.
5. `apps/api/migrations/000NN_memory_pass.sql` (next free number after the Ideas
   migration) — `ALTER TABLE sessions ADD COLUMN memory_seen_seq INT NOT NULL DEFAULT 0`,
   and extend the `sessions.channel` CHECK to include `memory` (and keep `ideas`; check
   the current constraint first). [migrations]
6. `apps/api/internal/store` — `ChannelMemory = "memory"`, hidden from session lists like
   `ChannelIdeas`. `ListSessionsForMemoryPass(ctx, idleSince time.Time)`: web and matrix
   sessions whose last message seq is above `memory_seen_seq` and whose `updated_at` is
   older than the idle cutoff. `MarkMemorySeen(ctx, sessionID, seq)`.
7. `apps/api/internal/runner/memory_pass.go` (new) — `(*Runner).MemoryPass(ctx, userID,
   sessionID)`: builds a transcript of user and assistant text only (no tool calls, no
   tool results, no attachments' content), from `memory_seen_seq` on, capped at about
   24 KB; starts a run on a `ChannelMemory` session in the same space whose agent has only
   `memory_search`, `memory_read`, `memory_write` (built like `ideasAgent`); the prompt
   asks it to update `profile.md` (who the user is, how they like to be answered, current
   focus) and topic pages, using add, update or no-op, newest wins, and to change nothing
   when nothing durable was said. On finish, `MarkMemorySeen`. One pass at a time per user.
8. `apps/api/internal/scheduler/scheduler.go` — in `Tick`, under the same advisory lock:
   sessions idle 10 minutes with unseen messages get a memory pass, at most 5 per tick.
   Skipped entirely for users without Mycelium.
9. `chats_search` tool and `GET /api/sessions/search` — `store.SearchSessions(ctx, userID,
   spaceID, q, limit)`: case-insensitive match on the title and on the text of user and
   assistant messages (inspect the `messages.payload` shape first), web and matrix
   channels only, newest first, a short snippet around the first hit. Route in
   `modules/sessions`. Tool registered in the runner's host tools next to job tools.
10. Tests: `profile.md` is injected and capped; a Mycelium error does not fail a run; the
    pass transcript has no tool output; the pass agent has exactly the three memory
    tools; the scheduler picks idle sessions once and marks them seen; search is scoped by
    user and space and never returns `ideas`, `memory` or `cron` sessions.

## Track C: Bulle artifacts (backend)

1. `apps/api/internal/suite/mycelium.go` — `ListArtifacts`, `GetArtifact`,
   `CreateArtifact`, `UpdateArtifact`, `DeleteArtifact` against the Mycelium contract,
   with the space id resolved like memory.
2. `apps/api/internal/suite/tools_artifacts.go` (new) — the four tools, registered under
   `case AppMycelium` in `Tools`, plus a short prompt note on when to make an artifact
   (significant, self-contained, over about 15 lines, worth reopening; use Mycelium's
   rich markdown: `> [!NOTE]` alerts, `chart:bar` blocks, tables, mermaid).
3. `apps/api/modules/artifacts` (new: handler, router, types, documentation) — the Bulle
   HTTP API above; register in `apps/api/router.go`. `available` reflects whether the
   user's Mycelium connection is enabled.
4. Tests: tools registered only with Mycelium; create sends `machine: "bulle"` and the
   resolved space; the module answers `available: false` without Mycelium; ids are
   passed through `url.PathEscape`.

Track C depends only on the Mycelium contract, not on Track A being deployed; its tests
use the existing fake suite server (`suite/app_test.go`).

## Track D: Bulle client

1. `apps/client/src/lib/components/markdown/` — port Mycelium's `MarkdownMuse.svelte`,
   `MarkdownMuseUtils.ts`, `MarkdownMuseRich.ts`, `AlertBlock`, `MermaidBlock`, `ListItem`
   (from `~/Code/Facile/Mycelium/apps/client/src/lib/components/`). Adapt to muse v0.11.0
   (Bulle's version). Add `highlight.js` and `mermaid` with `bun add`; load mermaid
   lazily. Render the user's content safely: the same sanitising the Mycelium renderer
   does, no raw HTML passthrough. [muse]
2. `apps/client/src/lib/artifacts.ts` — typed calls for the Bulle API.
3. `apps/client/src/routes/(app)/artifacts/+page.svelte` — the list for the current space,
   newest first, with a filter box, grouped by month; "connect Mycelium" empty state when
   `available` is false. `artifacts/[id]/+page.svelte` — the full viewer with delete.
4. `apps/client/src/lib/components/ArtifactCard.svelte` and `ArtifactPanel.svelte` — in
   `Message.svelte`, render one card per finished `artifact_create`/`artifact_update`
   call, outside the collapsible tool steps so it shows even when steps are hidden.
   Tapping opens the panel: a right-side panel on wide screens, a full-screen sheet below
   the breakpoint, with "Open in Artifacts" and close. The panel fetches content by id.
5. `apps/client/src/routes/(app)/+layout.svelte` — nav order: assistant, Ideas,
   Artifacts. Remove Jobs from the nav. Keep the jobs running dot somewhere visible:
   on the Settings/Profile menu entry, or drop it if muse offers no slot.
6. `apps/client/src/routes/(app)/account/[[section]]/+page.svelte` — a `jobs` section
   (label "Jobs", clock icon) that renders the existing jobs page content, moved into
   `lib/components/settings/JobsSettings.svelte`. `/jobs` redirects to `/account/jobs`.
7. `apps/client/src/lib/components/SessionList.svelte` (or the sidebar that hosts it) — a
   search field above the list; at 2+ characters it calls `/api/sessions/search`
   (debounced 250 ms) and shows results with title and snippet; Escape or clearing
   restores the list. Keyboard reachable; 44 px targets. [muse]
8. Tests (vitest): markdown utils parse alerts, charts, compare and metrics; the card
   extraction from tool calls; search debounce helper.

Exit for every Bulle track: `GOFLAGS=-tags=goolm go test ./...` in `apps/api`;
`env -u BULLE_ADDITIONAL_PROMPT XDG_CONFIG_HOME=<scratch> scripts/check.sh`;
`bun run check` (0 errors, 0 warnings) and `bun run test` in `apps/client`;
`filet check` with no errors.

## Order and parallel tracks

- First: merge the Ideas branches into Bulle `main` and ship Ideas, so these tracks start
  from a tree that has the Ideas migration, `ChannelIdeas` and the new nav entry.
- Then A, B, C and D run in parallel, each in its own worktree. B and C both touch
  `suite/tools.go`; keep C's change to the `case AppMycelium` line and B's to the prompt
  note, and merge B before C.
- Ship Mycelium first (Track A), then Bulle with B, C and D as one minor release.

## Files to modify or add

Mycelium: `internal/artifacts/{artifacts,format,store}.go`, `cmd/artifact.go`,
`internal/server/{artifacts_api,server}.go`, `internal/mcpserver/{tools,artifacts}.go`,
`apps/client/src/lib/backendTypes.ts`, `apps/client/src/routes/(app)/artifacts/**`,
`docs/{api,architecture}.md`, `README.md`, `CHANGELOG.md`, tests.

Bulle: delete `core/memory/`; `apps/api/internal/suite/{tools,tools_memory,mycelium}.go`,
new `suite/tools_artifacts.go`; new `runner/{memory_profile,memory_pass}.go`;
`internal/store/*` (channel, pass queries, search); `internal/scheduler/scheduler.go`;
new migration; new `modules/artifacts/*`; `modules/sessions` search route; `router.go`;
client files listed in Track D; `README.md`, `CHANGELOG.md`.

## Risks

- **The memory pass writes memory unattended.** Its only input is user and assistant
  text, and its only tools are memory tools, so an injected instruction in a mail cannot
  reach it, and the worst it can do is edit memory pages the user can read and fix in
  Mycelium. Pinned by tests.
- **Cost.** One pass per idle chat. Cap the transcript, cap 5 passes per tick, skip chats
  with no new messages.
- **Profile drift.** A wrong fact in `profile.md` follows every chat. The pass prompt says
  newest wins, and the user can edit the page; the prompt also tells the live agent to
  correct `profile.md` when the user corrects it.
- **Mycelium down.** Profile injection and artifact tools fail soft: the chat goes on
  without them and the tool answers with a plain error.
- **Old artifacts with `expires:`.** Parsing must ignore the field, not choke on it, and
  nothing may delete them. Pinned by test.
- **Renderer port.** Mycelium's renderer targets muse v0.7.4; Bulle is on v0.11.0. Expect
  small component API changes.
- **Release order.** Bulle's artifact tools need the Mycelium POST route in production.
  Ship and deploy Mycelium first.

## Skip, and why

- **A Bulle-side memory store or artifact store.** The user chose Mycelium only.
- **HTML artifacts.** Markdown only, by decision.
- **Editing artifacts in Bulle's UI.** Ask the agent, or edit in Mycelium.
- **A memory editor in Bulle.** Mycelium is the editor.
- **Vector search over chats.** A plain text match is enough for v1.
- **Lifting the renderer into `@facile/muse`.** Right long term, but it is a muse release
  and a Mycelium bump from v0.7.4. Do it as its own change; until then the renderer is
  copied, and the copy is noted in both repos' CHANGELOG.

## Sources

Consulted 2026-11-06.

- Claude Help Center, chat search and memory:
  https://support.claude.com/en/articles/11817273-use-claude-s-chat-search-and-memory-to-build-on-previous-context
- Claude Help Center, artifacts:
  https://support.claude.com/en/articles/9487310-what-are-artifacts-and-how-do-i-use-them
- Letta, sleep-time agents: https://docs.letta.com/guides/agents/architectures/sleeptime
- AWS, AgentCore long-term memory deep dive:
  https://aws.amazon.com/blogs/machine-learning/building-smarter-ai-agents-agentcore-long-term-memory-deep-dive/
- SmartScope, Meta Muse screen guide (Muse's MEMORY.md read at chat start and reorganised
  nightly): https://smartscope.blog/en/blog/meta-muse-screen-guide-features-main-chat-feed-2026
- Search results only, not read in full: https://www.glukhov.org/ai-systems/memory/memory-systems-in-ai-assistants/,
  https://mastra.ai/articles/long-term-memory-ai-agents, https://docs.letta.com/guides/core-concepts/memory/memory-blocks
