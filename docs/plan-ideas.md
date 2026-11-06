# Plan: Ideas, Bulle's proactive suggestions

Status: approved 2026-11-06. Written against v3.22.1.

## Goal

A new main section in the sidebar, after Jobs: **Ideas**. Each morning Bulle looks at
the space's mail, calendar, tasks, memory and goals, and leaves a short batch of
suggestions: what it could do for you now, and why. Tapping one starts a chat that does
it. Bulle never acts on its own from this page; every write still goes through the
approval card.

Modelled on Meta Muse's Ideas and Goals tabs and on ChatGPT Pulse, cut down to what
Bulle can already do.

## Decisions (from the user, 2026-11-06)

| Question | Answer |
|---|---|
| What an idea is | Suggestion plus prep: a card with a reason and a ready prompt. Bulle may pre-draft (a reply, an agenda) read-only. Nothing is sent or changed until the user starts the chat and approves. |
| Cadence | One daily batch per space, plus a "Think of more" button. |
| Spaces | Per space. Ideas follow the space switcher like chats and jobs. |
| Extras | A morning push when the batch is ready, and a Goals list that steers the generator. |

## What good practice says, and how it lands here

| Practice | Source | In Bulle |
|---|---|---|
| Avoid the blank page; give starting points | Muse screen guide | Ideas page is the starting point; each card carries a ready prompt. |
| A handful of cards, not an endless feed; "that's all for today" | ChatGPT Pulse | 3 to 7 ideas per batch, hard cap 7. Explicit end-of-batch line. |
| Steer in natural language, plus thumbs up/down | Pulse "curate", Muse feed | A steering box per space and thumbs on every card. Feedback feeds the next batch. |
| Ideas expire; stale suggestions are worse than none | Pulse, FarVision | A batch's unsaved ideas expire when the next batch lands or after 3 days. Saved ones stay. |
| Explain why | FarVision | Every card has a one-line "why now" and the sources it came from. |
| Least intrusive surface; start quiet | FarVision | Pull-first page. One push per day at most, only when the batch has ideas, and only if the user turned it on. |
| Easy recovery: dismiss, snooze, correct | FarVision | Dismiss with an optional reason, Save, thumbs. A dismissed idea's fingerprint is suppressed for 14 days. |
| Separate drafting from acting | FarVision | Generation runs read-only. Drafts live in the card as text. Acting happens in a normal chat behind the gate. |
| Time suggestions at boundaries, not mid-task | Kuo et al., IUI'26 field study | Daily morning batch, a natural boundary. No mid-day interruptions in v1. |
| Measure trust, not clicks | FarVision | Track started, dismissed, thumbs down per batch (counts only), for tuning. |
| Goals give the agent background context | Muse Goals | A Goals list per space with Muse's seven categories, read by the generator. |

## Security model

The generator reads untrusted input (mail) and private data. By the Rule of Two it must
then not change state. So the ideas run:

- Gets **only read-only tools**: every suite tool whose `IsReadOnly()` is true
  (mail, calendar, tasks, files, memory reads, `spaces_list`), plus one host tool,
  `idea_propose`. There are no VM tools, no job tools, no `memory_write`, no web tools and
  no outward tools, so the gate never fires and nothing waits for an approval nobody
  is there to give.
- Writes nothing outside Bulle's own `ideas` table, capped at 7 rows per batch.
- Produces text that the user reads before anything happens. "Do it" sends the card's
  prompt as the user's own first message in a normal web session, where the approval
  gate applies as usual. The card shows the full prompt before the user taps.
- Card text renders as plain text in the client, never as HTML.

## Data model

Migration `apps/api/migrations/00020_ideas.sql` (goose, sequential):

