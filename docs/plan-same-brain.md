# Plan: same brain (Phase 7)

Written 2026-11-02. Replaces steps 35 to 38 of `docs/plan-web-ui.md`.

## Goal

One assistant per user, living in the API. Web, Matrix and scheduled jobs are channels of it. They share one memory, one ledger and one job list in Postgres. The `bulle chat` daemon and the crontab assistant jobs on ruche retire.

## Decisions

| Topic | Decision |
|---|---|
| Matrix home | The Matrix adapter moves into the API process as a channel. The `bulle chat` daemon goes. |
| Matrix tools | Matrix runs use the same runner as web: VM tools in the user's boite VM, host tools for the suite. The bot loses direct shell on ruche and `auto_approve`. |
| Journal | Dropped. `journal_dir`, `memory/journal.go` and their callers are deleted. |
| Connectors | Done in v0.94.0 (`apps/api/internal/suite`). Assistant sources reuse those clients. |
| Voice notes | Dropped from the Matrix channel for now. whisper.cpp is a host binary and is not in the image. |
| Antenne trigger | Dropped. The watch runs every 15 minutes from the scheduler, as the crontab does today. |
| Brain import | A subcommand of the API binary reads a tarball on stdin. No new HTTP route, no CLI auth flow. |

## Current state (observed 2026-11-02 on ruche)

- `~/.local/share/bulle/memory/MEMORY.md`: 73 bytes. No `USER.md`.
- `~/.local/share/bulle/chat/`: one room (`rooms/8debe33b1b4741cc.jsonl`), a sqlite crypto store (`matrix.db`), `pickle.key`, `recovery.key`.
- `~/.local/state/bulle/`: `ledger.json`, `watch.json`, `mail-senders.json`.
- `~/.config/bulle/jobs/`: `enhanced-tech-news-summary`, `lyon-deals`, `moulaga-report`.
- crontab: `assistant-watch` (every 15 min), three digests (09:00, 14:00, 20:00), `assistant-consolidate` (03:30), the three jobs above.
- `bulle-chat` user unit: active.
- The API uses none of `core/memory`, `core/assistant` or `core/chat`. Web runs have no memory today. The runner has no compaction.

## Design

```text
             +-------------------- bulle API container --------------------+
Matrix  <--> | matrix channel (leader-locked /sync) --+                     |
browser <--> | web handlers (SSE) ---------------------+--> runner --> VM   |
             | scheduler: user jobs + assistant jobs --+      |             |
             |                                               v             |
             |   brain: memory, ledger, watch state, senders (Postgres)    |
             |   suite clients (delegated tokens) for mail/agenda/opus     |
             +-------------------------------------------------------------+
```

### Seams in `core/`

- **`memory.Backend`**: `Read(name)`, `Write(name, data)`, `Backup(name, suffix, data)`, `Backups(name)`. `memory.Store` keeps its rules (caps, secret guard, read-back check, consolidate, forget, revert) and calls the backend instead of `os`. Today `core/memory` has five direct file calls, so the change is small. The file backend stays until the cleanup phase.
- **`assistant.StateStore`**: `Update(ctx, key, func([]byte) ([]byte, error))` and `Load(ctx, key)`. It replaces the file lock plus JSON blob in `ledger_store.go`, `watchstate.go` and the mail-senders file. The Postgres version does `SELECT ... FOR UPDATE` in one transaction.
- **`assistant.Sources`**: typed readers (`Mail`, `SentMail`, `Calendars`, `Events`, `Workspaces`, `Tasks`) replace the `Runner func(name, args...)` exec shape in `source_*.go`. The API implements them with `apps/api/internal/suite`. The CLI keeps an exec adapter until cleanup.

### Postgres (API)

- the next free migration (`00007_brain.sql` today):
  - `memory_files(user_id, name, body, updated_at)`, primary key `(user_id, name)`.
  - `memory_backups(id, user_id, name, suffix, body, created_at)`.
  - `assistant_state(user_id, key, data jsonb, updated_at)`, primary key `(user_id, key)`. Keys: `ledger`, `watch`, `mail_senders`.
  - `assistant_profiles`: add `settings jsonb` for `settings.Assistant` (timezone, digests, quiet hours, VIP, drop senders, sources).
  - `users`: add `matrix_user_id text unique`, `matrix_room_id text`.
  - `sessions`: add `external_key text`, unique `(user_id, channel, external_key)`. A Matrix room maps to one session.
