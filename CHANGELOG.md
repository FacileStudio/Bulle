# Changelog

## [3.25.0] - 2026-11-07

### Added

- feat: the agent asks multiple-choice questions on a card (ask_user), each with a field to type your own answer; one click answers a single question
- feat: the agent marks mail read or unread and stars it with courrier_mark, without an approval card
- feat: mail the assistant already saw in an earlier run carries assistant_seen_at, so it stops reporting the same mail as new; your read flag is left alone

### Changed

- feat(runner): browser_navigate opens any URL without asking; browser_confirm still holds every non-GET request
- feat: courrier_update now only archives and deletes, still behind an approval card
- feat: memory dates facts that go stale

### Fixed

- fix(api): a failed memory pass no longer marks the chat learned, so the next pass retries it

## [3.24.0] - 2026-11-07

### Added

- feat: Artifacts, a new sidebar section. The agent writes markdown artifacts into the space's Mycelium with artifact_create and artifact_update; each one shows in the chat as a card that opens a side panel (a full-screen sheet on mobile) and stays in the Artifacts section, rendered with Mycelium's rich markdown (alerts, charts, metrics, compare, diff, mermaid)
- feat: Bulle remembers naturally. Mycelium is its only memory: profile.md is read at the start of every chat, the agent saves durable facts quietly as you talk, and a background pass learns from a chat after it has been idle for 10 minutes. Without Mycelium, the agent says once that connecting it gives memory and artifacts
- feat: search past conversations from the sidebar, and the agent can search them too with chats_search
- feat: the agent answers with rich components (tables, stats, charts, buttons, forms) whenever an answer is scanned or acted on, not only read once

### Changed

- feat: Jobs moved from the sidebar into Settings, Jobs; /jobs redirects there, and the profile menu shows when a job is running
- feat(runner): web_fetch reads any page without asking; only browser_navigate to a URL the run has not seen still asks
- style(client): no borders in OpenUI tables, dividers and buttons, nor in the new pages

### Removed

- refactor(core): the unused core/memory package

## [3.23.0] - 2026-11-06

### Added

- feat: Ideas, a new sidebar section after the chat. Each morning, at the space's local hour, Bulle reads the space's mail, calendar, tasks and memory with read-only tools and leaves 3 to 7 cards, each with a "why now", its sources, a ready prompt and an optional draft. "Do it" starts a normal chat behind the approval card
- feat: steer ideas with a "What should Bulle think about?" box, a Goals list per space, thumbs up and down, and dismiss with a reason; dismissed ideas stay away for 14 days
- feat: "Think of more" runs up to 3 extra batches a day; an optional morning push says when the batch is ready

### Fixed

- fix(api): a daily ideas batch that could not start, such as when the user already runs two runs, is retried within the same hour instead of being skipped for the day
- fix(api): an approval decision is never published before its approval card, so a run-wide grant no longer leaves a stale card behind
- fix(client): idea pages ignore late answers from a previous request or space, keep toggles in sync when a save fails, and show the server's reason in error toasts

## [3.22.2] - 2026-11-06

### Fixed

- fix(api): approving a suite write once no longer auto-approves later writes on the same target; only "stop asking" remembers it, and a mail move is never remembered

## [3.22.1] - 2026-11-06

### Fixed

- fix(api): suite approval cards show the call's own fields again (the new title, status, dates), with the resolved target and space on top
- fix(api): `spaces_list` marks only the conversation's own space as current, even when two spaces share the same pointers
- fix(client): the batch card offers "stop asking" for the same targets only, and never on a mail delete

## [3.22.0] - 2026-11-06

### Added

- feat(api): the agent can list the user's spaces with `spaces_list` and act in any of them with the user's own rights; the space prompt no longer claims the spaces are isolated

### Fixed

- fix(api): "stop asking" now covers one tool on one target (a task, an event, a mailbox and action) instead of the whole tool for the run
- fix(api): a mail delete is never remembered, whatever the account
- fix(api): suite write approval cards name the resource, the change and the space

## [3.21.0] - 2026-11-06

### Added

- feat(client): landing page shows the live plans (muse `PricingPlans`, hidden while the catalog is empty) and uses muse `DownloadCta` for the downloads link

### Changed

- build(client): `@facile/muse` v0.8.1 to v0.11.0

## [3.20.3] - 2026-11-06

### Fixed

- fix(client): remove the invalid tabindex on the batch approval list, which svelte-check flagged; the scrollable list stays keyboard reachable in current browsers

## [3.20.2] - 2026-11-06

### Fixed

- fix(api): approving a call with "stop asking" now also approves the calls of the same tool already waiting, so no stale cards remain
- fix(api): an abandoned approval no longer leaves its entry behind for the rest of the run
- fix(client): the mail card's sticky action row matches the chat panel background, and the draft resets when the agent re-proposes the email

## [3.20.1] - 2026-11-06

### Fixed

- fix(api): the run-wide "stop asking" grant no longer covers deletions, file saves, browser or web fetch tools (including any tool gated per URL); each of those still asks every time
- fix(client): batch rows show every field of each call instead of a truncated line, and the "stop asking" link only appears for tools the server accepts it for
- fix(client): a decision that fails after the server already settled the call no longer flips it back to pending; "Approve all" only sends calls still pending
- fix(client): settled approvals keep a one-line summary of what ran; the mail form has a visible keyboard focus, wider text-button targets and a fixed message height

## [3.20.0] - 2026-11-06

### Added

- feat(client): several pending approvals of the same tool collapse into one batch card with "Approve all N" and "Deny all"
- feat(api): an approval can carry `remember: true` to cover that tool for the rest of the run, so bulk deletes and creates ask once; mail sends never take it
- feat(api): `opus_project` takes an optional `workspace`, so projects outside the default workspace resolve

### Changed

- style(client): the approval card and the email review card lose their borders and shield header; the mail form uses plain underlined fields, a larger message area and a single primary Send action


## [3.19.1] - 2026-11-06

### Fixed

- fix(api): `opus_update_project` no longer lists `name` as required in its tool schema, so the agent can change only an icon or slug; an update with no field is refused before any request

## [3.19.0] - 2026-11-05

### Added

- feat(cli): every commit the agent makes through run_command now carries `Co-authored-by: Bulle Agent <bulle@facile.studio>`, appended as its own paragraph; quote-aware across chained commands and idempotent, with `tools.git_coauthor` (default on) and `BULLE_GIT_COAUTHOR` to opt out

## [3.18.0] - 2026-11-05

### Added

- feat(api): `opus_project`, `opus_create_project`, `opus_update_project` and `opus_delete_project`, so the agent can CRUD Opus projects; the three writes go through the approval gate like every other suite write
- feat(client): the device pairing page surfaces live pairing and relay status, a real empty state and a loading state, with Solar bold-duotone glyphs on the pair button, device tiles and Wi-Fi block

### Changed

- feat(api): the server's default model is now `xiaomi/mimo-v2.6-flash` (an explicit `BULLE_LLM_MODEL` still wins)

## [3.17.1] - 2026-11-05

### Changed

- style(client): the profile dropdown loses its ring and divider; the shadow alone lifts it off the sidebar

## [3.17.0] - 2026-11-05

### Added

- feat(client): connected apps draw the app's own Solar bold-duotone glyph, the one the vitrine uses, and provider rows show Anthropic, OpenAI, Google and OpenRouter marks

### Changed

- feat(client): the twelve settings tabs fold into five, the assistant moves under Profile next to sign-out, and old section links still land on the right tab

## [3.16.0] - 2026-11-05

### Added

- feat(api): `GET /models`, the current provider's model catalogue with display names, shared by the web picker and the CLI's `/model`
- feat(client): the Model setting searches the whole catalogue instead of the operator's short list

### Changed

- feat(cli): `session.model` names the model a session starts on; the server's catalogue still decides what is valid, and `-model` wins

### Removed

- feat(api): the CLI config saved in the web app (`GET /api/cli/config`, its settings tab and migration column). The CLI config stays on the machine

## [3.15.0] - 2026-11-05

### Added

- feat(api): pair bohs-01 devices and open the hardware gateway socket; devices are capped per user, live sockets per device, and every hello frame is re-checked against the socket's credentials
- feat(client): pair, relay and connect an already paired bohs-01 device over Web Bluetooth
- feat(api): a per-user default model, set in the web settings under Model; `GET /api/llm` returns it, so a new CLI session starts on it
- feat(api): `GET/PUT /api/preferences` and `GET /api/cli/config`, so the CLI settings saved in the browser reach every signed-in machine
- feat(client): Model and CLI sections in the settings page
- feat(cli): the server's CLI config is one layer, above the built-in defaults and below `~/.config/bulle/config.yml`

### Changed

- feat(cli): the iteration cap is lifted by default, and `-max-iterations` reports 0 as the flag default
- feat(runner): a web run has no iteration cap either; the 40-call default only ever cut long jobs short

### Fixed

- fix(api): long streaming answers, reasoning included, no longer die with `stream error: stream ID N; INTERNAL_ERROR; received from peer`. The server's 15 s write deadline became `RST_STREAM(INTERNAL_ERROR)` on every HTTP/2 response that outlived it, so it is gone; handlers that need a deadline set their own
- fix(client): a transient BLE write error no longer drops the rest of a bohs.v1 chunk
- fix(api): hardware ids must be 8 lowercase hex characters, and frames over 1024 bytes are refused

## [3.14.0] - 2026-11-04

### Added

- feat(client): interactive blobs and ambient illumination on the landing page, with borderless presentation

## [3.13.0] - 2026-11-04

### Added

- feat(api): `courrier_search` and `courrier_inbox` with no account cover every mailbox, each email tagged with `account_id` and `account_email`
- feat(api): `courrier_send` approvals show the email as it would leave; the decide endpoint takes edited `to`, `cc`, `bcc`, `subject`, `body` and `as: draft`
- feat(client): send approvals render as an editable email with Send, Save as draft and Cancel
- feat(client): tool calls collapse to one summary line that expands to the grouped list

### Fixed

- fix(api): a reply whose original sits in another mailbox is sent from that mailbox, in its thread, instead of failing
- fix(api): mail tool descriptions and prompt steer the model to search every mailbox, treat the user's own addresses as mailboxes, and send through the approval card instead of stopping at a draft

## [3.12.0] - 2026-11-04

### Added

- feat(api): accept Registre token-exchange bearers (`OIDC_MACHINE_AUDIENCE`, `OIDC_SCOPE_PREFIX`, both `bulle` in compose) so Glouton can use the signed-in user's Bulle persona and model to write mail

## [3.11.0] - 2026-11-03

### Added