```sql
CREATE TABLE ideas (
  id           TEXT PRIMARY KEY,
  user_id      BIGINT NOT NULL,
  space_id     TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
  batch_id     TEXT NOT NULL,
  kind         TEXT NOT NULL,
  title        TEXT NOT NULL,
  why          TEXT NOT NULL,
  prompt       TEXT NOT NULL,
  draft        TEXT NOT NULL DEFAULT '',
  sources      JSONB NOT NULL DEFAULT '[]',
  fingerprint  TEXT NOT NULL,
  status       TEXT NOT NULL DEFAULT 'new',
  feedback     TEXT NOT NULL DEFAULT '',
  reason       TEXT NOT NULL DEFAULT '',
  session_id   TEXT NOT NULL DEFAULT '',
  created_at   TIMESTAMPTZ NOT NULL,
  updated_at   TIMESTAMPTZ NOT NULL,
  expires_at   TIMESTAMPTZ NOT NULL
);
CREATE INDEX ideas_space_status ON ideas (user_id, space_id, status);
CREATE INDEX ideas_fingerprint ON ideas (space_id, fingerprint);

CREATE TABLE idea_settings (
  space_id        TEXT PRIMARY KEY REFERENCES spaces(id) ON DELETE CASCADE,
  user_id         BIGINT NOT NULL,
  enabled         BOOLEAN NOT NULL DEFAULT TRUE,
  hour            INT NOT NULL DEFAULT 8,
  timezone        TEXT NOT NULL DEFAULT 'UTC',
  push            BOOLEAN NOT NULL DEFAULT FALSE,
  steering        TEXT NOT NULL DEFAULT '',
  last_batch_date TEXT NOT NULL DEFAULT '',
  updated_at      TIMESTAMPTZ NOT NULL
);

CREATE TABLE idea_batches (
  id         TEXT PRIMARY KEY,
  user_id    BIGINT NOT NULL,
  space_id   TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
  run_id     TEXT NOT NULL DEFAULT '',
  trigger    TEXT NOT NULL,
  status     TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  finished_at TIMESTAMPTZ
);

CREATE TABLE goals (
  id         TEXT PRIMARY KEY,
  user_id    BIGINT NOT NULL,
  space_id   TEXT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
  title      TEXT NOT NULL,
  category   TEXT NOT NULL,
  done       BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);
```

Check the `spaces.id` column type and the existing FK style in earlier migrations before
writing this; match them. Every store query is scoped by `user_id`.

Values:

- `ideas.kind`: `mail`, `task`, `event`, `follow_up`, `research`, `other`.
- `ideas.status`: `new`, `saved`, `done`, `dismissed`, `expired`.
- `ideas.feedback`: `''`, `up`, `down`.
- `ideas.sources`: `[{"label": "Email from Ann", "ref": "courrier:5/email:812"}]`, at most 5.
- `idea_batches.trigger`: `daily`, `manual`. `status`: `running`, `finished`, `error`.
- `goals.category`: `health`, `relationships`, `finance`, `career`, `interests`,
  `productivity`, `other`.

Limits, enforced server-side: title 120 chars, why 280, prompt 2000, draft 4000,
steering 500, goal title 120, at most 20 open goals per space, reason 200.

## HTTP API (contract shared by backend and client)

All routes behind `requireAuth`, under `/api`, JSON. Every id lookup is scoped to the
caller; another user's id answers 404.

```
GET    /ideas?space_id=S
  200 {
    "ideas":    [Idea],          // status new or saved, not expired, newest batch first
    "goals":    [Goal],          // open first, then done
    "settings": IdeaSettings,
    "generating": bool,          // a batch for S is running
    "last_batch_at": string|null // RFC 3339
  }

PATCH  /ideas/{id}            body {"status"?: "saved"|"new"|"dismissed", "feedback"?: ""|"up"|"down", "reason"?: string}
  200 Idea

POST   /ideas/{id}/start
  200 {"session_id": string}  // creates a web session in the idea's space titled with the
                              // idea title, starts a run whose user message is the prompt
                              // (plus "\n\nDraft:\n"+draft when draft is set), marks the idea
                              // done and stores session_id

POST   /ideas/generate        body {"space_id": S}
  202 {"batch_id": string}
  409 {"error": "a batch is already running"}
  429 {"error": "..."}        // more than 3 manual batches for S today (user's timezone)

PUT    /ideas/settings        body {"space_id": S, "enabled"?: bool, "hour"?: 0-23,
                                    "timezone"?: IANA name, "push"?: bool, "steering"?: string}
  200 IdeaSettings

POST   /ideas/goals           body {"space_id": S, "title": string, "category": string}
  201 Goal
PUT    /ideas/goals/{id}      body {"title"?: string, "category"?: string, "done"?: bool}
  200 Goal
DELETE /ideas/goals/{id}
  204
```

Shapes:

```ts
type Idea = {
  id: string; space_id: string; batch_id: string;
  kind: 'mail'|'task'|'event'|'follow_up'|'research'|'other';
  title: string; why: string; prompt: string; draft: string;
  sources: { label: string; ref: string }[];
  status: 'new'|'saved'|'done'|'dismissed'|'expired';
  feedback: ''|'up'|'down';
  session_id: string;
  created_at: string; expires_at: string;
};
type Goal = { id: string; space_id: string; title: string;
  category: 'health'|'relationships'|'finance'|'career'|'interests'|'productivity'|'other';
  done: boolean; created_at: string };
type IdeaSettings = { space_id: string; enabled: boolean; hour: number;
  timezone: string; push: boolean; steering: string };
```

