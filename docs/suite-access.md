# Suite access

The agent reads the signed-in user's own Courrier mail, Agenda events, Opus tasks, Sablier timer, Nuage files and Mycelium memory. It does this in chat and in background jobs. The design and its reasons are in `plan-suite-delegation.md`. This page is the operating reference.

## How a call is authorised

1. The user signs in to Bulle through Registre. porte stores their access and refresh token in `porte_identities`.
2. A suite tool asks `suite.Tokens` for a token for one app.
3. On a cache miss, `Tokens` refreshes the user's Registre token if it is about to expire and writes the rotated pair back.
4. `Tokens` calls Registre's token endpoint with RFC 8693 token exchange: audience is the app's client id, scope is `<app>.read`.
5. Registre returns a JWT that lives five minutes. `Tokens` caches it in memory until 30 seconds before it expires.
6. The tool calls the app's API with that JWT as a bearer token. The app verifies it with porte and maps `sub` to its own user.

Bulle stores no app credential. Nothing is written to disk, to the environment of a process or to the VM.

## Configuration

An app's tools exist only when both its URL and its audience are set. Setting one without the other stops the API at boot.

| Variable | Value |
|---|---|
| `SUITE_COURRIER_URL`, `SUITE_AGENDA_URL`, `SUITE_OPUS_URL`, `SUITE_SABLIER_URL`, `SUITE_NUAGE_URL`, `SUITE_MYCELIUM_URL` | App origin, no trailing slash needed. |
| `SUITE_COURRIER_AUDIENCE`, `SUITE_AGENDA_AUDIENCE`, `SUITE_OPUS_AUDIENCE`, `SUITE_SABLIER_AUDIENCE`, `SUITE_NUAGE_AUDIENCE`, `SUITE_MYCELIUM_AUDIENCE` | The app's client id in Registre's `seed.yaml`. Nuage is `9U40zs8fnTkpKpR6MMobIOXk0OkvotC7vd1NECGH`, Mycelium is `mycelium`. |

`OIDC_CLIENT_ID` and `OIDC_CLIENT_SECRET` authenticate Bulle to the token endpoint. The `bulle` client in Registre needs a `token_exchange` entry for each audience.

## Endpoints called

Each client keeps its paths in one constant block at the top of its file in `apps/api/internal/suite/`.

| App | Paths (GET) |
|---|---|
| Courrier | `/api/accounts/`, `/api/accounts/{id}/mail/folders/inbox/emails`, `/api/accounts/{id}/mail/search` |
| Agenda | `/api/calendars/`, `/api/calendars/{id}/events?from=&to=&tz=` (RFC 3339 bounds computed in Go from the user's timezone) |
| Opus | `/api/workspace`, `/api/project?workspaceId=`, `/api/task/me?workspaceId=` (default), `/api/task/tasks/{projectId}` (when a project is named) |
| Sablier | `/api/time-entries/running` |
| Nuage | `/api/files?folder_id=`, `/api/folders?parent_id=`, `/api/search?q=&limit=`, `/api/files/{id}`, `/api/files/{id}/download`, POST `/api/files` (multipart, `origin_app=bulle`) |
| Mycelium | `/api/spaces` (GET, POST), POST `/api/memory/search`, GET and PUT `/api/sync/files/memory/{path}?space_id=` |

## Memory

`memory_search`, `memory_read` and `memory_write` work on the `memory/` directory of one Mycelium space. A Bulle space with a `mycelium_space_id` uses that space. Otherwise the tools use the user's personal space named "Bulle": the first call creates it in Mycelium (or adopts an existing one the user owns with that name) and `mycelium_spaces` remembers its id. The common tree is never used: an empty space id fails before any call. `memory_write` is not approval-gated (`suite.SelfWrites()`); `files_save` is.

## Failure states

| State | Cause | What the user sees |
|---|---|---|
| `reconnect` | The Registre refresh token is missing or was rejected. | "Sign out of Bulle and sign in again." |
| `no_account` | The app answered 401 for a fresh token: the user never signed in there. | A link to open the app once. |
| `unavailable` | The app is down or timed out. | A retry later message. |

`GET /api/suite/status` reports one state per configured app and drives the "Connected apps" section on the Account page. Signing out of Bulle and a back-channel logout both clear the user's cached tokens.
