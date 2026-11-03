# Plan: attachments, Nuage files, Mycelium memory and Bulle spaces

Status: approved 2026-11-03, building.

## Goal

1. The composer gets a + button. You attach images and files to a message, and the model sees them.
2. The agent reads and saves the user's files in Nuage.
3. The agent has long-term memory in Mycelium.
4. Bulle spaces separate work contexts. Each space has its own memory, chats, jobs and default suite accounts.

## Decisions

- Memory lives in Mycelium, one Mycelium space per Bulle space. Memory writes skip the approval gate: they are the agent's own notes, they show in the tool list, and the user edits them in Mycelium.
- Nuage is the user's file store. The agent reaches it with delegated tokens, like Courrier today. `files_save` is approval-gated.
- Message attachments are stored in Bulle's Postgres (`attachments` table, bytea). Every later turn resends the bytes to the model, so fetching them from Nuage each turn would be slow and fragile. The agent can still copy an attachment into Nuage with `files_save`.
- One owner per space in v1, one persona across spaces.

## Why

- The web agent remembers nothing between conversations, cannot see files, and cannot reach Nuage.
- nacelle message parts are a closed set (`Text`, `Reasoning`, `ToolCall`, ...). It has no image or document part (`message.go`), so no backend can send an image to a model today.
- Mycelium already has spaces, per-space memory search (`POST /api/memory/search` with `space_id`) and per-space files (`/api/sync/files/*?space_id=`). Its `auth()` (`internal/server/auth.go`) accepts only its own opaque tokens and keys users by email (`internal/server/users.go`).
- Registre's delegated tokens carry `sub`, `aud`, `scope` and `act`, and no email (`modules/oidc/storage.go`, `mintAccessToken`).
- Nuage and Mycelium are on porte v0.5.2. The scope gate and machine-audience bearer path need porte v0.7.0 or later.

## Contracts (all tracks code against these)

### nacelle (v0.30.0)

```go
type Image struct {
	MediaType string // image/png, image/jpeg, image/gif, image/webp
	Data      []byte
}

type Document struct {
	Name      string
	MediaType string // application/pdf
	Data      []byte
}
```

Both are `Part`s, valid in user messages. The anthropic backend sends `image` and `document` blocks. The openai and openrouter backends send `image_url` data URIs, and for PDFs openrouter sends a `file` part with a data URI. A backend that cannot send a part returns a clear error rather than dropping it silently.

### Bulle attachments API

- `POST /api/attachments`: multipart field `file`. Answers `{id, name, media_type, size}`. Limit 20 MB per file. Accepted: png, jpeg, gif, webp, pdf, and text-like types (`text/*`, json, csv, xml, yaml, markdown, source code by extension). Anything else gets 415.
- `GET /api/attachments/{id}`: the bytes, owner only.
- `POST /api/sessions/{id}/messages` body becomes `{text, attachments?: [id]}`, max 10. Text may be empty when attachments are present.
- The stored user message keeps a wire part `{"type":"attachment","id","name","media_type"}`. The runner turns it back into `nacelle.Image`, `nacelle.Document`, or a `Text` block (`<file name="...">...</file>`) when it builds the history.
- `GET /api/sessions/{id}` messages carry `attachments: [{id, name, media_type, size}]`.

### Registre

- Tokens minted by the token-exchange grant carry the subject's verified `email`.
- Bulle's `token_exchange` gains Nuage (`9U40zs8fnTkpKpR6MMobIOXk0OkvotC7vd1NECGH`, scopes `nuage.read nuage.write`) and Mycelium (`mycelium`, scopes `mycelium.read mycelium.write`).

### Mycelium delegated access

- `auth()` also accepts a JWT that porte verifies for audience `mycelium`. It maps the `email` claim to the user. GET needs `mycelium.read`, anything else needs `mycelium.write`.
- A delegated caller may: create a space (`POST /api/spaces`, caller becomes owner), list its spaces, search memory and read and write files in spaces it is a member of.
- A delegated caller may never: touch the common tree (`space_id=""`), call admin routes, manage tokens, users or settings.

### Nuage delegated access

- porte v0.7.0 or later, `OIDC_MACHINE_AUDIENCE` = Nuage's client id, `OIDC_SCOPE_PREFIX=nuage`, wired the same way as Courrier.
- The track reports the exact routes Bulle uses: list a folder, search, download a file, upload a file, create a folder.

### Bulle spaces API

- `GET/POST /api/spaces`, `PUT/DELETE /api/spaces/{id}`. A space has `id, name, icon, mycelium_space_id, opus_workspace, agenda_calendar, courrier_account, created_at`.
- Sessions and jobs carry `space_id`. `POST /api/sessions` and job create accept `space_id`. `GET /api/sessions?space_id=` and `GET /api/jobs?space_id=` filter.
- Every user has a default space. Deleting the last space is refused.

## Tracks

Wave 1 runs in parallel. Each track commits on its own branch or worktree and does not push. The lead merges, runs the gates and deploys.

| Track | Repo | Work | Exit |
|---|---|---|---|
| N | nacelle | `Image` and `Document` parts, all backends, tests | `go test ./...` and filet pass, each backend's request JSON tested |
| U | Nuage | porte bump, machine audience, scope prefix, route list | a JWT for audience Nuage with `nuage.read` lists files, POST without `nuage.write` gets 403, existing login unchanged |
| R | Registre | email claim on exchanged tokens, seed entries | seed test pins the six apps, exchange test sees `email` |
| M | Mycelium | delegated JWT auth with the limits above | tests: own space ok, other space 403, common tree 403, admin 403 |
| S | Bulle (worktree) | spaces: migration, module, runner defaults, client switcher | two spaces with separate chat lists, job in a space uses its defaults |
| C | Bulle (worktree) | client attachments: + button, picker, drag and drop, paste, chips with remove, upload progress, previews in sent messages | client tests and svelte-check pass against the contract |

Wave 2, after N, U, R and M:

| Track | Repo | Work | Exit |
|---|---|---|---|
| A | Bulle | attachments API and runner parts, Nuage tools (`files_list`, `files_search`, `files_read`, `files_save`), memory tools (`memory_search`, `memory_read`, `memory_write`) scoped to the session's space | API tests pass, a fake world covers each tool |

Then the lead merges S, C and A, runs `scripts/check.sh`, `filet check`, `bun run test`, svelte-check and vite build, deploys Registre, Mycelium and Nuage, and ships Bulle.

## Risks

- The Mycelium common tree is the owner's personal dev wiki. A delegated token must never read it. This is the security-critical line, and it gets its own test.
- The spaces migration touches every session and job: backfill and NOT NULL happen in one migration that can be rerun safely.
- Registre, Mycelium and Nuage each deploy production on push to main. The Mycelium push does not wait for CI.
- Another session may be editing Bulle at the same time. Stage explicit paths only.

## Skip

- Spaces shared between users.
- Separate personas per space.
- A memory editor inside Bulle.
- Audio and video attachments.
