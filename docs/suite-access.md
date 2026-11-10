# Suite access

The agent reads the signed-in user's own Courrier mail, Agenda events, Opus tasks, Sablier timer, Nuage files, Mycelium memory and Glouton prospects. It does this in chat and in background jobs. The design and its reasons are in `plan-suite-delegation.md`. This page is the operating reference.

## How a call is authorised

1. The user signs in to Bulle through Registre. porte stores their access and refresh token in `porte_identities`.
2. A suite tool asks `suite.Tokens` for a token for one app.
3. On a cache miss, `Tokens` refreshes the user's Registre token if it is about to expire and writes the rotated pair back.
4. `Tokens` calls Registre's token endpoint with RFC 8693 token exchange: audience is the app's client id, scope is `<app>.read`.
5. Registre returns a JWT that lives five minutes. `Tokens` caches it in memory until 30 seconds before it expires.
6. The tool calls the app's API with that JWT as a bearer token. The app verifies it with porte and maps `sub` to its own user.

Bulle stores no app credential. Nothing is written to disk, to the environment of a process or to the VM.

## Apps reached through MCP

Courrier has no built-in client any more: Bulle is an MCP client of Courrier's own server at `/mcp` (go-sdk, stateless 2026-07-28) and mounts its tools as the server names them: `courrier_list_accounts`, `courrier_search_mail`, `courrier_read_thread`, `courrier_draft`, `courrier_send`, `courrier_organize` and `courrier_delete_draft`. The client is `suite.MCP` on nacelle's `mcp/client`.

- **Token.** The same RFC 8693 exchange, with `resource=<SUITE_COURRIER_MCP_RESOURCE>` and no audience, scope `courrier.read courrier.write` (read only if Registre refuses write). Registre answers a token whose `aud` is the resource, so it works at that server only. Bulle's existing `token_exchange` grant for Courrier covers it. A 401 drops the cached token and mints a new one once.
- **Connection.** One connection per server for the whole process, opened lazily by the first run that needs it, under that user's context: the server is stateless, so each request carries the token of the user whose call sends it (nacelle `Remote.Token`), and private tool lists are cached per user (`Remote.Identity`). A server that cannot be reached is left out of the run and retried after 30 s; the other apps keep working.
- **Confirmations.** Courrier asks before `courrier_send` and `courrier_delete_draft` through MRTR `input_required`. Bulle shows that question as an approval card (`args.kind = "mcp"`: server, tool, the server's message, the call's arguments) on the call's own id and returns the owner's answer to the server. A denial, an unanswered card or a run that ends declines, and Courrier then sends nothing. Bulle keeps no per-tool policy for MCP tools and no card can edit the mail. A call, its confirmation included, has 15 minutes, which is how long Courrier keeps the confirmation open.
- **Taint.** A read-only tool the server marks open world (`courrier_search_mail`, `courrier_read_thread`) is an untrusted source, like web pages: once one returns, the shell, memory deletes and `profile.md` writes ask first.
- **History.** Stored transcripts with the old built-in names (`courrier_list_inbox`, `courrier_search`, `courrier_get_email`, `courrier_save_draft`, `courrier_mark`, `courrier_update`, `courrier_move`) replay under the MCP name that does the same job (`suite.CanonicalTool`), read only.
- **Dropped.** `assistant_seen_at` on listed mail and the `mail_seen` bookkeeping behind it. The table stays until a later migration drops it.

## Configuration

An app's tools exist only when both its URL and its audience are set. Setting one without the other stops the API at boot. An MCP app needs only its server URL.

| Variable | Value |
|---|---|
| `SUITE_COURRIER_MCP_URL` | Courrier's MCP endpoint, `https://courrier.facile.studio/mcp`. Empty switches mail off. |
| `SUITE_COURRIER_MCP_RESOURCE` | The resource Registre declares for that server. Defaults to `SUITE_COURRIER_MCP_URL`; set it only when Bulle dials an internal address. |
| `SUITE_AGENDA_URL`, `SUITE_OPUS_URL`, `SUITE_SABLIER_URL`, `SUITE_NUAGE_URL`, `SUITE_MYCELIUM_URL`, `SUITE_GLOUTON_URL` | App origin, no trailing slash needed. |
| `SUITE_AGENDA_AUDIENCE`, `SUITE_OPUS_AUDIENCE`, `SUITE_SABLIER_AUDIENCE`, `SUITE_NUAGE_AUDIENCE`, `SUITE_MYCELIUM_AUDIENCE`, `SUITE_GLOUTON_AUDIENCE` | The app's client id in Registre's `seed.yaml`. Nuage is `9U40zs8fnTkpKpR6MMobIOXk0OkvotC7vd1NECGH`, Mycelium is `mycelium`, Glouton is `1IdqK3bA2LWI-zuWVLqFpnIlCTVu117F` (scopes `glouton.read`, `glouton.write`). |

`OIDC_CLIENT_ID` and `OIDC_CLIENT_SECRET` authenticate Bulle to the token endpoint. The `bulle` client in Registre needs a `token_exchange` entry for each audience.