- feat(runner): `web_search` tool backed by sonar, with instance keys, a per-user daily limit (`BULLE_SEARCH_DAILY_LIMIT`, default 100) and per-user search keys that override them
- feat(runner): `web_fetch` and `browser_navigate` ask the owner before opening a URL the run has not seen; per-run caps on searches and fetches
- feat(browser): `bulle-browser` Playwright daemon in the VM with snapshot, find, click, type, screenshot and back tools, behind `BULLE_BROWSER=true` (off by default; needs a boite image with Playwright)
- feat(runner): non-GET page requests are held until the owner approves `browser_confirm`; the card shows the daemon's record with secret-looking fields redacted, and one approval covers that origin for the run
- feat(browser): page WebSockets are refused and the daemon exits after 15 idle minutes
- feat(credentials): per-user keys for parallel, brave, tavily and exa

## [3.10.0] - 2026-11-03

### Changed

- feat(chat): agents answer a direct message with a plain message instead of opening a thread, and keep one conversation per direct message; replies inside a thread and in shared rooms still go to the thread

## [3.9.2] - 2026-11-03

### Fixed

- fix(client): Android shows Bulle pushes while the app is open; Capacitor dropped them unless `presentationOptions` includes `alert`
- fix(client): the app creates a high-importance `bulle` notification channel, so pushes pop up instead of landing in a silent "Miscellaneous" channel
- fix(client): signing out forgets this phone or browser on the server and drops its push registration, so the next user does not get the last user's notifications
- fix(client): toasts sit at the bottom of the screen again; the `pb-28` offset was for a bottom nav Bulle does not have
- fix(api): FCM messages carry the `bulle` channel, a per-session tag that replaces older pushes for the same chat, and a 24 h TTL
- fix(api): a malformed FCM token is forgotten, and delivery errors keep FCM's reply so a sender mismatch is readable in the logs

## [3.9.1] - 2026-11-03

### Fixed

- test(settings): config tests blank every exported `BULLE_` and `NACELLE_` variable, so a value in the developer's shell such as `BULLE_ADDITIONAL_PROMPT` no longer fails them

## [3.9.0] - 2026-11-03

### Added

- feat(api): only an agent's owner can use its tools from Prisme. Anyone else who mentions or messages the agent gets a guest session with no tools, so a colleague cannot make it read your mail, calendar, files or memory, or act for you
- feat(api): direct messages to an agent get an answer without a mention; the owner's DMs run in their default space with every tool
- feat(api): replies posted where people other than the owner can read them pass a jev disclosure screen first and are withheld when they reveal private data, when jev is unsure, or when `TYPESAFE_API_KEY` is unset
- feat(api): every run's prompt carries the current date and time, in the server's `TZ` (Europe/Paris in compose), so the agent no longer says it cannot read the clock

## [3.8.0] - 2026-11-03

### Added

- feat(api): Glouton suite tools behind `SUITE_GLOUTON_URL` and `SUITE_GLOUTON_AUDIENCE`: read prospects, companies, people, outreach and suggestions; save and edit outreach drafts, file suggestions, and create or edit companies and people. Drafts and suggestions go straight to Glouton's agent inbox with no approval card, and nothing is sent without a click there; company and person edits need approval
- feat(api): Glouton shows in the connected apps status once configured
- feat(api): new accounts get a "Glouton daily review" starter job when Glouton is configured

## [3.7.5] - 2026-11-03

### Changed

- The prompt shows the signed-in agent's name instead of Bulle, and its loader takes the agent's blob colour
- The line under the prompt starts at the left edge and uses one muted colour
- The banner drops the agent name, the model and the context file count, and shows a root under home as ~/...

## [3.7.4] - 2026-11-03

### Changed

- Spaces use the share-circle icon

## [3.7.3] - 2026-11-03

### Removed

- The sidebar's new space button; spaces are created from Settings > Spaces

## [3.7.2] - 2026-11-03

### Fixed

- fix(api): give each Matrix bot device its own crypto database; mautrix checks tables by name across schemas, so only the first bot could start, and a new device needs a fresh store

## [3.7.1] - 2026-11-03

### Fixed

- fix(api): close each Matrix bot's crypto connection pool when the bot stops and cap it at 3 connections; leaked pools exhausted Postgres and kept every Prisme agent offline

## [3.7.0] - 2026-11-03

### Added

- feat(api): send the assistant's https avatar to its Prisme agent; Prisme draws the Facile blob otherwise

### Fixed

- fix(api): deleting a background job stops its running runs, including a scheduled run that is just starting
- fix(client): draw the assistant avatar with @facile/blob instead of blobatar.dev
- fix(api): stop sending blobatar avatars to Sablier; an assistant without an https avatar sends none

## [3.6.0] - 2026-11-03

### Added

- Spaces follow Registre: every shared space you belong to appears in Bulle, synced at boot, every 5 minutes and on `POST /api/registre/sync`. Leaving one moves its chats and jobs to your default space.
- A shared space's Opus workspace and Mycelium space are found by its name, so nothing has to be typed.

### Changed

- Space settings only list your spaces: shared ones are read-only, your own keep rename and delete. The id fields are gone.

### Fixed

- The web client polled `GET /api/jobs` in a loop, about 18 times a second per open tab.
- The Postgres pool is capped at 20 connections; it had grown until Postgres refused every client.

## [3.5.0] - 2026-11-03

### Added

- feat(client): desktop self-update system with reactive state store, progress tracking, and settings section

## [3.4.0] - 2026-11-03

### Added

- feat(cli): bulle login with no address signs in to https://bulle.facile.studio instead of refusing

### Changed

- feat(client): no decorative borders: muse v0.8.0 and the app's own login split, page header and footer, usage list, code blocks, chips and download cards separate by tone and spacing

## [3.3.0] - 2026-11-03

### Changed

- feat(cli): the input bar follows grok-cli: a flat muted panel with a Bulle label that turns into a three-box loader during a run, and one line under it with the model, context left and spend on the left and the keys that work now on the right
- feat(cli): no green ready line above the input any more; only approvals, stops and errors appear there

## [3.2.0] - 2026-11-03

### Added

- feat(api): each finished run is booked in Sablier as the user's assistant, an agent user with its name and blob avatar (POST /api/agents/sessions, Sablier 1.4.0). Time blocked on an approval is left out, and cron runs are marked as not cowork
- feat(api): a message in a web or Matrix session sends a presence heartbeat for the human to Sablier, at most one per session every 2 minutes
- BULLE_SABLIER_TIME_TRACKING (default true) turns both off; they only run when SUITE_SABLIER_URL is set

## [3.1.2] - 2026-11-03

### Fixed

- fix(api): chat agent provisioning read Prisme's numeric space ids as strings and failed on every pass

## [3.1.1] - 2026-11-03

### Added

- feat(client): drop files anywhere in a conversation or on the new-chat screen to attach them, with a full-area drop overlay; a file dropped elsewhere no longer opens in the browser and leaves Bulle
- feat(client): in the mobile app the + button offers Take a photo, Photo library or File, so the camera and gallery are one tap away

## [3.1.0] - 2026-11-03

### Added

- feat(api): the server provisions each user's chat agent in Prisme automatically: an account named after the assistant, a member of every channel of the user's spaces, each channel working in the Bulle space of the same name; needs PRISME_URL and PRISME_SERVICE_KEY
- feat(client): settings, Chat agent shows the agent's account, rooms and status

### Removed

- feat(api): the hand-made Matrix bot account, password, allowlist and PUT /api/matrix; there is nothing to configure

## [3.0.1] - 2026-11-03

### Fixed

- fix(api): each user's attachments are capped at 1 GB in total; an upload past it gets 413 instead of filling the database

## [3.0.0] - 2026-11-03

### Added

- feat(api): each user's Matrix bot runs on the server; it joins each space's linked room, answers only when mentioned, replies in a thread and keeps one Bulle session per thread in that space
- feat(api): GET and PUT /api/matrix for the bot account, the Matrix users allowed to call it and the room linked to each space; the password is sealed with BULLE_CREDENTIALS_KEY
- feat(client): a Matrix tab in settings with the bot status, account, allowlist and a room field per space

### Removed

- refactor(cli)!: bulle chat, bulle assistant and bulle cron, the chat IPC socket, local voice transcription and the chat: and assistant: config blocks; the Matrix bot and jobs now run on the server

## [2.4.0] - 2026-11-03

### Added

- feat(api): GET /api/push/stream is now the user's live event stream: session.created, session.updated, session.deleted, run.started, run.finished, approval.pending and approval.decided, next to notification
- feat(api): every session in GET /api/sessions carries status: idle, running or waiting (a tool call waits for approval)
- feat(client): conversations show a working or needs-approval dot in the list and the header, and new, renamed, deleted and running conversations sync live across web, mobile and desktop
- feat(client): a browser with no push service (Helium, Brave without Google push) shows notifications from an open Bulle tab
- feat(client): collapse the sidebar on web and desktop with a button or Ctrl+B (Cmd+B); the choice is remembered per device

## [2.3.0] - 2026-11-03

### Added

- feat(api): the user's Registre picture is stored on every sign-in and backfilled at boot for users who signed in before, and GET /api/auth/me returns it as user.avatar_url
- feat(client): the profile menu and the account card show the user's Registre picture, falling back to the initial when there is none

### Changed

- build(client): @facile/muse v0.7.3

## [2.2.0] - 2026-11-03

### Added

- feat(cli): ${NAME} placeholders in config.yml values are filled from the environment, so a tracked config can keep personal values such as the Matrix identity out of the repo

## [2.1.0] - 2026-11-03

### Added

- feat(client): attach images, PDFs and text files to a message with the + button, drag and drop or paste; sent images show as thumbnails
- feat(api): attachments are stored per user and sent to the model as image, document or text parts (nacelle v0.30.0)
- feat(api): the agent keeps long-term memory in Mycelium (memory_search, memory_read, memory_write), one Mycelium space per Bulle space
- feat(api): the agent lists, searches, reads and saves the user's Nuage files (files_list, files_search, files_read, files_save behind approval)
- feat(api): spaces group chats and jobs, with a default Courrier account, Agenda calendar and Opus workspace the agent uses when none is named
- feat(client): a space switcher in the sidebar and a Spaces tab in settings

### Changed

- feat(api): the existing sessions and jobs move into a Personal space on upgrade

### Fixed

- fix(client): a browser with no push service gets a plain explanation instead of "Registration failed - push service error"

## [2.0.0] - 2026-11-03

### Added