- Matrix crypto: no migration. mautrix creates its own tables. `cryptohelper.NewCryptoHelper` accepts a `*dbutil.Database` (verified in the module cache), so the crypto store goes in the same Postgres. No volume needed.

### Matrix channel

- `apps/api/internal/matrix/`: wraps `core/chat.Matrix`. One replica holds `pg_try_advisory_lock` on a dedicated connection and runs `/sync`. The others retry every 30 s.
- Inbound message from a linked MXID: find or create the session (`channel=matrix`, `external_key=<room>`), call `runner.Start`, subscribe, and post the final text to the room. `INTERRUPT` posts an approval line; a reaction or reply goes to `runner.Decide`.
- Commands map onto API operations: `/stop` -> `runner.Stop`, `/new` -> new session for the room, `/forget` -> `memory.Store.Forget`, `/snooze` and `/quiet` -> ledger and profile settings, `/status` -> a text summary.
- Linking: the web settings page shows a one-time code. The user sends `/link <code>` to the bot. That room becomes the user's assistant room. Unlinked MXIDs get no answer. This replaces the `allow` and `rooms` config lists.
- Env (casier, Dokploy): `MATRIX_HOMESERVER`, `MATRIX_USER_ID`, `MATRIX_PASSWORD`, `MATRIX_RECOVERY_KEY`, `MATRIX_PICKLE_KEY`. Unset means the channel is off.
- New device: the API logs in as a fresh device and self-signs with the recovery key (`core/chat/matrixcrypto.go` already does this). The old sqlite Olm sessions are not migrated; the bot only needs new messages.
- Build: the Dockerfile adds `-tags goolm` (pure-Go Olm, `CGO_ENABLED=0`).

### Runner changes

- Memory for every channel: inject `memory.Store.Snapshot` into the system prompt and add `memory.RememberTool`, per user.
- Compaction: use `core/compaction` (as `agent.ChatCompactor` does in the CLI) when the conversation passes `compact_at`. Matrix sessions are long-lived, so they need it. Web sessions benefit too.
- Delivery: jobs gain `delivery: matrix`, which posts to the user's linked room through the Matrix channel.

### Assistant jobs

- The scheduler gets built-in per-user assistant jobs: `watch` (every 15 min), the three digests, and `consolidate` (03:30). They run in user time zone from `assistant_profiles.settings`. They are code jobs, not agent prompts, so they call `core/assistant` directly with API deps: suite sources, `StateStore`, the API's LLM backend as `Completer`, and Matrix delivery.
- They run only for users with a linked Matrix room and the assistant switched on.
- The same advisory lock pattern as user jobs (`pg_try_advisory_xact_lock`) keeps single firing across replicas.

## Steps

Each step names its files and its exit check. Phases A to C ship with no user-visible change and can be released one by one. Phase D is the cutover.

### Phase A: seams in `core/` (no behavior change)

1. `core/memory`: add `Backend`, a `fileBackend`, route the five `os` calls through it. Delete `journal.go`, `EnsureJournal`, `JournalPath`, `settings.Assistant.JournalDir` and their callers in `apps/cli`. [exit: `go test ./core/memory ./apps/cli/...` pass unchanged; after a rebuild and `bulle-chat` restart on ruche, "remember X" then `/forget X` works in Matrix]
2. `core/assistant`: add `StateStore` with a file version that keeps today's lock files. Move `ledger_store.go`, `watchstate.go` and the senders file onto it. [exit: ledger and watch tests pass unchanged]
3. `core/assistant`: add the `Sources` interfaces. Wrap today's exec calls in `apps/cli` as `execSources`. [exit: `source_*_test.go` pass against fakes; the next 09:00 digest on ruche arrives as before]

### Phase B: brain in the API (web gets memory)