## Endpoints called

Each client keeps its paths in one constant block at the top of its file in `apps/api/internal/suite/`.

| App | Paths (GET) |
|---|---|
| Courrier | POST `/mcp` only (see above) |
| Agenda | `/api/calendars/`, `/api/calendars/{id}/events?from=&to=&tz=` (RFC 3339 bounds computed in Go from the user's timezone) |
| Opus | `/api/workspace`, `/api/project?workspaceId=`, `/api/task/me?workspaceId=` (default), `/api/task/tasks/{projectId}` (when a project is named) |
| Sablier | `/api/time-entries/running` |
| Nuage | `/api/files?folder_id=`, `/api/folders?parent_id=`, `/api/search?q=&limit=`, `/api/files/{id}`, `/api/files/{id}/download`, POST `/api/files` (multipart, `origin_app=bulle`) |
| Mycelium | `/api/spaces` (GET, POST), POST `/api/memory/search`, GET and PUT `/api/sync/files/memory/{path}?space_id=` |
| Glouton | `/api/agent/contacts?q=&stage=&temperature=&limit=`, `/api/agent/companies/{id}`, `/api/agent/people/{id}`, `/api/agent/outreach?companyId=&personId=`, `/api/agent/suggestions?status=`; writes: POST `/api/agent/companies`, `/api/agent/people`, `/api/agent/drafts`, `/api/agent/suggestions`, PATCH `/api/agent/companies/{id}`, `/api/agent/people/{id}`, `/api/agent/drafts/{id}` |

## Memory

`memory_search`, `memory_read` and `memory_write` work on the `memory/` directory of one Mycelium space. A Bulle space with a `mycelium_space_id` uses that space. Otherwise the tools use the user's personal space named "Bulle": the first call creates it in Mycelium (or adopts an existing one the user owns with that name) and `mycelium_spaces` remembers its id. The common tree is never used: an empty space id fails before any call. `memory_write` is not approval-gated (`suite.SelfWrites()`); `nuage_save` is.

Every web and matrix run starts with the space's `profile.md` in its system prompt, cut at 4 KB. The runner reads it with a 3 second limit; a missing page or a Mycelium error leaves it out and the run goes on.

A chat that has been quiet for 10 minutes and holds messages no pass has read gets a memory pass (`scheduler.tickMemory`, at most 5 per tick, only for users whose Mycelium answers, only chats changed in the last 7 days). The pass is a run on a hidden `memory` session in the same space. Its input is the user and assistant text from `sessions.memory_seen_seq` on, cut near 24 KB at a message boundary; tool calls, tool results, reasoning and attachments never reach it. Its only tools are `memory_search`, `memory_read` and `memory_write`. A user has one pass at a time. When the run ends, `memory_seen_seq` moves past the last message it read.

Without Mycelium the agent is told it has no memory and no artifacts, and says once that connecting Mycelium in Settings, Apps gives it both.

## Chat search

`chats_search` (read-only, every chat and job run) and `GET /api/sessions/search?q=&space_id=` match the title and the user and assistant text of the user's web and matrix chats in one space, ignoring case. Tool output is never searched, and ideas, memory, cron and guest sessions never match. The query needs 2 characters; the route answers at most 20 results, the tool 10.

## Glouton

Glouton tools never send email. `glouton_save_draft` and `glouton_update_draft` store an outreach email with status `DRAFT`; the user reviews and sends it from Glouton's agent inbox at `{SUITE_GLOUTON_URL}/app/inbox`. Bulle has no send, convert, merge or delete tool for Glouton, and Glouton refuses those calls from a delegated token anyway.

The agent inbox is the confirmation gate. `glouton_save_draft`, `glouton_update_draft` and `glouton_suggest` (`suite.InboxTools()`) use the write token but raise no approval card: drafts and suggestions go straight to the inbox, and nothing is sent or applied without a click there. `glouton_create_company`, `glouton_update_company`, `glouton_create_person` and `glouton_update_person` change contacts directly, so they stay in `suite.WriteTools()` behind an approval card.

A user who signs up while Glouton is configured also gets the "Glouton daily review" starter job (weekdays 09:30 Paris): it drafts first contacts for new HOT and WARM prospects, drafts follow-ups for outreach sent more than 5 days ago without a reply, and files RESCORE and NEXT_ACTION suggestions, all straight to the inbox without approval cards. Existing users do not get it; they can create the same job by hand.

## Failure states

| State | Cause | What the user sees |
|---|---|---|
| `reconnect` | The Registre refresh token is missing or was rejected. | "Sign out of Bulle and sign in again." |
| `no_account` | The app answered 401 for a fresh token: the user never signed in there. | A link to open the app once. |
| `unavailable` | The app is down or timed out. | A retry later message. |

For Courrier, `no_account` and `reconnect` come from listing its MCP tools. `GET /api/suite/status` reports one state per configured app and drives the "Connected apps" section on the Account page. Signing out of Bulle and a back-channel logout both clear the user's cached tokens.
