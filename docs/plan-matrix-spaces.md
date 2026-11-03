# Plan: the Matrix bot as a teammate in each space's room

## Goal

The Matrix bot stops being a private assistant on a laptop. It runs on the Bulle server and works
like `@claude` in a Slack channel: it sits in each space's team room, answers only when someone
mentions it, and replies in a thread. The private assistant room goes away; the web app covers that.

Each Facile project is a Bulle space, and each space has a Matrix room on our homeserver. Linking
the two gives every project an agent the whole team can talk to.

## Decisions

- **The server runs the bot.** One bot account per Bulle user, configured in the web app
  (settings, Matrix): homeserver, bot user id, password, and who may call it. Nothing about Matrix
  stays in `config.yml`.
- **Mention only.** The bot answers a message only when it is mentioned: `m.mentions.user_ids`
  carries the bot, or the body or formatted body names its user id. Everything else in the room is
  ignored, including replies in a thread that do not mention it.
- **One thread, one session.** A mention outside a thread starts a thread rooted at that message;
  a mention inside a thread continues that thread's session. Answers always go into the thread.
- **One room per space.** Settings, Matrix lists each space with a room id. A room linked to a
  space runs in that space: its suite defaults, its Mycelium memory and its chat list. A mention in
  a room that is linked to no space gets no answer.
- **Who may call it.** The bot runs tools with its owner's access, so only Matrix users on the
  bot's allowlist may call it. A mention from anyone else is ignored. Writes still go through the
  owner's approval in the web app and the push notification, as for every run.
- **Crypto state on the server.** Each bot keeps its olm account and megolm sessions in its own
  Postgres schema, `matrix_u<user id>`, pickled with a key derived from `BULLE_CREDENTIALS_KEY`.
  The bot password is sealed with AES-GCM under the same key.

## Schema

Migration `00012_matrix.sql`, forward-only:

- `matrix_bots`: `user_id` (pk), `enabled`, `homeserver`, `bot_user`, `password` + `nonce`
  (sealed), `allow` (jsonb list of Matrix user ids), `updated_at`.
- `matrix_space_rooms`: `space_id` (pk, cascades with the space), `user_id`, `room_id`, unique on
  (`user_id`, `room_id`).
- `matrix_threads`: (`user_id`, `room_id`, `thread_id`) primary key, `session_id` cascading with
  the session.

## API

- `GET /api/matrix`: the bot (password never returned, only whether one is set), the allowlist,
  each space with its linked room, and the bot's live status (`off`, `connecting`, `running`, or
  the last error).
- `PUT /api/matrix`: saves the same shape; an empty password keeps the stored one. Saving restarts
  that user's bot.

## Code

- `core/chat`: `Message` gains `Thread` and `Mentioned`; replies go into the thread. The `Adapter`
  interface stays the seam a second chat platform plugs into.
- `apps/api/internal/matrixbot`: the store, the sealing, a manager that starts one bot per enabled
  account at boot and restarts it on save, and the responder that maps a thread to a session and
  waits for the run, as the scheduler does for jobs.
- `apps/api/modules/matrix`: the two routes.
- `apps/client`: a Matrix section in settings.
- The API builds with `-tags goolm`, as the CLI already does, so no libolm is needed.

## Removed from the CLI

`bulle chat` (daemon, systemd unit, in-room approvals, IPC socket), `bulle assistant` (digests,
watch, consolidate, which delivered through the chat daemon) and the `chat:` and `assistant:`
config blocks. A config that still has them is refused with a message pointing at the web app.
The server's jobs (morning brief, check-in, recap) replace the digests.

## Later

- Prisme, the suite's own chat app (`FacileStudio/Prisme`, plan in its `docs/plan.md`), runs on
  Matrix, so this bot works there unchanged. Its API provisions the agent account and room
  membership, so "Use Prisme" in the Matrix settings replaces the hand-made bot account.
- Shared spaces: when spaces gain several members, one bot per space instead of per user.
- Reactions as approvals inside the thread, so a teammate can approve without opening the web app.
