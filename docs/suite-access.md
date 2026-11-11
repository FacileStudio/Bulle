# Suite access

The agent reads the signed-in user's own Courrier mail, Agenda events, Opus tasks, Sablier timer, Nuage files, Mycelium memory and Glouton prospects. It does this in chat and in background jobs. The design and its reasons are in `plan-suite-delegation.md`. This page is the operating reference.

## How a call is authorised

1. The user signs in to Bulle through Registre. porte stores their access and refresh token in `porte_identities`.
2. A suite tool asks `suite.Tokens` for a token for one app. `Tokens` runs on porte/exchange: one refresh per user at a time across every app, and the identity under the issuer that still holds a refresh token, newest first, so a leftover row without one never reads as "reconnect".
3. On a cache miss, `Tokens` refreshes the user's Registre token if it is about to expire and writes the rotated pair back.
4. `Tokens` calls Registre's token endpoint with RFC 8693 token exchange: audience is the app's client id, scope is `<app>.read`.
5. Registre returns a JWT that lives five minutes. `Tokens` caches it in memory until 30 seconds before it expires.
6. The tool calls the app's API with that JWT as a bearer token. The app verifies it with porte and maps `sub` to its own user.

Bulle stores no app credential. Nothing is written to disk, to the environment of a process or to the VM.

## Apps reached through MCP

Courrier, Agenda, sonar, Nuage, Opus and Sablier have no built-in tools any more, and Plume, Capsule, Journal, Antenne, Vision and Echo never had any. Courrier: Bulle is an MCP client of Courrier's own server at `/mcp` (go-sdk, stateless 2026-07-28) and mounts its tools as the server names them: `courrier_list_accounts`, `courrier_search_mail`, `courrier_read_thread`, `courrier_draft`, `courrier_send`, `courrier_organize` and `courrier_delete_draft`. The client is `suite.MCP` on nacelle's `mcp/client`.

