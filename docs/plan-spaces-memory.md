# Plan: Mycelium memory and Bulle spaces

Status: approved 2026-11-03, not started.

## Goal

Give the Bulle agent long-term memory stored in Mycelium. Then add Bulle spaces: separate work contexts, each with its own memory, files, default suite accounts, chats and jobs.

## Decisions

- Memory lives in Mycelium, one Mycelium space per Bulle space.
- Memory writes do not go through the approval gate. They are the agent's own notes, they show in the tool list, and the user edits them in Mycelium.
- Space files live in a Nuage folder per space.
- One owner per space in v1. One assistant persona across spaces.

## Why

- The web agent remembers nothing between conversations. Nothing in `apps/api` stores or recalls facts.
- Mycelium already has spaces with members (`internal/server/spaces.go`), memory search per space (`POST /api/memory/search` with `space_id`) and file read/write per space (`/api/sync/files/*?space_id=`).

## Approach

A Bulle space is a thin record of pointers to things the suite already has: one Mycelium space (memory), one Nuage folder (files), and optional defaults (an Opus workspace, an Agenda calendar, a Courrier account). Every session and job belongs to one space. The runner gives the agent only that space's memory and defaults. Mycelium and Nuage are reached the same way as Courrier, Agenda, Opus and Sablier today: per-user delegated tokens from Registre (`internal/suite`).

## Findings that shape track A

- Mycelium's `auth()` (`internal/server/auth.go`) accepts only its own opaque tokens, looked up by hash. It does not accept suite JWTs.
- Mycelium keys users by email (`internal/server/users.go`, `.users.json`) and stores no OIDC subject.
- Registre's delegated tokens carry `sub`, `aud`, `scope` and `act`, but no email (`modules/oidc/storage.go`, `mintAccessToken`). So either Registre adds a verified `email` claim to exchanged tokens, or Mycelium records the subject at OIDC login. The Registre claim is the general fix.
- Mycelium is on porte v0.5.2. The scope gate needs porte v0.7.0 or later.
- A push to Mycelium main deploys production with no CI wait. The pre-push hook is the only gate.

## Steps

### Track A: Mycelium accepts suite tokens (Mycelium, Registre)

1. Registre: add a verified `email` claim to tokens minted by the token-exchange grant. Test that a plain access token is unchanged.
2. Registre `seed.yaml`: add a Mycelium entry to Bulle's `token_exchange` with `mycelium.read` and `mycelium.write`, and update the seed test.
3. Mycelium `internal/server/auth.go`: `auth()` also accepts a porte-verified JWT whose audience is Mycelium. It maps the `email` claim to the user. GET needs `mycelium.read`, anything else needs `mycelium.write`.
4. Mycelium: refuse delegated tokens on the common tree (`space_id=""`) and on admin routes. The common tree is the owner's dev wiki.
5. Exit: a delegated token reads and writes the user's own space. It gets 403 on another user's space, on the common tree, and on admin routes.

### Track B: memory tools in Bulle (after A)

6. `apps/api/internal/suite/config.go`: add `AppMycelium` (`SUITE_MYCELIUM_URL`, `SUITE_MYCELIUM_AUDIENCE`). Forward both in `docker-compose.yml`.
7. `apps/api/internal/suite/mycelium.go`, `tools_mycelium.go`: `memory_search`, `memory_read`, `memory_write`. These are not in `WriteTools()`, so the gate does not apply. On first use they create a personal space named "Bulle (name)".
8. `apps/api/internal/runner/personality.go`: tell the agent to search memory before answering about the user, and to write durable facts.
9. Exit: a fact told in one chat is recalled in a new chat, and the page is visible in `mycelium memory search` and the Mycelium UI.

### Track C: Bulle spaces (in parallel with B; memory part needs B)

10. `apps/api/migrations/000NN_spaces.sql`: `spaces (id, user_id, name, icon, mycelium_space_id, nuage_folder_id, opus_workspace, agenda_calendar, courrier_account, created_at)`, `space_id` on `sessions` and `jobs`, a default space per existing user, rows backfilled, then NOT NULL, all in one migration that can be rerun safely.
11. `apps/api/modules/spaces/`: list, create, rename, edit defaults, delete, with `documentation.go`. Mutations need `X-Facile-CSRF`.
12. `apps/api/internal/runner`: pass the session's space to the suite tools. The space's defaults fill in when the agent omits an account, calendar or workspace. Memory goes to the space's Mycelium space.
13. `apps/client`: a space switcher in the sidebar (muse `SpaceSwitcher`), chats and jobs filtered by space, and a space settings page with pickers for the defaults.
14. Exit: two spaces have separate memory and chat lists, and a job in one space files Opus tasks into that space's workspace without being told which one.

### Track D: space files (after C)

15. Registre: add Nuage to `token_exchange`. Check that Nuage reads `OIDC_MACHINE_AUDIENCE` and `OIDC_SCOPE_PREFIX`.
16. `apps/api/internal/suite/nuage.go`: `files_list`, `files_read`, `files_save`, limited to the space's folder. `files_save` is in `WriteTools()`.
17. Client: drop a file into the composer to upload it into the space's Nuage folder, and the agent sees it as an attachment.
18. Exit: drop a PDF, ask about it, and the agent reads it. The file shows in Nuage.

## Exit criteria

- Each track's exit holds.
- `filet check` exits 0, `scripts/check.sh` passes, `bun run test` and svelte-check pass.
- One test per boundary: no access to another user's space, no access to the common tree, the space's defaults get used.

## Risks

- The Mycelium common tree is the owner's personal dev wiki. A delegated token must never read it. This is the security-critical line.
- The spaces migration touches every existing session and job.
- Track A deploys the SSO (Registre) and Mycelium production.

## Skip

- Spaces shared between users (Mycelium supports members, v1 does not use them).
- Separate personas per space.
- A memory editor inside Bulle.
- Attachments stored in Bulle itself.