- feat(api): GET /api/llm returns the instance backend, default model, the models a client may use (BULLE_LLM_MODELS) and the user's assistant persona
- feat(cli): the CLI takes its backend, model and persona from the server on every run; /model picks from the server's models
- feat(cli): a paste of 10+ lines or 1000+ characters shows as [Pasted text #N +X lines] and expands on submit; long questions echo as three lines plus a count
- feat(api): a 30-day usage summary of tokens and cost, shown in a usage settings tab

### Changed

- feat(cli): the assistant's name follows the version in the banner instead of sitting beside the blob
- fix(cli): the welcome blob is centred on its shape (FacileStudio/blob v0.1.1), so round blobs stay round

### Removed

- feat(cli)!: provider:, profile:, profiles/, chat.profile, assistant.profile, a job's provider: and profile:, -backend, -profile and the BULLE_BACKEND, BULLE_MODEL, BULLE_PROVIDER_* and BULLE_PROFILE variables; without a server login bulle refuses to start
- feat(cli)!: the web_fetch tool, -fetch and BULLE_FETCH
- feat(cli): the startup context, system prompt and filet notes, and the bash, fetch and MCP tool counts in the banner

### Fixed

- fix(client): the mobile app no longer shows the desktop app entry

## [1.8.0] - 2026-11-03

### Added

- feat(cli): a signed-in session opens on the assistant's blob avatar and the name set in the web app, drawn as truecolor pixel art; the text logo stays when start_message is set, NO_COLOR is set, or the server does not answer

### Changed

- feat(client): avatars render locally with @facile/blob (FacileStudio/blob v0.1.0) instead of blobatar and images from blobatar.dev; every avatar keeps its shape and colours

### Fixed

- fix(cli): bulle login over SSH, with --device, or with no browser prints an address that works from any device and reads the pasted code, instead of redirecting the remote browser to 127.0.0.1

## [1.7.0] - 2026-11-03

### Added

- feat(desktop): the window has no title bar; the traffic lights sit in the sidebar header (native on macOS, drawn on Windows and Linux), and the headers drag the window
- feat(desktop): desktop notifications for approvals and finished jobs; a toast with Open shows instead when the window has focus
- feat(api): GET /api/push/stream sends each notification to the user's open desktop apps as server-sent events
- feat(client): the sidebar shows your avatar and name; it opens a menu with Settings, Notifications, the apps page and Sign out
- feat(client): /install lists every download inside the app, so a signed-in user no longer has to sign out to reach it
- feat(api): the agent can create, change and delete the user's suite data: send, reply to, draft, flag, move and delete mail in Courrier; create, update and delete Agenda events; create, update, move and delete Opus tasks; start and stop Sablier timers
- feat(api): new read tools: courrier_read, agenda_event, opus_task, opus_projects, sablier_projects

### Changed

- style(client): the logo, favicon, desktop icon and Android launcher icon are the smile circle
- style(client): the downloads page features your platform, shows every platform with its logo and the current version, and groups the terminal install and checksums
- feat(api): POST /api/push/test always succeeds, since the desktop stream is always on
- feat(api): every suite write now waits for an approval card in the chat; the approval gate existed before but nothing armed it
- feat(client): the approval card names the action in plain words ("Send an email"), lists the fields, and marks deletes in red
- feat(client): suite tool calls are grouped under "Facile apps" in the tool list

### Fixed

- fix(registre): after a desktop or mobile sign-in, the browser tab says you are signed in and can close it, instead of spinning on "Connecting to application" (shipped in Registre 324f7ef)

## [1.6.0] - 2026-11-02

### Added

- feat(desktop): a Tauri desktop app for macOS (universal), Windows (x64) and Linux (AppImage, .deb), built from the web client in apps/desktop
- feat(desktop): sign-in opens the system browser and returns through the studio.facile.bulle://auth link, like the Android app
- feat(desktop): the app checks GitHub for a new release at launch and offers it in a toast; `latest.json` and the update signatures ship with each release
- feat(client): the desktop app skips the landing page and opens on a two-pane sign-in screen
- feat(client): the downloads page links every desktop installer, puts the visitor's OS first, and links the checksums
- ci: the Desktop workflow builds the three installers on each tag and attaches them to the release
- feat(api): native Android notifications through Firebase Cloud Messaging and browser Web Push (VAPID), alongside ntfy; devices register at POST /api/push/devices
- feat(client): the app asks for notification permission from Settings > Notifications or a prompt on the Jobs page, and a tap opens the job's conversation

### Changed

- feat(api): CORS accepts the Tauri origins `tauri://localhost` and `http://tauri.localhost`

## [1.5.0] - 2026-11-02

### Added

- feat(api): sessions carry their `channel`, and each assistant message carries `answer`, the text after its last tool call
- feat(client): a job's conversation hides the job prompt, the tool calls and the thinking, and shows only the final answer; "Working on it" shows while it runs, and "Show steps" reveals the rest
- feat(client): a job's conversation appears in the sidebar as soon as the run starts

## [1.4.0] - 2026-11-02

### Added

- feat(api): GET /api/jobs returns `running` and `last_session_id` for each job, read from the runs table so a restart never leaves a job stuck as running
- feat(client): a running job shows "Running" on its card, and the Jobs entry in the sidebar shows an orange dot; the list refreshes every 5s while a run is going and every 30s otherwise
- feat(client): each job card links to the conversation its last run wrote to ("View output", or "Watch" while it runs)

### Changed

- style(client): the run history shows the status and start time only, without run IDs

## [1.3.2] - 2026-11-02

### Changed

- style(client): job cards show the name, a readable schedule ("Weekdays at 08:00"), the switch and a two-line prompt; no borders, badges or labelled buttons
- style(client): the command groups in a conversation and the chat header have no borders or row dividers

## [1.3.1] - 2026-11-02

### Fixed

- fix(client): the spinner on a running command in a conversation is a round 16px ring again; it was a 24px ring squeezed into a 16px slot

## [1.3.0] - 2026-11-02

### Changed

- feat(client): settings are split into tabs (Profile, Assistant, Notifications, Providers, Computer, Apps), each at its own /account/<section> URL
- style(client): the settings cards, rows and personality cards have no borders or dividers

## [1.2.0] - 2026-11-02

### Added

- feat(api): a new account starts with three jobs it can edit or delete: a morning brief (08:00), an afternoon check-in (14:00) and an evening recap (19:00), weekdays, Europe/Paris, covering mail, calendar and tasks

### Changed

- feat(client): one sidebar on desktop and mobile holds the pages, New chat, the chat list and the account card with a settings cog; the separators and the mobile drawer border are gone
- feat(client): the new chat screen has no top bar; on mobile a floating, blurred round button opens the sidebar

## [1.1.0] - 2026-11-02

### Changed

- feat(client): on mobile, the bottom nav is gone so the composer sits at the bottom; the top-left button or a right swipe opens a sidebar with chats, jobs, sessions and the account

## [1.0.1] - 2026-11-02

### Fixed

- fix(client): the composer has its left padding and focus border back; a missing space had merged the two classes

### Changed

- style(client): the mobile intro and login screens have a plain background, without the gradient

## [1.0.0] - 2026-11-02

### Changed

- feat(client): the mobile intro and login screens show the Pixel blobatar; the agent is still named Bulle

## [0.100.1] - 2026-11-02

### Fixed

- fix(client): Android sign-in now finishes after the SSO redirect; the login state only uses characters porte keeps, and the API allows the app origin `https://localhost` by default
- fix(client): a failed sign-in now shows its error on the login page instead of nothing
- fix(client): the face in the Android icon is smaller, so it no longer crowds the edges

## [0.100.0] - 2026-11-02

### Changed

- feat(client): the composer has fully rounded ends, on the new chat screen and in conversations

### Removed

- feat(client): the computer status chip, start/error banners and sidebar row are gone from the chat; the status lives in settings only

## [0.99.0] - 2026-11-02

### Added

- feat(client): the Android app icon is the Bulle face on black, with a monochrome layer for themed icons

### Changed

- feat(client): Providers and Computer moved from the sidebar into the settings page
- feat(client): the computer can no longer be started from the UI; the agent boots it when a task needs it, and stopping it stays possible

### Fixed

- fix(api): Android sign-in now returns to the app; the API ran porte v0.5.2, which ignored `flow=mobile`, and now runs v0.8.0 with `OIDC_MOBILE_REDIRECT` (default `studio.facile.bulle://auth`)

## [0.98.1] - 2026-11-02

### Fixed

- fix(client): the traits on the account page are left-aligned like the other fields; the onboarding keeps them centered
- fix(client): the agent avatar next to the assistant name in the sidebar and mobile nav draws about 25% larger

## [0.98.0] - 2026-11-02

### Added

- feat(client): the Android app opens on a three-slide intro instead of the web landing page, then a mobile sign-in screen; the intro shows once per install

### Changed

- feat(client): the Android status and navigation bars take the page color in light and dark mode, with readable icons, including when the theme switches while the app is open
- feat(client): the Android launch screen uses the page color instead of the stock Capacitor splash

## [0.97.0] - 2026-11-02

### Added

- feat(client): a public `/downloads` page with the signed Android APK and its checksum, the terminal install commands, and the desktop app marked as coming soon
- feat(client): the landing page links Downloads in the header and hero, and a "Take Bulle with you" section offers the APK and announces the desktop app

### Changed

- refactor(client): the landing and downloads pages share one header and footer, and one install commands block

## [0.96.1] - 2026-11-02

### Fixed

- fix(client): the Providers page crashed on open (`props_invalid_value`) because each key field was bound to an undefined value; the web app and the Android app are both affected

## [0.96.0] - 2026-11-02

### Added

- feat(cli): `bulle login [url]`, `bulle logout` and `bulle whoami` sign the CLI in to a Bulle server; `server.url`, `BULLE_SERVER_URL` and `BULLE_TOKEN` configure it
- feat(cli): with no local key source, model calls go through the server at `/api/llm/<backend>`; `api_key`, `api_key_command`, `base_url` and vendor env keys still win
- feat(provider): the Anthropic backend accepts a custom `base_url`
- feat(api): users store provider keys per backend, encrypted with AES-GCM under `BULLE_CREDENTIALS_KEY`; without a user key the instance env key is used
- feat(api): `/api/llm/{backend}/*` proxies and streams model calls for signed-in users
- feat(api): push notifications through ntfy; background jobs with `delivery: push` notify when they finish
- feat(client): Providers page at `/settings/providers`
- feat(client): Android app (Capacitor); each release attaches a signed `bulle-android.apk`

### Removed

- BREAKING: `assistant.journal_dir` and the nightly journal template are gone. Delete the key from `config.yml`, because unknown keys fail the config load

### Changed

- refactor(memory): memory files go through a storage backend, ready for a Postgres store
- refactor(assistant): the ledger, watch, quiet and sender state, and the mail, calendar and task reads, go through interfaces instead of fixed files and CLI calls

- feat(client): replace the onboarding modal with a fullscreen six-step `/welcome` flow (welcome, name, personality, style, apps, meet), with a "Connect your apps" step that shows each suite app's connection state with live sample replies, a labelled progress bar, resumable drafts and reduced-motion support
- refactor(client): the account page reuses the onboarding name field, personality cards and trait chips
- fix(client): signed-in users without an assistant name are redirected to `/welcome` instead of seeing a modal over the chat

## [0.95.0] - 2026-11-02

### Changed

- perf(runner): the computer boots on the first VM tool call, not at the start of a run. Host-only runs and cron jobs (inbox, agenda, tasks, memory, web fetch) no longer start the VM
- feat(api): `BULLE_IDLE_MINUTES` defaults to 120 instead of 30. A negative value such as -1 disables the idle stop; 0 is rejected
- feat(client): opening the app or sending a message no longer starts the computer; the server boots it when a tool needs it
- feat(client): the banner says "Waking your computer" for a stopped computer and keeps the first-run text for a new one

### Fixed

- fix(computers): a non-positive idle window can no longer make the sweeper stop every running computer

## [0.94.1] - 2026-11-02

### Changed

- feat(client): redesign the login page as a split layout like the other Facile apps, with a brand panel and a "Back to home" link

## [0.94.0] - 2026-11-02

### Security

- fix(runner): stop reading the server owner's mail, calendar and tasks for every user. The suite tools no longer run the Courrier, Agenda, Opus and Sablier CLIs with server-wide tokens

### Changed

- feat(suite): the suite tools call Courrier, Agenda, Opus and Sablier over HTTP as the signed-in user, with a short-lived read-only token that Registre issues by RFC 8693 token exchange (`apps/api/internal/suite`). Nothing is stored; tokens live five minutes in memory
- feat(suite): `opus_tasks` lists the user's own tasks by default and a whole project when `project` is given; `agenda_events` reads every calendar the user can see and takes RFC 3339 bounds computed in the user's timezone
- feat(suite): an app's tools are registered, and named in the system prompt, only when its `SUITE_<APP>_URL` and `SUITE_<APP>_AUDIENCE` are both set
- feat(suite): signing out and back-channel logout clear the user's cached tokens
- feat(api): `GET /api/suite/status` reports connected, no_account, reconnect or unavailable per app
- feat(client): "Connected apps" section on the Account page
- build(docker): runtime image back to Alpine, with no downloaded CLI binaries
- build(deploy): replace `COURRIER_*`, `AGENDA_*`, `OPUS_*`, `SABLIER_*` env vars with `SUITE_<APP>_URL` and `SUITE_<APP>_AUDIENCE`

### Removed

- fix(runner): the `CommandExecutor` host-command path and the Sablier config file written into the shared home directory

### Documentation

- docs: `docs/plan-suite-delegation.md` (design) and `docs/suite-access.md` (operations)

## [0.93.5] - 2026-11-02

### Fixed

- fix(runner): capture combined stdout and stderr from system commands in defaultCommandExecutor
- fix(runner): generate Sablier config from environment tokens automatically before status check
- fix(deploy): populate Facile suite tokens in Dokploy environment for production agent use

### Fixed

- fix(docker): package Facile suite CLI tools (Courrier, Agenda, Opus, Sablier) into production container with Ubuntu runtime
- fix(runner): auto-detect default Courrier account and default Agenda calendar when not explicitly provided
- fix(runner): add `courrier_accounts` and `agenda_calendars` host tools for assistant account and calendar discovery
- fix(runner): support project and limit filtering on `opus_tasks`
- fix(runner): backward-compatible fallback for `sablier_status` when `--json` flag is not supported
- fix(deploy): forward Facile suite tokens and server URLs in docker-compose environment

## [0.93.3] - 2026-11-02

### Added

- feat(runner): agent background job management host tools (`create_background_job`, `list_background_jobs`, `update_background_job`, `delete_background_job`) with recursion guards
- feat(runner): native Facile suite host tools for Courrier (`courrier_inbox`, `courrier_search`), Agenda (`agenda_events`), Opus (`opus_tasks`), and Sablier (`sablier_status`)

## [0.93.2] - 2026-11-02

### Changed

- feat(client): rename Sessions nav button to agent name and display agent avatar icon
- feat(client): amplify horizontal caret gaze tracking and increase eye travel
- fix(client): remove duplicate loading label in top-right computer chip

## [0.93.1] - 2026-11-02

### Added

- feat(client): pointer-driven eye tracking and input caret following for the empty chat avatar

## [0.93.0] - 2026-11-02

### Added

- feat(jobs): scheduled background jobs with cron expressions, Postgres advisory lock deduplication, and runner execution
- feat(jobs): dedicated `/jobs` management surface with active toggle, run now trigger, history viewer, and creation modal
- feat(openui): extended OpenUI component library with Text and Divider components, updated Card description support, and prompt generator
- feat(client): centered 128px interactive Blobatar avatar with hover animation on empty session chat start

## [0.92.7] - 2026-11-02

### Added

- feat(client): dedicated `/computer` page in sidebar and navigation displaying virtual machine state, hardware specs, and power controls
- feat(client): auto-start agent virtual machine on chat load and message dispatch
- feat(client): smooth 3-dots animated typing indicator replacing spinner, and active streaming indicator

### Fixed

- fix(runner): replace conflicting default system prompt identity so named assistants never refer to themselves as Bulle
- fix(client): remove border, padding, and duplicate text from agent avatars across chat replies, header, account, and onboarding modals
- fix(client): live booting feedback and status polling during computer creation and startup

## [0.92.6] - 2026-11-02

### Added

- feat(assistant): brand new onboarding experience on `/sessions` with name customization, persona archetypes (Pragmatic Senior, Terse Minimalist, Supportive Mentor, Witty Hacker), and trait chips
- feat(assistant): dynamic Blobatar profile pictures (`https://blobatar.dev/avatar/<name>?gen=2`) with live preview, randomizer, and display in chat header and message replies
- feat(assistant): personality master prompt compiler injecting custom personas directly into the agent system prompt
- feat(assistant): assistant profile editor in user Account page

### Added

- feat(client): real Facile suite landing page with responsive hero, Muse tokens, 3-card features grid, and interactive terminal installation card

### Fixed

- fix: move `monorepo_root` to file root in `mise.toml` to eliminate mise unknown field warning
- fix(sessions): restore command outputs, tool calls, execution statuses, and thinking across conversation reloads

### Fixed

- fix(computers): `BULLE_BOX_KEY` is stored base64-encoded on one line, which is what `docs/deploy.md` tells you to set, and the API read everything that was not PEM as a file path: `open LS0tLS1CRUdJTi…: file name too long` failed the boot and the container restart-looped. The key is now read as PEM, as its base64, or as a path to either, and that is also why the box refused it — a base64 blob was handed to `ssh -i`
- fix(computers): a key that cannot be read no longer stops the API from starting. It warns, generates a pair for that boot, and publishes its public half to each box, so a wrong secret costs a warning instead of an outage

## [0.92.3] - 2026-11-02

### Fixed

- fix(computers): a box now trusts the key the API logs in with. boite authorizes only the pair it generates for the instance, so every run failed at the preflight probe with `Permission denied (publickey,password)` on `boite@10.128.0.1`. The API sends its public key with the create request and pushes it to boxes that predate the change, so an existing computer heals on its next run instead of needing a rebuild (`internal/computers/key.go`)
- fix(computers): `docker-compose.yml` forwards `BULLE_BOX_KEY`, which is the name the API reads; `BOITE_CONTROL_KEY`, the name it forwarded, was never read, so the container started with no key at all. An empty key generates its own pair at boot instead of failing every run
- fix(sessions): a run that answers nothing — the box failed to open, the user stopped it — left two user messages side by side, and the engine merged them into one on the next run, so a session of failed turns came back as a single bubble of glued text. The outcome is recorded as the assistant's turn, and a transcript already merged is split back into one message per text part when it is read

## [0.92.2] - 2026-11-02

### Fixed

- fix(api): ship `ssh` in the container image (alpine with openssh-client) so the preflight isolation probe no longer fails with `exec: "ssh": executable file not found`
- fix(api): `GET /sessions/{id}` returns the latest run when it ended in error, so the client shows the failure after a reload instead of a silent first message

## [0.92.1] - 2026-11-02

### Fixed

- fix(api): align boite worker payload sizing and parse memory and disk units
- fix(runner): default model to OpenRouter with claude-haiku-5.5 and fall back when session backend is unconfigured
- fix(compose): forward OpenRouter environment variables to the Bulle container

## [0.92.0] - 2026-11-01

### Added

- feat(web): Bulle web UI serving the suite web app at `bulle.facile.studio` with Registre SSO authentication
- feat(computers): per-user persistent boite microVM computer lifecycle management and offline disk snapshots
- feat(api): live runner with AG-UI events, approval gates for outward tools, and session store
- feat(client): SvelteKit 2 + Svelte 5 chat interface with streaming tool calls, approvals, and muse components
- feat(openui): render generative OpenUI blocks inside assistant replies with muse tokens and fallback

### Changed

- refactor: the repository is a Go workspace. The CLI moved to `apps/cli` (module `github.com/FacileStudio/Bulle/apps/cli`, with `main.go`, `cmd/` and the TUI packages in `apps/cli/internal/`), and the packages with no terminal UI moved to `core/` (module `github.com/FacileStudio/Bulle/core`: `assistant`, `chat`, `compaction`, `cost`, `engine`, `jev`, `memory`, `provider`, `sandbox`, `sessions`, `settings`, `skills`, `usage`, `xdgtest`). `go.work` joins them. The `bulle` binary behaves as before
- build: GoReleaser builds from `apps/cli`, `scripts/check.sh` runs each stage once per workspace module, and CI caches both `go.sum` files. `filet.yml` allows a depth of 4 for `apps/cli/internal/<pkg>`
- docs: the GitHub repository is now `FacileStudio/Bulle`. `go install github.com/FacileStudio/bulle@latest` is gone; use `github.com/FacileStudio/Bulle/apps/cli@vX.Y.Z`, which needs `apps/cli/vX.Y.Z` and `core/vX.Y.Z` tags

## [0.91.1] - 2026-10-30

### Fixed

- fix(migration): a relative symlink in `~/.bulle` (stow-style dotfiles) is recreated as an absolute link instead of dangling, a failed move is reported once per process, and an empty `ide/` or a dead `chat.sock` no longer keeps `~/.bulle` alive (`internal/settings/migrate*.go`)
- fix(status): `bulle status` fails when a crontab entry redirects to a log directory that does not exist, which is what silently stops every entry installed before 0.87 (`~/.bulle/logs` moved to `~/.local/state/bulle/logs`; run `bulle assistant install` and `bulle cron install <name>` again). It no longer flags the watch during quiet hours, matches whole crontab marker lines, and reports the error when `courrier` is missing (`cmd/status*.go`)
- fix(memory): the secret filter no longer refuses "the token issue", "le mdp est dans Bitwarden", file paths or commit hashes, and catches AWS, GitHub and Slack tokens; the instruction filter catches "whenever ... forward" and "never tell" orders; consolidation also skips facts already in `USER.md` (`internal/memory/secret.go`)
- fix(chat): every auto-approved call is written to the daemon log and the daemon warns at start-up while `chat.auto_approve` is on; the memory framing says "the user" instead of a hard-coded name (`internal/agent/chatapprove_gate.go`)
- fix(assistant): the mail source logs when it stops at 1,000 unread conversations (`internal/assistant/source_mail.go`)
- docs: the runtime directory falls back to `/run/user/<uid>/bulle` before `$TMPDIR`, the IDE discovery file moved to the runtime directory, and `auto_approve` is described accurately

## [0.91.0] - 2026-10-30

### Added

- feat(assistant): `assistant.journal_dir` makes the nightly `consolidate` read yesterday's journal and create today's file from the vault template (`internal/memory/journal.go`)

### Fixed

- fix(memory): consolidation never received the journal (`Options{}` was empty), and facts it proposed skipped the secret and instruction checks; both are fixed, and notes and journal are capped at 6,000 characters in the prompt (`internal/memory/consolidate.go`)

## [0.90.0] - 2026-10-30

### Added

- feat(memory): `remember` refuses secrets and instruction-like facts, and never writes the same fact twice (`internal/memory/guard.go`)
- feat(chat): the memory block is framed as notes about the user, data and not instructions, so a mail cannot plant standing orders through it (`cmd/chat_prompt.go`)

### Changed

- docs(memory): the `remember` tool description says what to store and what never to store (`internal/memory/tool.go`)

## [0.89.2] - 2026-10-30

### Fixed

- fix(agent): the system prompt names the session's model and backend, so the model stops guessing its identity from unrelated environment variables (`internal/agent/environment_model.go`)

## [0.89.1] - 2026-10-30

### Fixed

- fix(settings): `chat.profile` was dropped when the config layers merged, so the chat daemon always ran on the global profile instead of the one named under `chat:` (`internal/settings/chat.go`)

## [0.89.0] - 2026-10-30

### Added

- feat(assistant): digests now carry the mail you sent in the last three days as context (`sent_recently`), so the model knows what you are waiting on; it never notifies on them (`internal/assistant/source_sent.go`)

## [0.88.2] - 2026-10-30

### Fixed

- fix(assistant): split the triage page loop so `filet check` passes again (`internal/assistant/source_mail.go`)

## [0.88.1] - 2026-10-30

### Fixed

- fix(assistant): the mail source reads every `courrier triage` page (100 per page, at most 10) instead of only the first 50 conversations (`internal/assistant/source_mail.go`)

## [0.88.0] - 2026-10-30

### Added

- feat(status): `bulle status` checks the config, the chat daemon and its socket, the assistant crontab entries, the last watch, every cron job (trusted, installed) and the mail triage in one call, and exits 1 on any failure (`cmd/status*.go`)

## [0.87.3] - 2026-10-30

### Added

- feat(chat): `chat.auto_approve` runs every tool call in a room without asking; off by default (`internal/agent/chatapprove_gate.go`)

## [0.87.2] - 2026-10-30

### Fixed

- fix(settings): a process without `XDG_RUNTIME_DIR`, cron above all, falls back to `/run/user/<uid>/bulle`, so cron jobs reach the chat daemon's socket instead of looking under `/tmp` (`internal/settings/xdg.go`)

## [0.87.1] - 2026-10-30

### Fixed

- fix(tests): tests that pointed only `HOME` at a temp dir now isolate every `XDG_*` variable, so CI no longer reads the runner's own directories (`internal/xdgtest`)
- fix(lint): clear every info-level filet finding so the `fail-on: info` gate passes (`internal/assistant`, `cmd`, `internal/settings`)
- fix(assistant): ledger timestamps use `omitzero`, so unset times are left out of the JSON instead of written as year 1

## [0.87.0] - 2026-10-30

### Changed

- refactor(settings): config, data, state and runtime files follow the XDG base directories; `~/.bulle/` and `~/.bulle.yml` migrate on first run and are never overwritten (`internal/settings/xdg.go`, `internal/settings/migrate.go`)
- refactor(chat): the chat socket and the IDE discovery files live in the runtime directory, not `~/.bulle/`

### Fixed

- fix(tests): eleven tests read or wrote the developer's real home; they now run under `xdgtest.Isolate`

## [0.86.0] - 2026-10-30

### Added

- feat(assistant): `bulle assistant watch|digest|consolidate|install|status`, quiet hours, three daily digests and a dedupe ledger (`internal/assistant`, `cmd/assistant*.go`)
- feat(assistant): classified mail through `courrier triage`, VIP override and a local fallback for older servers (`internal/assistant/source_mail*.go`)
- feat(chat): rooms persist across restarts, nudges are recorded into the room and the conversation rolls over at 04:00 (`cmd/chat_room*.go`)
- feat(chat): quiet notices, reactions as replies, `/stop /new /snooze /quiet /forget /status` and voice notes (`internal/chat`, `cmd/chat_commands.go`)
- feat(chat): in-room approvals gate every non-read-only tool call (`internal/agent/chatapprove*.go`, `cmd/chat_approve*.go`)
- feat(memory): bounded `USER.md` and `MEMORY.md`, a verified `remember` tool, `/forget` and nightly consolidation with backup (`internal/memory`)

### Fixed

- fix(cron): matrix delivery sends only text inside `<deliver></deliver>` and drops untagged output (`internal/agent/cronmatrix.go`)
- feat(chat): read receipts on accepted messages (`internal/chat/presence.go`)
- fix(chat): refuse tool inputs whose keys differ only by case, so the approval gate and the shell tool read the same command (`internal/agent/chatapprove_gate.go`)
- fix(chat): allow only listed flags on read-only `courrier`, `agenda`, `opus` and `mycelium` commands, and ask before `web_fetch` (`internal/agent/chatapprove*.go`)
- fix(chat): approval cards show decoded fields, wait for delivery before listening, and refuse a typed verdict when several cards are pending (`cmd/chat_approve*.go`)
- fix(chat): cap voice downloads and ffmpeg output, and keep temp paths out of chat errors (`internal/chat/voice*.go`)
- fix(assistant): lock the ledger across processes, run digests with no tools, and lock watch and digest runs (`internal/assistant`, `internal/agent/assistanttoolless.go`)
- fix(assistant): digests honour snooze and noise, sender rules match exact addresses, and cron gets `CRON_TZ` (`internal/assistant`, `cmd/assistant_install.go`)
- fix(memory): `/forget` needs four characters and reports its backups, consolidation drops at most three lines, and `Revert` saves a safety copy (`internal/memory`)

## [0.85.4] - 2026-10-27

### Fixed

- fix(chat): normalize room key and deliver triggered replies to Matrix (`cmd/chat_ipc.go`)
- fix(chat): check `BULLE_CHAT_SOCK` in `SocketPath` and ensure clean connection lifecycle in IPC server (`internal/chat/ipc.go`)
- fix(cron): format timezone layout via standard library to support negative UTC offsets (`internal/agent/cronapply.go`)
- fix(cron): fallback to local log file when Matrix delivery is unavailable (`internal/agent/cronlog.go`)
- fix(cron): expand silent output matching for no-alert variants (`internal/agent/cronmatrix.go`)
- fix(agent): strip only balanced outer quotes in `parseExportLine` (`internal/agent/mcp_env.go`)
- fix(chat): provide fallback error message when daemon response error is empty (`cmd/chat_send.go`, `cmd/chat_trigger.go`)

## [0.85.3] - 2026-10-27

### Added

- feat(cron): inject exact local time and timezone header into cron job prompts (`internal/agent/cronapply.go`)

## [0.85.2] - 2026-10-27

### Fixed

- fix(agent): load all store secrets into environment via `tiroir export` in `ensureUserEnv` (`internal/agent/mcp_env.go`)
- fix(cron): strip leading empty lines and stay silent on `OUTPUT NOTHING` in Matrix delivery (`internal/agent/cronmatrix.go`)
- fix(cron): extend Matrix delivery IPC socket timeout to 60s (`internal/agent/cronmatrix.go`)

## [0.85.1] - 2026-10-27

### Added

- feat(chat): support profile inheritance with `chat.profile` in configuration (`internal/settings/chat.go`, `internal/agent/chatturn.go`)

### Fixed

- fix(build): drop local `nacelle` replace directive from `go.mod` to allow standalone CI builds (`go.mod`)

## [0.85.0] - 2026-10-27

### Added

- feat(chat): local IPC socket server for the chat daemon (`~/.bulle/chat.sock`), allowing external tools and cron jobs to dispatch messages and trigger conversational turns (`internal/chat/ipc.go`, `cmd/chat_ipc.go`)
- feat(chat): CLI commands `bulle chat send` and `bulle chat trigger` to interact directly with the running chat daemon (`cmd/chat_send.go`, `cmd/chat_trigger.go`)
- feat(cron): native Matrix delivery (`delivery: matrix` and `delivery: matrix:<room>`) via chat daemon IPC socket (`internal/agent/cronlog.go`)
- feat(cron): support `profile: <name>` inheritance in job specs to cleanly inherit provider configuration without duplicating YAML blocks (`internal/settings/cron.go`, `internal/agent/cronapply.go`)

### Fixed

- test(tui): isolate `TestResolveEditorWithVariousEnvironmentVariables` from ambient `GIT_EDITOR` environment variable (`internal/tui/prompt_test.go`)

## [0.82.0] - 2026-10-07

### Added

- feat(compaction): pick the judge's decision model with `limits.compaction.judge.provider` — `jev` (TypeSafe System One, the default) or `clef` (Cloudflare's decision model through OpenRouter's Decisions API). The provider derives the endpoint, host, model and vendor key (`TYPESAFE_API_KEY` / `OPENROUTER_API_KEY`); an explicit `model` or `base_url` still wins as an escape hatch for a proxy or a pinned build. Clef is calibrated against the labeled corpus before shipping: zero false prunes at the shipped threshold, 65% label agreement (`internal/compaction`, `internal/jev`, `internal/agent`, `internal/settings`, `docs/configuration.md`)
- feat(compaction): the judge classifies against where the session stands — session goal plus the latest user directive — instead of the anchor alone, capped at 4000 runes; `GoalText` keeps its pinned-task contract for ledger consolidation (`internal/compaction/goal.go`)
- feat(status): the judge's answer line names the serving provider when the Decisions API reports one and prints its billed cost (`internal/compaction`, `internal/tui`)

### Changed

- refactor(settings): first boot writes a minimal scaffold carrying only real user decisions, and compaction's ratios, tail bounds, judge internals and the grind knobs are sunstopped — still read when present, no longer documented or scaffolded, never removed from the schema, so old files keep parsing unchanged (`internal/settings`, `example.bulle.yml`, `docs/configuration.md`, `docs/development.md`, `README.md`)

### Fixed

- fix(chat): gopls did not carry the repo's `-tags=goolm`, so the local filet gate could not type-check `internal/chat` and reported build-constraint errors on `mautrix`; `.gopls.json` now declares the build flag the repo already builds with, and `chat.go` uses `WaitGroup.Go` (`.gopls.json`, `internal/chat`)

## [0.81.0] - 2026-10-02

### Changed

- refactor: rename the project from `kori` to `bulle` — binary, module path, config directory (`~/.bulle.yml`), environment variables (`BULLE_*`), install script, Homebrew cask, GoReleaser project name and every remaining `kori` reference in docs and tests (`main.go`, `cmd`, `internal`, `docs`, `README.md`, `ROADMAP.md`, `filet.yml`, `.goreleaser.yml`, `install.sh`, `.gitignore`)

### Fixed

- fix(settings): the start banner's old three-line glyph block left `Defaults` one line over filet's 35-line function limit, and the new five-line `bulle` wordmark pushed it further. The default's own `maxConcurrency` is now grouped with the other `Defaults` locals on the same line, so the banner is the only change in the function and the body lands at 35 lines exactly (`internal/settings`, `filet.yml`)

## [0.79.0] - 2026-09-25

### Added

- feat(chat): send typing indicators while generating turn responses (`internal/chat`, `filet.yml`)

## [0.78.0] - 2026-09-25

### Added

- feat(chat): inbound chat surface over Matrix with end-to-end encryption (`bulle chat`). Runs a supervised daemon holding a Matrix connection, turning allowlisted inbound messages into headless agent turns and returning threaded answers into the room. Supports device verification via `bulle chat verify --recovery-key`, cross-signing key generation, room invite auto-joining, and Megolm group session management for encrypted rooms (`internal/chat`, `cmd`, `internal/settings`, `docs/configuration.md`)
- docs(ide): document the surface for users, and be honest about diff (`README.md`, `docs/`)

## [0.77.0] - 2026-09-24

### Added

- feat(ide): a session can publish itself to an editor over a unix socket — `--ide`, or `$BULLE_IDE` for every session in a shell. It writes `~/.bulle/ide/<pid>.json` (0600 in a 0700 directory, removed on exit) and listens on `$XDG_RUNTIME_DIR/bulle/<pid>.sock`, newline-delimited JSON, one object per line, every object carrying a protocol version and a type. The frozen contract is `docs/ide-protocol.md` and its first consumer is `bulle.nvim`. The session reports the run's shape (`hello`, `turn`, `done`), what it is doing (`tool`, `edit`, `error`), and the one thing that needs an answer (`approval`); the editor can answer an approval, run a prompt with its cursor's place attached, ask the session to scroll to a file, and stop a run. The `edit` event carries the changed line span in the new file plus the added and removed counts, so an editor marks a change without re-deriving it or reading the file. Forward compatibility is mandatory on both sides — a receiver ignores an unknown type and unknown fields inside a known one — so a protocol version bump is never a breaking change. A session that publishes to nobody pays nothing for it: with neither the flag nor the variable there is no socket, no file, no goroutine and no hook, and the nil surface costs a nil check (`internal/ide`, `internal/tui`, `internal/agent`, `cmd`, `docs/ide-protocol.md`)
- feat(tui): the stats line under the state line now opens with `provider · model` — `openrouter · deepseek/deepseek-v4.1-flash` — ahead of the price, in the banner's own words, so which model is being billed is readable mid-run rather than only at launch. It follows a `/model` switch, and it names the *resolved* model: a config that leaves `provider.model` empty lets the backend supply its own default, and the banner and the session config now read one shared helper for it, so the two cannot disagree. The session total, the token pair and the context figure now come from `total()`/`tokenTotals()` for the footer, `/cost`, `/status` and the recap alike, instead of four hand-written sums (`internal/agent`, `internal/tui`). Incidental: `/status` no longer prints an empty `anthropic/` for a session that never named a model

### Fixed

- fix(ide): an approval is refused only when an editor was actually asked. The socket's first cut read "nobody attached" as a refusal, so `--ide` — or a `BULLE_IDE=1` left in a shell profile — denied every tool call the moment it was asked, with the terminal prompt still on screen asking a question that had already been answered. The answer is now three-valued: with no editor attached nobody was asked, and the terminal decides, which is where that decision sat before the socket existed. An editor that *was* asked and then went quiet, lost its connection, or ran out the clock is still a refusal, never an implicit yes (`internal/ide`, `internal/tui`)
- fix(ide): a session takes its socket and its discovery file down with it. Both were left on disk when a `--ide` session exited, so the next reader found a file naming a socket nobody was listening on and had to fall back to the pid-liveness check the protocol keeps for a killed process (`internal/agent`)
- fix(tui): the context counter was sized from the *run* total instead of the conversation. `KindDone` carries every turn's usage summed — each turn re-bills the whole conversation, so that figure is a bill, not a size — and sizing from it showed a three-turn run as three times the context it held (201k against a conversation of 101k), pinned in `TestTheContextSizeIsTheLastTurnsInputNotTheRunsTotal`. Compaction reads that figure and selects its tier from it, so the inflation also compacted early on every multi-turn run. Only a turn sizes now (`internal/tui`)
- fix(tui): the live price estimate carried the cost-per-token rate it had learned from the last turn across `/clear` and across `/model`. After a clear it priced a session that had spent nothing, putting a phantom dollar figure on the next stream; after a switch it priced the new model's tokens at the old model's rate until a turn reported its own cost. Both drop it now, and every model switch goes through one `activate` that writes the active-model fields and the usage sink in one place, instead of three call sites repeating the same five lines (`internal/tui`)
- fix(cost): raise nacelle to v0.28.2. Its OpenAI-schema backends (`openai`, `openrouter`, `google`) reported a prompt token twice whenever any of it was cached or written: `prompt_tokens` is the whole prompt and its `prompt_tokens_details` parts are parts *of* it, not neighbours of it — a 34375-token prompt with 32768 written read as 67145 against a `total_tokens` of 34377. Since the context is sized as `InputTokens + CacheReadTokens + CacheCreationTokens`, the figure was up to double on those backends, and compaction ran early. v0.28.1 subtracted the cached share and v0.28.2 the written share as well, so `Usage.Total` equals the provider's own `total_tokens` on every backend (`go.mod`)

## [0.76.0] - 2026-09-24

### Changed

- feat(compaction)!: the tier ladder is two rungs instead of three. `limits.compaction.hard_ratio` is **removed**, and `limits.compaction.mid_ratio` is **renamed to `smart_ratio`** — the old name is refused at load by the strict decoder rather than silently ignored, and under the environment layer `BULLE_COMPACTION_MID_RATIO` and `BULLE_COMPACTION_HARD_RATIO` are now refused too, since that layer would otherwise ignore them and quietly run on the default. Migrating is deleting one line and renaming the other; there is nothing to re-tune, since the forced fold is now derived rather than configured. Whether a pass *forces* is no longer scheduled by that second ratio: it is derived, from whether the fold the judge produced would leave the conversation under its trigger (`compaction.LandsUnder`, `compaction.Fold.Forced`). A fixed ratio was a proxy for that question and a coarse one — two conversations at 0.81 and 0.89 of the usable window ran the same pass, though only the second needed forcing — and it drifted on large windows, because the reserve is capped at 64k and stops growing above a 320k window: the old `hard` rung sat at 72% of the raw window at 200k but 84% at 1M, later than the 70–75% the field converges on. What forcing costs is bounded but real: it never drops a block, since a `prune` still needs the judge's probability and confidence — but the blocks it overrides are exactly the ones the judge said must survive verbatim, and they go to a summarizer whose prompt says to be as short as correctness allows, so a fact in one can be lost to summarization. That is why the decision is made on an upper bound of the rebuilt size rather than a guess. The judge is no longer asked to force either — `JudgeRequest.Force` is gone, and forcing is one lever pulled after the verdicts rather than a parameter to the question that produced them (`internal/compaction`, `internal/settings`, `internal/tui`, `internal/agent`)
- docs(compaction): the `limits.compaction` block in `~/.bulle.yml` now says which settings are a decision and which are a default — `judge.enabled` and `compact_at` are the two that are yours, and the ratios carry the rule that the soft tier is free and the smart tier is a call, so raise `soft_ratio` before touching `smart_ratio`. No value changed (`internal/settings`, `README.md`, `docs/configuration.md`)

### Fixed

- fix(settings): `TestShowHooksFromEnv` and `TestSetupAgentCustomTools` resolved settings through the real chain and so read the developer's own `~/.bulle.yml` and whichever profile it names — an unrelated key in someone's config could fail the suite, which is exactly how the `hard_ratio` removal surfaced. Both now run on a home of their own, the way their siblings already did (`internal/settings`, `internal/agent`)

## [0.75.0] - 2026-09-24

### Added
- feat(settings): `provider.api_key_command` and `limits.compaction.judge.api_key_command` — a key can be the output of a command instead of the value of a file, so a config file, a profile or a dotfiles repo can carry no secret at all: `api_key_command: "tiroir get OPENROUTER_API_KEY"`, run through `sh -c`, its trailing newline trimmed, its stdin closed so a prompt cannot seize the terminal, killed after 10 seconds. It is a source and not an override — a key from a flag, the environment, a literal `api_key` or a profile wins, and the command fills only what is empty — so an exported `TYPESAFE_API_KEY` keeps working untouched and a judge command never runs while one is set. The command runs only where a key is used, when a session builds its provider or its judge, so inspection commands (`bulle list`, `bulle sessions`, `bulle cron list`) never run it and a locked secret store cannot break them. A command that fails, times out or prints nothing is refused at startup naming the field, never left as an empty key, which would reach the backend as "no credential" and read exactly like a config that forgot one (`internal/settings`, `internal/agent`, `internal/tui`)

## [0.74.0] - 2026-09-23

### Added
- feat(settings): `limits.compaction.keep_turns` is refused at load when the floor it names alone exceeds the half of a pass the tail may fill, which it previously overrode silently: a generous floor pinned the window open above the trigger, so no pass could move it, because the active window is never rewritten. The check fires only where the cap is decidable at that layer, a written `window_tokens` or `compact_at`, since with neither the window is the backend's own and refusing would judge a number nobody wrote. The floor is weighed with the package's existing one-block bound rather than a second estimate (`internal/settings`, `internal/compaction`)

### Fixed
- fix(compaction): a consolidating pass now folds the history instead of trusting the judge's keep verdicts, so the rewrite it asks for is measured against turns no earlier pass compressed. A keep-heavy classification used to leave the pass with no ledger block to fold, and those turns are what carry the earlier ledger into the ask: the summarizer was never called, the consolidating addendum never reached the model, and a body past `MaxLedgerTokens` stayed over budget for the rest of the session (`internal/tui`)
- fix(compaction): the ledger's identity is claimed by a marker a reader cannot type. The sentinel now ends in the private-use codepoint U+E000, because matching the whole first line (the 0.73.0 fix) was not enough on its own: a turn whose first line spells the marker the way a document prints it still matched, and identity decides whether a message is folded into or pruned away. `ledgerIndex` takes the first match in the `anchor..active` range, so such a turn took the ledger's zone while the ledger the package wrote read as history behind it, breaking I3a by ordering alone. Binding identity to the assistant role is not available, since assembly picks the ledger's role as the opposite of its neighbour and an assistant-anchored session is folded into a *user*-role ledger; neither is recording it outside the message text (`nacelle.Message` is a role and a list of parts, and `Part` is sealed to `nacelle`). The marker being untypeable is therefore the discriminator. Nothing on disk is orphaned by the change: a resumed session is rebuilt from its question and answer lines alone (`internal/sessions`), never from the ledger (`internal/compaction`)

## [0.73.0] - 2026-09-23

### Added
- feat(compaction): a run the provider refuses for context length now compacts and retries once, instead of failing with the request still too long. The refusal is recognised across the providers' own wording (`internal/overflow`), held quietly rather than printed, and answered by a *forced hard* pass followed by the run being started again on the compacted conversation — once per turn, and not at all with `compact_at: 0` (`internal/tui`)
- feat(compaction): a ledger body past its budget (2000 tokens, the same ceiling one summary is written under) is *consolidated* — one rewritten block instead of an addition — which is the only thing that lets a ledger that otherwise only ever grows get smaller again (`internal/compaction`, `internal/tui`)
- feat(compaction): `limits.compaction.keep_tokens` — the verbatim tail is now sized by a token budget (40k by default) instead of a count of messages, so two heavy file reads can no longer hold the window open just by being the newest turns. `keep_turns` drops to 1 and becomes the *floor* under that budget: the live turn alone, the one turn no summary may stand in for. Neither bound may take more than half of what a pass may fill, since a tail that size could not land under the trigger it fires (`internal/compaction`)
- feat(compaction): `limits.compaction.reserve_tokens` — a runway held back from the context window for the model's own answer (a fifth of the window within 8k–64k when unset). The ratio ladder is now read against the window *less* that reserve, so `hard_ratio: 0.90` leaves the reserve plus a tenth of the usable window for the response instead of a tenth of the raw one
- feat(compaction): `limits.compaction.window_tokens` — an explicit override of the backend's reported context window, for the gateways that under-report one and the runners that report none at all (which otherwise leave every session on one global `75000`)
- feat(tui): `/status` names the window the ladder is measured against — `window · 200k raw, 160k usable, 40k reserved for the answer` — and the footer's ratio is the load against that usable window, so the figure and the tier can no longer disagree

### Changed
- fix(compaction): the ledger no longer grows without bound, nor doubles itself on every pass. The old fold handed the previous ledger to the summarizer and concatenated whatever came back unless it textually contained the old body, so a model that *reworded* what it was shown produced `previous + reworded previous + new` — the body doubled a pass and nothing measured it, because the never-grow check compares whole-conversation estimates. The fold is now a structural merge (per section, line by line, on a normalized key) and the prompt says plainly why the earlier ledger is shown: so it is not repeated. The ledger's one rewrite is a consolidation, gated on the rewrite still naming every identifier the old body named (`MissingIdentifiers`), and a rewrite that would drop one is refused and the merge stands — growth, which the next pass can attack, rather than a fact nobody can recover (`internal/compaction`, `internal/tui`)
- fix(compaction): I3 is restated in three parts rather than one, because the single sentence was already violated in letter while the property it protected held: the ledger is never a judge block, never history and never a prune candidate (I3a); no summarizer call is ever handed the ledger *alone*, so a rewrite is always measured against turns no earlier pass compressed (I3b); and the body is never silently rewritten (I3c) (`docs/plan-compaction-robust.md`)
- feat(compaction): the pre-send guard now compacts at the trigger itself rather than `compactSlack` above it. The trigger is already the soft ratio of the window a turn can fill, so waiting no longer bought headroom — it only changed which tier answered, and upward. The thrash guard's "did the pass land" test moved to that same figure: a margin between the two reset the count on a size the next send compacts again, so a session in that band fired a full pass on every send with nothing said. `compactSlack` is now a fixture offset in the tests only (`internal/tui`)
- fix(compaction): the byte measure counts what a request actually carries. A tool call's own arguments — often the largest thing in an edit or write turn — counted as nothing, and a `Reasoning` block counted in full although every backend drops it when it builds a request. Both fed the freed-token report, the never-grow check and `evictionCanLandUnder` (`internal/compaction`)
- feat(settings): `keep_turns`, `keep_tokens`, `reserve_tokens` and `window_tokens` are validated at load, and the last three are threaded through the environment as `BULLE_COMPACTION_KEEP_TOKENS`, `BULLE_COMPACTION_RESERVE_TOKENS` and `BULLE_COMPACTION_WINDOW_TOKENS`

### Fixed
- fix(settings): the `limits.compaction` guards were written as range tests, which are false for NaN, and `strconv.ParseFloat` accepts `nan` while the YAML resolver accepts `.nan` — so `hard_ratio: .nan` loaded clean and, through `int64(NaN*window+0.5)`, pinned every message at the hard tier: a judge call and a summarizer call on every turn, forever, with the config reading as valid. The guards now reject NaN. A negative `compact_at` is refused too, rather than silently read as "off" by every gate that tests for a positive ceiling, which took overflow recovery with it (`internal/settings`, `internal/agent`)
- fix(compaction): the thrash guard counted a pass as "did not land" only above the trigger plus a margin, while the guard that fires a pass acts at the trigger itself — so a pass stopping between the two cleared the count, the guard never stood down, and a full pass fired on every send with nothing said. Both now compare against the same figure (`internal/tui`)
- fix(compaction): the ledger sentinel was matched by prefix, so a message that merely opened with the marker was read as the ledger and the real one became a prune candidate, breaking I3a. It is now matched as a whole first line, and the package's own ledger identity is pinned by a test either way (`internal/compaction`)
- fix(compaction): the selectors trusted a span that reached past the conversation, which is a process-killing panic the moment a plan and a conversation disagree; they now trim instead. An empty-body ledger — a shape the package installs on purpose — is no longer dropped in place of a role repair (`internal/compaction`)
- fix(compaction): `foldOrphanResults` was untested, so the guard that keeps a tool pair whole at the history boundary could have been deleted with the suite staying green; it now has a test on the shape it exists for, and so do I3a and I3b (`internal/compaction`, `internal/tui`)
- fix(compaction): a job's `compaction` block was validated before it merged over the session's, so a usable job could combine into a ladder that cannot tier with nothing re-reading the merged result (`internal/agent`)
- fix(compaction): the judge calibration harness ran a live, billed call inside `go test ./...` whenever a credential happened to be in the environment. It is now behind an explicit `BULLE_CALIBRATION=1` opt-in as well as the key (`internal/compaction`)
- fix(tui): overflow recovery ignored an aborted turn and could retry and answer it anyway; and the pass snapshot shared the `Parts` arrays it read, leaving its isolation resting on a comment rather than on the code (`internal/tui`)
- fix(settings): `FromFlags` declared its names on the process's flag set on every call, so `go test -count=N` died on a redefined-flag panic and only `-count=1` passed. The declaration is now memoized once per process (`internal/settings`)

## [0.72.0] - 2026-09-22

### Added
- feat(compaction): a labeled-example calibration harness for the judge — one batched call over a corpus of labeled history blocks, re-scored across a sweep of prune thresholds, printing where the model's confidence and its accuracy part company (`TYPESAFE_API_KEY=... go test ./internal/compaction -run JudgeCalibration -v`; skipped without a key, so CI never pays for it)

### Changed
- feat(compaction): `limits.compaction.judge.prune_threshold` defaults to `0.75`, not `0.85`. The old number had never been measured against the model; the harness shows it recovering 17% of the blocks a careful operator would drop where 0.75 recovers 50%, at the same zero false prunes, and lifts agreement with the labels from 65% to 76%

### Fixed
- fix(compaction): the pre-send guard did nothing at all on a backend that cannot count tokens. It read `if count, err := CountTokens(...); err == nil && count > trigger` — and the OpenAI-compatible runner answers with `*Unsupported`, so the error branch skipped the whole check while the post-turn trigger kept running and compaction still looked healthy. A conversation already past the ceiling was sent anyway and the overshoot was absorbed into the next turn instead of prevented; the guard now falls back to the last usage-reported size, so the two automatic triggers agree on what the context costs (`internal/tui`)
- fix(compaction): a usage event reporting no input tokens erased `m.size`, the only measure both automatic triggers read, silently standing compaction down until a real usage arrived — a backend that reports usage on some events but not others was enough to switch it off. A zero is now read as "unknown" and the last real reading kept (`internal/tui`)
- fix(compaction): the judge's choice questions were rejected by the live endpoint — TypeSafe takes a choice's `criteria` as an object keyed by option, not a list, so every classification was a 422 and fell back to the mask. The judge had never successfully classified anything, and no test could see it because they all talk to a stub (`internal/compaction`)
- fix(compaction): every question in a batch was identical, so the model had no way to tell which block it was being asked about and answered the same way for all of them; each question now names its own block in its instructions (`internal/compaction`)
- fix(compaction): the criteria described a dead end under both `ledger` and `prune`, leaving the model no confident answer — the distribution flattened under the confidence floor and no verdict was ever acted on. The three descriptions are now disjoint, and confidence rises from 0.2–0.3 to 0.8–0.9
- fix(compaction): the soft tier tombstoned however little it found, rewriting the prompt-cache prefix for a few hundred bytes; it now stands down below `MinCleared`, and `DroppableBytes` is the pre-flight a caller measures a pass with before paying for the cache it invalidates
- fix(compaction): the judge's request was bounded in block count but not in bytes, so a handful of large tool results produced a multi-hundred-KB call to a small decision model; each block is now cut to `maxBlockText`, the goal to the same cap — it is the anchor's own text, which a pasted first message can make larger than every block together — and the batch to `defaultMaxState`, folding the oldest blocks that do not fit rather than asking about none of them
- fix(compaction): a tombstoned result's placeholder never closed its bracket — `[dropped 40000 bytes. Re-run the tool…` where the documented shape is `[dropped N bytes]`
- fix(compaction): the versioned model id that answered was decoded and thrown away, and the setting defaults to the drifting `jev-latest` alias; `/status` now names the model that answered and what it billed, so the setting can be pinned to the build the thresholds were tuned against (`internal/compaction`, `internal/tui`)
- refactor(compaction): a pass runs on a snapshot taken on the update loop instead of reading the model off its own goroutine, so its read safety is local rather than an invariant spread across every command that might mutate the session mid-pass (`internal/tui`)

### Removed
- chore(compaction): the dead `alignedEvictCut` wrapper, whose only caller in the repository was its own test, and the superseded `compaction-upgrade.md` spec at the repository root

## [0.71.0] - 2026-09-21

### Added
- feat(compaction): ratio-based context compaction with a soft/mid/hard ladder, a pinned anchor and a persistent `[state ledger]` that folds new facts into itself instead of re-summarizing (`limits.compaction`, `BULLE_COMPACTION_*`)
- feat(compaction): an opt-in TypeSafe System One judge (`limits.compaction.judge`, `TYPESAFE_API_KEY`) that classifies each history block keep/prune/ledger before the ledger is written; off by default because enabling it sends conversation history to a third party
- feat(tui): the status line shows the live context ratio and tier (`↕120k/200k · 0.60 · soft`), and `/status` reports the accumulated ledger's size, the last pass's tier, and whether the judge is on — the one setting that sends history off the machine is otherwise invisible once enabled

### Changed
- feat(settings): `limits.compact_at` is unset by default and the ceiling now derives from the compaction ratios against the backend's context window; an absolute value still overrides and `0` still disables compaction
- feat(settings): `limits.compaction` ratios and `judge.prune_threshold` are validated at load — a ratio outside `(0,1]`, a ladder with its rungs out of order, or an unusable prune threshold is now a startup error instead of a trigger that silently never fires

### Fixed
- fix(compaction): a pass can no longer orphan a tool pair the ledger absorbed — the ledger zone claims the replies to its own calls and carries an absorbed block forward, so a later prune cannot split a call from its result and a later rebuild cannot shed what a fold put in (`internal/compaction`)
- fix(compaction): a judged pass that comes back with an empty summary masks instead of folding away the turns the judge tagged for the ledger, so a provider that streams no text can no longer drop history silently and call it a summary
- fix(compaction): a rebuild heavier than the conversation it replaces is refused and the context left unchanged, so a pass never grows the context it exists to shrink (folding a two-byte turn can no longer install a verbose ledger)
- fix(compaction): the pinned head is extended over the replies answering the calls it carries, and a pass that folds nothing still stands a ledger between the head and the active window — the two cases where the assembly could merge a kept turn into the anchor or emit two messages of one role
- fix(jev): a transport error (a dropped connection or a per-attempt timeout) is retried like a 429/529, bounded by the caller's context rather than spending the attempt budget on requests that are already out of time
- fix(compaction): the soft tier tombstones tool results only — every backend drops a `Reasoning` block when it builds a request, so stubbing one freed no context while taking the chain of thought out of a transcript the reader can still scroll back to, and the pass reported the saved bytes as if it had bought headroom
- fix(compaction): a judge built without a prune threshold falls back to the shipped `0.85` instead of pruning on any probability at all, and a block the model kept can no longer be dropped by a threshold of zero
- fix(compaction): `soft_ratio: 0` no longer turns compaction off while the tier ladder still reports it on — the derived ceiling falls back to `75000` instead of zero, which every trigger reads as "disabled", and the setting is refused at load
- fix(compaction): a pass refuses a plan that no longer covers the conversation instead of indexing past the end of it, so a `/clear` or `/resume` landing mid-pass cannot take the session down with it, and the mask fallback refuses one too rather than promising a fallback it did not run
- fix(compaction): a tool call's own arguments (abbreviated) reach the judge, which was classifying every block from the tool's name alone
- fix(jev): a gateway's 502/503/504 is retried like a 429/529, and a zero-value client makes one attempt instead of returning an empty response with no error

## [0.70.2] - 2026-09-20

### Fixed
- fix(sandbox): `bulle remote` passes `-p` only when a port is configured; a bare `user@host` or `~/.ssh/config` alias keeps ssh's own `Port` instead of being overridden with 22, and `remote.port` defaults to 0 so the group no longer forces one
- fix(sandbox): a preflight host-key failure names the host and the command that resolves it — `ssh` for an unknown key (`ssh-keyscan` when the port is literal), `ssh-keygen -R` for a rotated one — instead of only the ssh exit status, and points a loopback target at `bulle sandbox`

## [0.70.1] - 2026-09-20

### Fixed
- fix(sandbox): a boite target starts in the guest's `/workspace` — the directory boite bakes and syncs — not the host directory an instance records; the preflight guard refused every VM session with `project directory "/home/yann" does not exist in target`
- fix(sandbox): a target session with no configured workdir records `target:<name>` as its root, so the banner, system prompt and session header name the target instead of the directory bulle was launched from
- fix(sessions): a `target:<name>` root groups and resumes by target identity, never folded into the launch directory
- fix(usage): usage records for a target session carry the target as the project and no branch, instead of the local repository bulle ran inside

### Changed
- refactor(sandbox): `bulle sandbox list` prints the directory a session's tools run in, rather than the instance's host-side workspace

## [0.70.0] - 2026-09-19

### Added
- feat(sandbox): split sandbox into local boite VMs and bulle remote execution backends

### Fixed
- fix: remove duplicate `session.additional_prompt` definition

## [0.69.0] - 2026-09-18

### Added
- feat(settings): `session.additional_prompt` (`BULLE_ADDITIONAL_PROMPT`, `-additional-prompt`), appended after the base prompt and skills catalog so a specialist persona layers on without replacing the harness guidance
- feat(settings): profiles carry `limits` and `additional_prompt`, keeping identity and model-aware tuning together while tools and security stay in `~/.bulle.yml`

### Fixed
- fix(settings): a profile's identity now beats `~/.bulle.yml` however the profile is selected — before, the file's own `profile:` key lost to the file's other keys while `-profile` and `BULLE_PROFILE` won, so one profile meant two different things and a scaffolded config could name a profile and never see its backend
- fix(docs): `example.bulle.yml` is byte-identical to the first-boot scaffold again, and a test now fails when they drift; the documented `ui.rendering_mode` and `ui.transparent_blocks` defaults follow the code (`tui`, `true`)

## [0.68.0] - 2026-09-18

### Added
- feat(sandbox): host-side agent execution with remote SSH tool remoting for microVM sandboxes
- feat(sandbox): remote implementations of `run_command`, `read_file`, `write_file`, `edit_file`, `list_directory`, `find_files`, and `search_content`
- feat(agent): `RunSessionWithTools` and `RunHeadlessWithTools` for custom tool set mounting

### Changed
- refactor(sandbox): remove guest binary synchronization; guest VMs require zero Bulle binaries and zero API keys
- refactor(sandbox): update preflight check to verify SSH connectivity and workspace readiness without requiring guest binaries

## [0.67.1] - 2026-09-18

### Fixed
- fix(tui): isolate model command unit tests from environment API keys

## [0.67.0] - 2026-09-18

### Added
- feat(tui): `/model` command for dynamic profile and model switching mid-session
- feat(settings): `~/.bulle/profiles/` directory for reusable provider and reasoning configs
- feat(cmd): `--profile` flag and `BULLE_PROFILE` env var for profile selection
- feat(sessions): resume last session by default and generate short 8-char hex session IDs

### Fixed
- fix(tui): preserve slash command autocomplete menu when model picker opens
- fix(provider): add provider catalog inspection and backend construction

## [0.66.0] - 2026-09-17

### Added
- feat(sessions): background session management with `bulle sessions` (alias `bulle list`), attach/resume, and kill
- feat(sessions): detached execution via `--detach` (`-d`) flag, `/detach` TUI command, and `ctrl+d`
- feat(settings): configurable `limits.max_concurrency` and `limits.max_parallel_agents` in `~/.bulle.yml` (default 16)

### Fixed
- fix(parallel): upgrade default subagent concurrency limit to 16 concurrent workers

## [0.65.9] - 2026-09-17

### Added
- feat(sandbox): multi-target sandbox support with ssh and boite backends
- feat(cron): display enabled and crontab installation status in cron list

### Fixed
- fix(editor): use ctrl+g as default keybinding for external prompt editor
- fix(sandbox): direct ssh fallback when boite is not installed

## [0.65.8] - 2026-09-17

### Fixed
- fix(editor): use ctrl+o and alt+e for opening external editor
- fix(sandbox): add list subcommand and styled cobra/fang help for boite VMs
- fix(sandbox): fix guest workdir resolution, shell quoting, and auto binary sync
- fix(cron): improve legacy crontab job line stripping with boundary checks

## [0.65.7] - 2026-09-17

### Fixed
- fix(cron): direct crontab install and uninstall with atomic blocks and log redirection
- fix(cron): automatic PATH augmentation and secret retrieval for unattended runs
- fix(cron): context timeout enforcement and error reporting in delivery logs
- feat(sandbox): isolated boite sandbox environment support

## [0.65.6] - 2026-09-17

### Fixed
- fix(editor): update tests for ctrl+shift+u default and gofmt
- fix: restore sandbox subcommand and suppress linter for unusedfunc
- fix(cmd): remove orphaned newSandboxCmd call
- fix: use markdown for editor prompt file
- chore: rebuild sandbox system per SANDBOX-PLAN.md

## [0.65.5] - 2026-09-16

## [0.65.4] - 2026-09-16

### Fixed
- fix: cron install command installs to crontab