- **Token.** The same RFC 8693 exchange with the resource URI (`SUITE_COURRIER_MCP_RESOURCE`) as the audience, scope `courrier.read courrier.write` (read only if Registre refuses write). Registre resolves a URI audience to the app that declares it and answers a token whose `aud` is the resource, so it works at that server only. porte/exchange sends no `resource` parameter; the URI audience is equivalent at Registre. Bulle's existing `token_exchange` grant for Courrier covers it. A 401 drops the cached token and mints a new one once.
- **Connection.** One connection per server for the whole process, opened lazily by the first run that needs it, under that user's context: the server is stateless, so each request carries the token of the user whose call sends it (nacelle `Remote.Token`), and private tool lists are cached per user (`Remote.Identity`). A server that cannot be reached is left out of the run and retried after 30 s; the other apps keep working.
- **Confirmations.** Courrier asks before `courrier_send` and `courrier_delete_draft` through MRTR `input_required`. Bulle shows that question as an approval card (`args.kind = "mcp"`: server, tool, the server's message, the call's arguments) on the call's own id and returns the owner's answer to the server. A denial, an unanswered card or a run that ends declines, and Courrier then sends nothing. Bulle keeps no per-tool policy for MCP tools and no card can edit the mail. A call, its confirmation included, has 15 minutes, which is how long Courrier keeps the confirmation open.
- **Taint.** A read-only tool the server marks open world (`courrier_search_mail`, `courrier_read_thread`) is an untrusted source, like web pages: once one returns, the shell, memory deletes and `profile.md` writes ask first.
- **History.** Stored transcripts with the old built-in names (`courrier_list_inbox`, `courrier_search`, `courrier_get_email`, `courrier_save_draft`, `courrier_mark`, `courrier_update`, `courrier_move`) replay under the MCP name that does the same job (`suite.CanonicalTool`), read only.
- **Agenda.** The same client, token and connection rules for `SUITE_AGENDA_MCP_URL` (resource `https://agenda.facile.studio/mcp`, scope `agenda.read agenda.write`, covered by Bulle's `token_exchange` grant for Agenda). Tools: `agenda_list_calendars`, `agenda_list_events`, `agenda_read_event`, `agenda_find_free_slots`, `agenda_create_event`, `agenda_update_event` and `agenda_delete_event`. Agenda confirms only `agenda_delete_event`, through the same approval card; creates and updates run without a card, so the prompt tells the model to change events only when the user asked. `agenda_list_events` and `agenda_read_event` are open world, so they taint the run. Old transcripts replay `agenda_get_event` and `agenda_event` as `agenda_read_event`, `agenda_events` and `agenda_calendars` as the list tools, with `calendar`, `start_at`, `end_at` and `is_all_day` renamed to `calendar_id`, `start`, `end` and `all_day`. A space's `agenda_calendar` reaches the model as a prompt line, not as a default the client fills in.
- **sonar (the web).** The same client for `SUITE_SONAR_MCP_URL` (resource `https://sonar.facile.studio/mcp`, scope `sonar.read` only: Bulle never asks sonar for write, and its `token_exchange` grant for `sonar` allows read alone). Tools: `sonar_search` and `sonar_fetch`, always mounted beside the catalog. Both are open world, so they taint the run, and the prompt calls their results untrusted. Each search spends one of the user's daily sonar searches; an exhausted quota comes back as a failed call naming the upgrade page. A run may call `sonar_search` 20 times and `sonar_fetch` 30 times. Users' own engine keys live in sonar's settings, not in Bulle. Old transcripts replay `web_search` and `web_fetch` as the sonar tools.
- **Nuage (files).** The same client for `SUITE_NUAGE_MCP_URL` (resource `https://nuage.facile.studio/mcp`, scope `nuage.read nuage.write`, covered by Bulle's `token_exchange` grant for Nuage; Nuage lets the `bulle` actor reach the drive). Tools: `nuage_list_folder`, `nuage_search`, `nuage_read_file`, `nuage_list_shares`, `nuage_save_file`, `nuage_create_folder`, `nuage_move`, `nuage_revoke_share`, `nuage_delete` and `nuage_create_share`. They reach the personal drive and every shared space the user belongs to. Nuage confirms `nuage_delete` (to the trash) and `nuage_create_share` (a public link) through the same approval card; saves, folders, moves and revokes run without a card, so the prompt tells the model to change files only when the user asked. `nuage_list_folder`, `nuage_search` and `nuage_read_file` are open world, so they taint the run. Saving a chat attachment to Nuage is gone with the built-in `nuage_save` (the model never saw attachment ids). Old transcripts replay `nuage_list`/`files_list` as `nuage_list_folder`, `nuage_read`/`files_read` as `nuage_read_file` and `nuage_save`/`files_save` as `nuage_save_file`, with `folder` and `id` renamed to `folder_id` and `file_id`.
- **Opus (tasks).** The same client for `SUITE_OPUS_MCP_URL` (resource `https://opus.facile.studio/mcp`, scope `opus.read opus.write`, covered by Bulle's `token_exchange` grant for Opus). Tools: `opus_list_workspaces`, `opus_list_projects`, `opus_search_tasks`, `opus_read_task`, `opus_create_task`, `opus_update_task`, `opus_comment_task`, `opus_create_project`, `opus_update_project`, `opus_delete_task` and `opus_delete_project`. Opus confirms only the two deletes, through the same approval card; creates, updates, comments and archiving run without a card, so the prompt tells the model to change tasks only when the user asked. `opus_list_projects`, `opus_search_tasks` and `opus_read_task` are open world, so they taint the run. A space's `opus_workspace` reaches the model as a prompt line, not as a default the client fills in. Old transcripts replay `opus_list_tasks`/`opus_tasks` as `opus_search_tasks`, `opus_get_task`/`opus_task` as `opus_read_task` and `opus_get_project`/`opus_project`/`opus_projects` as `opus_list_projects`, with `id`, `start_date` and `due_date` renamed to `task_id`, `start` and `due`. Agent time booked to Sablier still lists the Opus tasks a run read, created or updated when their result carries a `facile_id` (Opus sends none yet).
- **Sablier (time).** The same client for `SUITE_SABLIER_MCP_URL` (resource `https://sablier.facile.studio/mcp`, scope `sablier.read sablier.write`, covered by Bulle's `token_exchange` grant for Sablier). Tools: `sablier_timer_status`, `sablier_list_entries`, `sablier_list_projects`, `sablier_weekly_summary`, `sablier_start_timer`, `sablier_stop_timer`, `sablier_add_entry`, `sablier_update_entry` and `sablier_delete_entry`. Sablier confirms only `sablier_delete_entry`, through the same approval card; the timer and entry writes run without a card, so the prompt tells the model to change time only when the user asked. `sablier_list_entries`, `sablier_list_projects` and `sablier_weekly_summary` are open world, so they taint the run. Old transcripts replay `sablier_status`, `sablier_projects`, `sablier_start` and `sablier_stop` under the MCP names, with `project` and `task` renamed to `project_id` and `task_id`. The REST client (`SUITE_SABLIER_URL`, `SUITE_SABLIER_AUDIENCE`, `suite/agent_time.go`) stays for one job no MCP tool covers: booking the agent's own runs and presence to the user (`POST /api/agents/sessions`, `POST /api/heartbeats`), see `BULLE_SABLIER_TIME_TRACKING`. It no longer has a status probe; Sablier's state in `/api/suite/status` comes from its MCP server.
- **Plume, Capsule, Journal, Antenne, Vision and Echo.** The same client for `SUITE_<APP>_MCP_URL` (resource `https://<app>.facile.studio/mcp`, scope `<app>.read <app>.write`, covered by Registre v0.12.1's `token_exchange` grants for Bulle). Tools are mounted as each server names them (`suite/mcp_tools.go` lists them). Each server confirms its own sensitive actions through the same approval card: Plume sends, reminders and deletes; Capsule revokes; Journal deletes of saved queries and alert rules; Antenne replays; Vision site and goal deletes and public dashboard links; Echo room deletes. Every other write runs without a card, so each app's prompt note tells the model to change data only when the user asked. Their read-only open-world tools taint the run. Capsule never sees a capsule's content.
- **Access.** Before it opens a connection, Bulle mints the person's token. When Registre refuses the exchange with `access_denied` (the app's access policy does not include the person, such as Journal and Antenne for non-admins), or the server answers 403, the app is left out of the person's runs without a log line and shows as `unavailable` in `/api/suite/status`. Bulle remembers a refusal for 5 minutes per person and app, and it never marks the server down for everyone else.
- **Never Casier.** Bulle has no Casier entry: an assistant never handles the owner's secrets, even though Registre's seed grants the exchange.
- **Dropped.** `assistant_seen_at` on listed mail and the `mail_seen` bookkeeping behind it. The table stays until a later migration drops it.

## Configuration

An app's tools exist only when both its URL and its audience are set. Setting one without the other stops the API at boot. An MCP app needs only its server URL.

| Variable | Value |
|---|---|
| `SUITE_COURRIER_MCP_URL` | Courrier's MCP endpoint, `https://courrier.facile.studio/mcp`. Empty switches mail off. |
| `SUITE_COURRIER_MCP_RESOURCE` | The resource Registre declares for that server. Defaults to `SUITE_COURRIER_MCP_URL`; set it only when Bulle dials an internal address. |
| `SUITE_AGENDA_MCP_URL`, `SUITE_AGENDA_MCP_RESOURCE` | The same pair for Agenda, `https://agenda.facile.studio/mcp`. Empty switches the calendar off. |
| `SUITE_SONAR_MCP_URL`, `SUITE_SONAR_MCP_RESOURCE` | The same pair for sonar, `https://sonar.facile.studio/mcp`. Empty leaves runs with no web tools. |
| `SUITE_NUAGE_MCP_URL`, `SUITE_NUAGE_MCP_RESOURCE` | The same pair for Nuage, `https://nuage.facile.studio/mcp`. Empty switches files off. |
| `SUITE_OPUS_MCP_URL`, `SUITE_OPUS_MCP_RESOURCE` | The same pair for Opus, `https://opus.facile.studio/mcp`. Empty switches tasks off. |
| `SUITE_SABLIER_MCP_URL`, `SUITE_SABLIER_MCP_RESOURCE` | The same pair for Sablier, `https://sablier.facile.studio/mcp`. Empty switches the time tools off. |
| `SUITE_PLUME_MCP_URL`, `SUITE_CAPSULE_MCP_URL`, `SUITE_JOURNAL_MCP_URL`, `SUITE_ANTENNE_MCP_URL`, `SUITE_VISION_MCP_URL`, `SUITE_ECHO_MCP_URL` (and each `_MCP_RESOURCE`) | The same pair for each app, `https://<app>.facile.studio/mcp`. Empty switches that app off. |
| `SUITE_SABLIER_URL`, `SUITE_MYCELIUM_URL`, `SUITE_GLOUTON_URL` | App origin, no trailing slash needed. Sablier's is only for booking the agent's own time. |
| `SUITE_SABLIER_AUDIENCE`, `SUITE_MYCELIUM_AUDIENCE`, `SUITE_GLOUTON_AUDIENCE` | The app's client id in Registre's `seed.yaml`. Mycelium is `mycelium`, Glouton is `1IdqK3bA2LWI-zuWVLqFpnIlCTVu117F` (scopes `glouton.read`, `glouton.write`). |

`OIDC_CLIENT_ID` and `OIDC_CLIENT_SECRET` authenticate Bulle to the token endpoint. The `bulle` client in Registre needs a `token_exchange` entry for each audience.

## Endpoints called

Each client keeps its paths in one constant block at the top of its file in `apps/api/internal/suite/`.

| App | Paths (GET) |
|---|---|
| Courrier | POST `/mcp` only (see above) |
| Agenda | POST `/mcp` only (see above) |
| sonar | POST `/mcp` only (see above) |
| Opus | POST `/mcp` only (see above) |
| Sablier | POST `/mcp` for the tools (see above); POST `/api/agents/sessions` and `/api/heartbeats` for the agent's own time |
| Nuage | POST `/mcp` only (see above) |
| Mycelium | `/api/spaces` (GET, POST), POST `/api/memory/search`, GET and PUT `/api/sync/files/memory/{path}?space_id=` |
| Glouton | `/api/agent/contacts?q=&stage=&temperature=&limit=`, `/api/agent/companies/{id}`, `/api/agent/people/{id}`, `/api/agent/outreach?companyId=&personId=`, `/api/agent/suggestions?status=`; writes: POST `/api/agent/companies`, `/api/agent/people`, `/api/agent/drafts`, `/api/agent/suggestions`, PATCH `/api/agent/companies/{id}`, `/api/agent/people/{id}`, `/api/agent/drafts/{id}` |

## Memory

`memory_search`, `memory_read` and `memory_write` work on the `memory/` directory of one Mycelium space. A Bulle space with a `mycelium_space_id` uses that space. Otherwise the tools use the user's personal space named "Bulle": the first call creates it in Mycelium (or adopts an existing one the user owns with that name) and `mycelium_spaces` remembers its id. The common tree is never used: an empty space id fails before any call. `memory_write` is not approval-gated (`suite.SelfWrites()`).

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
| `no_account` | The app answered 401 for a fresh token: the user never signed in there, or its MCP server refused to create the account (no verified email). | A link to open the app once. |
| `unavailable` | The app is down or timed out. | A retry later message. |

For Courrier, Agenda and sonar, `no_account` and `reconnect` come from listing their MCP tools. A server that still answers 401 after Bulle refreshed the token is `no_account`, and it never blocks other users: only a server that did not answer is left alone for 30 seconds. `GET /api/suite/status` reports one state per configured app and drives the "Connected apps" section on the Account page. Signing out of Bulle and a back-channel logout both clear the user's cached tokens.