4. `apps/api/migrations/00007_brain.sql` and `apps/api/internal/brain/`: the Postgres `memory.Backend` and `StateStore`, scoped by user id. [exit: store tests with `TEST_DATABASE_URL`, including two concurrent `Update` calls on one key]
5. Runner: memory snapshot in the prompt, `remember` tool, compaction. [exit: a web chat that says "remember X" answers X in a new session]
6. `apps/api/internal/suite`: implement `assistant.Sources`. [exit: a test against the existing suite fakes in `world_test.go`]

### Phase C: Matrix channel and assistant jobs (off by default)

7. `apps/api/internal/matrix/`: leader lock, linking, inbound to runner, outbound text, approvals, commands. Postgres crypto store. `-tags goolm` in the Dockerfile. [exit: in staging env with a test bot account, a linked user's message gets a reply and an approval round trip works]
8. Settings: `PUT /api/assistant/settings` and `POST /api/matrix/link` (code), with the porte auth floor. Client: a small settings section for link status, digest times and quiet hours. [exit: `scripts/check-web.sh` green]
9. Scheduler: built-in assistant jobs and `delivery: matrix`. [exit: a forced digest for a test user lands in their room once, with two API replicas running]
10. Import: `api import-brain --user-email <email>` reads a tarball on stdin with `memory/`, `ledger.json`, `watch.json`, `mail-senders.json`, `jobs/*.yml` and the assistant section of `config.yml`. It upserts, so a second run changes nothing. It skips room history: the room starts fresh and memory carries the facts. [exit: two runs against a test database give the same rows]

### Phase D: cutover on ruche (one sitting, in order)

11. Back up: `crontab -l > ~/.local/state/bulle/crontab.backup-<date>`, copy `~/.local/share/bulle` and `~/.local/state/bulle`.
12. Stop the old writers: `systemctl --user disable --now bulle-chat`, remove the `BULLE JOB` blocks for the assistant and the three jobs from the crontab.
13. Import: `tar` the files from step 10 and pipe them into `docker exec -i <api container> /api import-brain --user-email <you>`.
14. Set the `MATRIX_*` env in Dokploy (you do this; secrets are yours to place), deploy.
15. Link: get the code in the web app and send `/link <code>` to the bot in your room.
16. Verify. [exit: the memory fact answers in both Matrix and web; `/status` answers; the next digest arrives once; `crontab -l` has no Bulle assistant entries; the three jobs show in `/jobs`]

Rollback: re-enable `bulle-chat`, restore the crontab backup, unset `MATRIX_*`. The files are untouched, so nothing is lost. Never run both at once: two `/sync` loops on one account answer twice, and two ledgers send digests twice.

### Phase E: cleanup (breaking for the CLI)

17. Delete `bulle chat`, `bulle assistant` and `chat_*.go`, `assistant_*.go` in `apps/cli/cmd`. Delete the file `memory.Backend`, the file `StateStore`, `execSources`, the IPC socket (`core/chat/ipc.go`) and CLI cron `delivery: matrix`. Local `bulle cron` without Matrix delivery stays. Remove the assistant and chat keys from `docs/configuration.md`. Release note with `BREAKING CHANGE:`. [exit: `scripts/check.sh` and `filet check` green; `go build ./...` in `apps/cli` has no `core/chat` import]
18. `ROADMAP.md`: close Tracks K and L as moved into the web app. Wiki: update `projects/bulle.md`.

## Risks

- **Two writers.** The worst failure is the daemon and the API both live: double replies and double digests. Phase D stops the old writers before the import and before setting `MATRIX_*`.
- **Matrix device trust.** A new device needs cross-signing. If the recovery key fails, the bot cannot decrypt until you verify it in Element. Test with a throwaway account in step 7 first.
- **Lost abilities.** The bot no longer has shell on ruche, voice notes or the journal. Anything that used them now goes through the VM or stops.
- **Model cost.** CLI jobs ran on per-job profiles (`free`, `lerouteur`). API jobs use the server model (`BULLE_LLM_MODEL`). The import maps profiles to nothing; check job cost after a day.
- **Replica count.** The leader lock makes several replicas safe. One replica is still the expected setup.

## Skip (YAGNI)

- Memory page in the web app beyond what `/forget` and the settings section give.
- Room history import.
- Several Matrix rooms per user, and other chat platforms.
- Migrating Olm sessions from the sqlite store.