`GET /ideas` with no settings row returns the defaults (`enabled: true, hour: 8,
timezone: "UTC", push: false, steering: ""`) without writing one. The client sends the
browser's `Intl.DateTimeFormat().resolvedOptions().timeZone` with its first `PUT`, and
on page load when the stored timezone is still `UTC` and the browser's is not.

## Backend

Package layout follows the existing modules (`modules/jobs` is the model).

1. `apps/api/migrations/00020_ideas.sql`: the schema above.
2. `apps/api/internal/store`: `Idea`, `Goal`, `IdeaSettings`, `IdeaBatch` types and an
   `Ideas` interface embedded in `Store`, implemented for Postgres and for the in-memory
   store used by tests (`store.NewMemory`). Methods: list active ideas for a space,
   get/update an idea, create ideas in a batch, expire a space's previous `new` ideas,
   recent feedback (last 30 rated or dismissed ideas), fingerprints suppressed (dismissed
   in the last 14 days, or currently active), settings get/upsert, list settings due
   (enabled rows), mark batch date, batch create/finish/running check, count manual
   batches since a time, goals CRUD.
3. `apps/api/internal/store`: add `ChannelIdeas = "ideas"`. Sessions on this channel
   never appear in the session list or the sidebar (check how `ListSessions` filters
   `ChannelCron` and do the same).
4. `apps/api/internal/runner/ideas.go`: `(*Runner).GenerateIdeas(ctx, userID, spaceID,
   trigger string) (batchID string, err error)`. It refuses when a batch is running
   (`ErrIdeasBusy`), creates the batch row and a `ChannelIdeas` session, and starts a
   run with a built prompt. In `buildAgent`, `session.Channel == store.ChannelIdeas`
   routes to `ideasAgent`, built like `guestAgent`: the user's backend and persona, the
   suite tools filtered to `IsReadOnly()`, and the `idea_propose` tool. No VM, job,
   web or write tool. A small `MaxIterations` (about 24).
5. `idea_propose` input: `{kind, title, why, prompt, draft?, sources?, fingerprint}`. It
   validates kinds and limits, refuses a fingerprint already active or suppressed (it
   answers the model "already suggested, pick something else"), and refuses beyond 7
   per batch. Rows get `expires_at = now + 72h`.
6. The generator prompt, in `ideas.go`: today's date in the space's timezone, the space
   name, the open goals, the steering text, the last 30 feedback items (title, thumbs,
   status, reason), the active fingerprints, and these rules: 3 to 7 ideas, each one
   something Bulle can do with its tools in a chat; a concrete "why now" grounded in
   something it read; prefer what is due, waiting on the user, or moves a goal; never
   repeat a fingerprint; drafts are text only; no idea that only says "check your
   inbox"; fingerprints are stable keys like `courrier:5/email:812/reply` or
   `opus:task:44/due`; when nothing is worth suggesting, propose nothing.
7. When the run ends (watch it as `scheduler.trackRun` does, through
   `Runner.Subscribe`): finish the batch, expire the space's older `new` ideas from
   earlier batches, and when `trigger == daily`, `settings.push` is on and the batch
   produced at least one idea, send one push: title "Bulle", body "N ideas for SPACE",
   path `ideas`.
8. `apps/api/internal/scheduler/scheduler.go`: in `Tick`, after jobs, list enabled idea
   settings; for each, compute the local time in its timezone; when the local hour
   equals `hour`, and `last_batch_date` is not today's local date, mark the date and
   call `GenerateIdeas(..., "daily")`. The same advisory lock already covers it. Spaces
   with no settings row use the defaults: create the row lazily the first time the
   user opens the Ideas page, so spaces nobody opened never generate.
9. Also in `Tick`: expire `new` ideas past `expires_at`.
10. `apps/api/modules/ideas`: handler, router, types, documentation, like
    `modules/jobs`. Register the routes in `apps/api/router.go` next to jobs. Manual
    generate: 429 past 3 manual batches today. `start` uses the same session creation
    and `Runner.Start` path the sessions module uses for a web chat.
11. Space deletion already cascades by FK. Check `DeleteSpace` in the memory store and
    make it drop ideas, settings, batches and goals too.

## Client

1. `apps/client/src/lib/ideas.ts`: typed API calls for the contract above, using the
   same fetch helper as `$lib/jobs` (CSRF header and all).
2. `apps/client/src/lib/ideas-activity.svelte.ts`: a small store with the current space's
   count of `new` ideas, refreshed on space change and after actions, for the sidebar dot.
3. `apps/client/src/routes/(app)/+layout.svelte`: add the nav entry after Jobs:
   `{ href: '/ideas', label: 'Ideas', icon: <a muse lightbulb or sparkles icon>,
   active: isActive('/ideas'), dot: newIdeas > 0 }`.
4. `apps/client/src/routes/(app)/ideas/+page.svelte`, muse components only, tokens not
   hex, Svelte 5 runes, mobile-first at 360px, hit targets at least 44px, no hover-only
   affordance:
   - Header "Ideas" with a "Think of more" button (spinner and disabled while
     `generating`; a toast on 409 or 429) and a settings icon button.
   - Steering: a one-line input "What should Bulle think about?" that saves on blur or
     enter to `steering`.
   - Goals: a collapsible section. Open goals as rows with a done checkbox and delete;
     an add row with a category select (the seven categories) and a title.
   - Ideas: cards, newest batch first, saved ideas in their own "Saved" group. Each card
     shows a kind icon, title, why, and source chips. Expanding shows the full prompt
     and the draft in a read-only block. Actions: "Do it" (primary, calls `start`, then
     navigates to `/s/{session_id}`), Save/Unsave, thumbs up, thumbs down, Dismiss (with
     an optional reason: not relevant, already done, wrong time).
   - End of list line: "That's all for now." plus the last batch time.
   - Empty states: never generated ("Bulle will leave ideas here each morning", with a
     "Think of ideas now" button); generating; nothing new.
   - Settings modal: enabled switch, morning hour select, push switch.
   - Poll `GET /ideas` every 5 s while `generating` is true, stop when false.
   - All text renders as text, never `{@html}`.
5. Tests: vitest for any pure helper (grouping, expiry filtering, dot count).

## Tests (backend)

- Store: CRUD scoped by user; another user's idea, goal or settings are invisible.
- `idea_propose`: validates kind and limits; refuses the 8th idea in a batch; refuses an
  active or recently dismissed fingerprint.
- `ideasAgent`: every tool is read-only or `idea_propose`; no tool name matches the
  gate's outward set, so a background run never blocks on an approval.
- `GenerateIdeas`: 409-equivalent while a batch runs; the batch expires older `new`
  ideas, keeps `saved` ones.
- Scheduler: due at the local hour in the space's timezone, once per local day; disabled
  settings never fire.
- Handler: 404 on another user's idea; `start` creates a web session in the idea's space
  and marks it done; manual generate past 3 a day answers 429.
- Push: sent once for a daily batch with ideas and push on; not for manual or empty
  batches.

## Release

One minor version, `v3.23.0`. Changelog under Added. Gates before tagging: `go test ./...`
and `scripts/check.sh` (run as `env -u BULLE_ADDITIONAL_PROMPT XDG_CONFIG_HOME=<scratch>
scripts/check.sh`), `bun run check` and `bun run test` in `apps/client`, `filet check`
with no errors.

## Skip, and why

- **Event-driven ideas during the day.** The user picked a daily batch. Mid-task
  suggestions were dismissed 62% of the time in the IUI'26 field study.
- **Overnight research per idea (full Pulse).** Costs the most and needs web tools in an
  unattended run; the user picked suggestion plus prep.
- **Acting from the card.** Every action goes through a chat and the approval gate.
- **Web search in the generator.** It is a paid, outward tool; v1 grounds ideas in the
  user's own suite data. Revisit with a per-batch quota.
- **A Feed tab and a Library tab.** Muse has them; Bulle's chats and Nuage already hold
  outputs. Out of scope.
- **Writing goals to Mycelium memory.** Goals live in Bulle's DB so the page can edit
  them directly; the generator reads them from there.

## Sources

Consulted 2026-11-06.

- SmartScope, Meta Muse screen guide (Feed, Ideas, Goals, Library):
  https://smartscope.blog/en/blog/meta-muse-screen-guide-features-main-chat-feed-2026
- Meta, Muse product page: https://ai.meta.com/muse (search snippet only)
- Meta, Introducing Muse: https://about.fb.com/news/2026/09/introducing-muse-personal-ai-agent
  (search snippet only)
- CNN, Muse hands-on: https://www.cnn.com/2026/09/23/tech/meta-muse-ai-agent (search snippet only)
- OpenAI, Introducing ChatGPT Pulse: https://openai.com/index/introducing-chatgpt-pulse/
- FarVision AI Academy, Proactive Assistant Design:
  https://academy.farvision.com/intermediate/proactive-assistant-design
- Kuo et al., Developer Interaction Patterns with Proactive AI (IUI'26):
  https://arxiv.org/abs/2601.10253 (the HTML full text at
  https://arxiv.org/html/2601.10253v1 failed to fetch; abstract only)
- Meta, Agents Rule of Two: https://ai.meta.com/blog/practical-ai-agent-security/
  (cited in docs/plan-multi-space-agent.md)
