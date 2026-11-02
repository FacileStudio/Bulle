# Plan: per-user delegated access from Bulle to the suite

Status: final, 2026-11-02. Nothing here is implemented. Written so that one agent can take each work package on its own.

## Goal

A user signs in to Bulle once. From then on, the Bulle agent reads that user's own Courrier mail, Agenda events, Opus tasks and Sablier timer, in chat and in background jobs. The user pastes no token and approves nothing extra. No user ever sees another user's data.

## Why the current code goes

Commits `0018ca0`, `4dd8bda` and `161a96b` run the suite CLIs with server-wide tokens (`COURRIER_TOKEN`, `AGENDA_TOKEN`, `OPUS_TOKEN`, `SABLIER_TOKEN`). Every Bulle user and every cron job reads the owner's data. The same commits also:

- write the server's Sablier token into the shared `$HOME/.sablier.yml` on every call (concurrent runs race, and a `'` in the token breaks the YAML);
- switch the image from Alpine to Ubuntu only to run four CLI binaries downloaded with no checksum;
- choose a fallback by matching CLI error text (`"several mail accounts"`);
- advertise seven tools in the system prompt even when `buildSuiteTools` silently returned none;
- fail filet (`gen.file.long`, `go.file.funcs`, `go.func.long` in `suite_tools.go` and its test).

All of this is replaced, not patched.

## Design

### Principles

These come from the OAuth security BCP, the token-exchange RFC and the MCP authorization spec (sources at the end).

1. **Delegation, not impersonation.** Each token Bulle uses names the user as `sub` and Bulle as the actor (`act.sub`). Apps can log who acted (RFC 8693 §4.1).
2. **One audience per token.** A token for Courrier opens Courrier and nothing else. No token passthrough: Bulle never forwards its own login token to an app (RFC 9700 §2.3, MCP security best practices).
3. **Least privilege.** Tokens carry `<app>.read` only. Write scopes are defined now and granted later, behind Bulle's existing approval gate.
4. **Short-lived, nothing at rest.** Delegated tokens live 5 minutes and stay in Bulle's memory only. Bulle stores no app credentials anywhere.
5. **Revocation through the IdP.** Revoking Bulle's grant at Registre, or deactivating the user, stops new tokens. Existing ones die within 5 minutes.
6. **No secrets near the model or the VM.** Tokens never reach the LLM context, a process env, a file or the boite VM.

### Why this shape

| Option | Verdict |
|---|---|
| Per-user pasted tokens in Bulle settings | Not seamless. Long-lived secrets at rest. Rejected. |
| Bulle runs the device grant as `facile-cli` | One click per user, and the token opens every app. Rejected. |
| Registre token exchange to `facile-cli`, then porte's device exchange per app | Mints a long-lived, full-power, unlabelled app session per user that Bulle has to store. Rejected (it was the first draft). |
| ID-JAG / Cross App Access (`draft-ietf-oauth-identity-assertion-authz-grant`) | Built for apps behind **different** authorization servers. The suite has one (Registre), so it adds a hop for nothing. Rejected for now. |
| **Registre token exchange straight to an app-audience JWT, verified by porte's existing bearer path** | **Chosen.** porte already verifies JWTs for `OIDC_MACHINE_AUDIENCE` offline and maps `sub` to the local user through `porte_identities` (`porte/oidc/bearer.go`). Nothing is stored, every token is scoped to one app, and it fits zitadel/oidc, which Registre already uses. |

### Flow

```text
user ──login──▶ Bulle (porte) ── stores Registre access + refresh token in porte_identities
                     │
 agent calls courrier_inbox(userID)
                     │
                     ▼
     suite.Tokens.For(ctx, userID, "courrier")
        ├─ cache hit (exp > now+30s) ─▶ token
        └─ miss:
             1. load the user's Registre token, refresh it if expired
             2. POST sso.facile.studio/oauth/token
                  grant_type=urn:ietf:params:oauth:grant-type:token-exchange
                  subject_token=<user token, aud=bulle>
                  subject_token_type=urn:ietf:params:oauth:token-type:access_token
                  audience=<courrier client id>
                  scope=courrier.read
                  (client auth: bulle id + secret)
             3. ◀─ JWT {sub:user, aud:courrier, scope:"courrier.read",
                        act:{sub:"bulle"}, client_id:"bulle", exp:+5m, jti}
                     │
                     ▼
     GET https://courrier.facile.studio/api/... (Authorization: Bearer <JWT>)
                     │
     porte: verify sig/iss/aud/exp ─▶ Find(issuer, sub) ─▶ scope gate ─▶ handler
```

### Frozen contract

Every work package builds against this. A change to it needs the user's sign-off.

**Token-exchange request (Registre)**

| Field | Value |
|---|---|
| Endpoint | Registre's existing token endpoint (from discovery) |
| `grant_type` | `urn:ietf:params:oauth:grant-type:token-exchange` |
| Client auth | Required. Confidential clients only; public clients are refused. |
| `subject_token_type` | `urn:ietf:params:oauth:token-type:access_token` only |
| `subject_token` | A Registre access token whose `aud` contains the calling client's id |
| `audience` | Exactly one target app client id |
| `scope` | Space-separated subset of what the seed allows for that audience |
| No `actor_token` | The authenticated client is the actor |

**Issued token**

| Claim | Value |
|---|---|
| `iss` | `https://sso.facile.studio` (unchanged) |
| `sub` | Same as the subject token |
| `aud` | `[<target app client id>]` |
| `scope` | Granted scopes, space-separated |
| `act` | `{"sub": "<calling client id>"}` |
| `client_id` | Calling client id |
| `exp` | At most 300 s after `iat` |
| `jti` | Recorded in Registre's access-token table, like every grant |

Response: `issued_token_type=urn:ietf:params:oauth:token-type:access_token`, `token_type=Bearer`, `expires_in`, `scope`. No refresh token.

**Refusals:** `invalid_grant` for any problem with the subject token. `invalid_target` for an audience the client may not request. `invalid_scope` for a scope outside the allowance. Each refusal returns the same body for the same class, with no hint about which check failed.

**Scope names:** `courrier.read`, `courrier.write`, `agenda.read`, `agenda.write`, `opus.read`, `opus.write`, `sablier.read`, `sablier.write`. v1 grants only `*.read`.

**Seed shape (Registre `seed.yaml`, on the `bulle` client)**

```yaml
token_exchange:
  - audience: tniN7PWx8mSOCwVAAvYpBexYnMFWAM5flt19rNl0   # courrier
    scopes: [courrier.read]
  - audience: vvVAmcun8BFuAxMdNrp6lhqs2anzvz3vnzkeZh3U   # agenda
    scopes: [agenda.read]
  - audience: aHKmbSdWbJKqUMdTjIqFz2h4YtKYkvMiTB39zr4D   # opus
    scopes: [opus.read]
  - audience: 5TVdL3zCbhhh7XCs3MvaIr5nqoV2ZIMLeEdjnVYu   # sablier
    scopes: [sablier.read]
```

**porte scope gate:** a request authenticated by a JWT that has a `scope` claim passes only when:
- `GET`, `HEAD` or `OPTIONS`: the claim contains `<prefix>.read` or `<prefix>.write`;
- any other method: the claim contains `<prefix>.write`.

`<prefix>` comes from a new `porte.Config.ScopePrefix` (`courrier`, `agenda`, `opus`, `sablier`). A JWT with no `scope` claim is refused when `ScopePrefix` is set. Cookie and opaque-session requests are untouched.

**Env names**

| Where | Var | Value |
|---|---|---|
| each app | `OIDC_MACHINE_AUDIENCE` | the app's own Registre client id |
| each app | `OIDC_SCOPE_PREFIX` | `courrier` / `agenda` / `opus` / `sablier` |
| Bulle | `SUITE_COURRIER_URL`, `SUITE_AGENDA_URL`, `SUITE_OPUS_URL`, `SUITE_SABLIER_URL` | app base URLs; an empty one disables that app's tools |
| Bulle | `SUITE_<APP>_AUDIENCE` | the app's Registre client id |

Open item for WP2: Registre's `suite-ci` service account declares `audiences: [courrier]`, a name and not the client id, while porte's comment says `MachineAudience` is the app's client id. WP2 confirms which string Registre actually puts in `aud`. The contract uses the client id. If Registre resolves names, WP2 documents that and the seed may use names instead.

## Work packages

Dependencies: WP1 and WP2 can start now and run in parallel. WP4 can start now against fakes. WP3 needs a WP1 release. WP5 needs everything else.

```text
WP1 porte ─────────▶ WP3 apps ─┐
WP2 Registre ──────────────────┼──▶ WP5 deploy + live check
WP4 Bulle (fakes) ─────────────┘
```

### WP1: porte (repo `porte`, release v0.7.0)

Auth floor applies.

- `oidc/jwt.Claims`: add `Scopes []string` (parsed from the space-separated `scope` claim) and `Actor string` (from `act.sub`). Ignore a nested `act`.
- `porte.Identity`: add `Scopes []string` and `Actor string`. Fill them in `bearerVerifier.VerifyJWT`. Cookie and session identities leave them empty.
- `porte.Config`: add `ScopePrefix string`, read from `OIDC_SCOPE_PREFIX`. Boot refuses `ScopePrefix` without `MachineAudience`.
- `session.Manager.RequireAuth`: apply the scope gate from the contract. A refusal is `403` with the same body for every reason.
- Expose `porte.ActorFrom(ctx)` so apps can log "Bulle on behalf of <user>".
- Tests, one per row: read scope + GET passes; read scope + POST is 403; write scope + POST passes; no scope claim with prefix set is 403; prefix unset leaves today's behaviour unchanged (keeps `suite-ci` working); a cookie session is never scope-gated.
- Update `README.md` and `SPEC.md`, add a CHANGELOG entry, tag v0.7.0.

### WP2: Registre (repo `Registre`)

Auth floor applies.

- Implement zitadel/oidc v3's `op.TokenExchangeStorage` (the library ships `pkg/op/token_exchange.go` in v3.49.2). Do not hand-roll the grant.
- Advertise the grant in discovery (`modules/oidc/discovery.go`).
- Seed: add `token_exchange` (contract shape) to the application schema in `internal/seed`. Reject unknown keys, a scope that does not start with the audience's prefix, and an audience that is not a seeded client. Add the `bulle` block.
- Validation, in this order, all before any store lookup that leaks timing:
  1. caller is a confidential client authenticated with its secret;
  2. subject token signature, `iss`, `exp`;
  3. subject token `aud` contains the caller's client id;
  4. subject token has no `act` claim (no chained delegation in v1);
  5. subject token `jti` not revoked; the user exists and is active;
  6. `audience` is in the caller's `token_exchange` list;
  7. requested scopes are a subset of that entry; an empty request means the full allowance.
- Mint with the existing `mintAccessToken`, adding `scope`, `act` and `client_id`. TTL is `min(300 s, subject token remaining lifetime)`.
- Write one audit line per exchange: caller, subject, audience, scopes, jti. Never log the tokens.
- Resolve the `aud` naming open item from the contract and record the answer in the seed comment.
- Tests: one refusal per validation step, one success for `bulle` to each audience, and a check that the TTL never outlives the subject token.
- Add the seed block in the same PR. Registre's `seed.Apply` upserts on boot.

### WP3: Courrier, Agenda, Opus, Sablier (one agent per repo, or one for all four)

Each needs WP1 v0.7.0.

- Bump porte to v0.7.0. Run `go mod vendor` where the repo vendors (Agenda does).
- Wire `OIDC_MACHINE_AUDIENCE` and `OIDC_SCOPE_PREFIX` in `apps/api/internal/env/env.go` into `porte.Config`. Add both to `.env.example` and `docker-compose.yml` with empty defaults.
- Wherever the app writes an audit or activity line, include `porte.ActorFrom(ctx)` when set.
- Check that each read endpoint WP4 needs (list below) works with a bearer JWT and needs no cookie-only data such as a CSRF token or a session row. Fix any that do.
- Test: a fake-signed JWT with `courrier.read` reads the inbox and gets 403 on a send.
- Release each app.

Endpoints WP4 needs. The WP3 agent confirms each path from the app's router and writes the real path back into this table:

| App | Read |
|---|---|
| Courrier | list accounts; list conversations (account, unread, limit); search (query, account, limit) |
| Agenda | list calendars; list events (calendar, from, to) |
| Opus | list workspaces; list tasks (workspace, project, status, limit) |
| Sablier | current running timer |

### WP4: Bulle (repo `Bulle`)

Can start now against an in-process fake Registre and fake apps.

**Remove**
- `apps/api/internal/runner/suite_tools.go` and its test, `CommandExecutor`, `Config.SuiteExecutor`, `ensureSablierConfig`.
- The suite CLI downloads in the `Dockerfile`. Revert the runtime image to `alpine:3.21` as before `4dd8bda`.
- `COURRIER_*`, `AGENDA_*`, `OPUS_*`, `SABLIER_*` from `docker-compose.yml`.
- The fixed `suitePrompt` string.

**Add `apps/api/internal/suite/`**
- `tokens.go`: a `Tokens` type with `For(ctx, userID, app) (string, error)`. It reads the user's Registre token through porte's `IdentityStore`, refreshes it with the stored refresh token when needed (and writes the rotated pair back), calls the exchange, and caches results in memory keyed `(userID, app)` until `exp - 30 s`. Concurrent misses for the same key share one exchange (`golang.org/x/sync/singleflight` if already in the module graph, otherwise a small mutex map). `Forget(userID)` drops the cache.
- `errors.go`: typed errors the tools turn into plain sentences:
  - `ErrNotConnected`: refresh token missing or rejected. Tell the user to sign in to Bulle again.
  - `ErrNoAccount`: the app answered 401 for an unknown subject. Tell the user to open the app once.
  - `ErrAppDown`.
- `courrier.go`, `agenda.go`, `opus.go`, `sablier.go`: thin typed HTTP clients for the WP3 endpoints only, using `net/http` and `encoding/json`. Resolve defaults (default mail account, personal calendar) by reading the account or calendar list, never by matching error text.
- `tools.go`: `Tools(tokens, cfg, userID) []nacelle.Tool`. It registers an app's tools only when that app's URL and audience are set, and returns an error instead of a silent `nil`. It also returns the prompt line listing exactly the tools it registered.
- Keep each file under the filet limits (250 lines, 8 functions).

**Wire**
- `runner/run.go`: build suite tools per run with `lr.userID`. The cron channel gets them too, under the job owner's `userID`.
- Logout and back-channel logout handlers call `tokens.Forget(userID)`.
- The cron runner treats `ErrNotConnected` as a delivered notice ("Reconnect Bulle to let this job read your mail"), not a crash loop.

**Client**
- A "Connected apps" section in Settings that calls a new `GET /api/suite/status`. For each app it shows connected, no account yet (with a link), or reconnect needed. This is the only UI. Use muse tokens and Svelte 5 runes, with no hover-only affordance.

**Tests**
- `Tokens`: cache hit, expiry refresh, singleflight under 50 concurrent calls, `Forget`, refresh-token rotation written back, each typed error.
- Isolation: two users in parallel through fake apps that echo `sub`. Each tool result contains only its caller's subject.
- Tools: an app with no URL registers no tools, and the prompt line matches the registered set.
- `filet check` clean.

### WP5: deploy and live verification (human plus one agent)

1. Deploy Registre with the seed block. Check discovery lists the token-exchange grant.
2. Deploy the four apps with `OIDC_MACHINE_AUDIENCE` and `OIDC_SCOPE_PREFIX` set.
3. With a real Bulle user token, call the exchange with `curl`, then call one read endpoint per app. Expect 200. Expect 403 on one write endpoint.
4. Deploy Bulle and remove the old suite tokens from Dokploy's compose env in the same step.
5. Two real accounts each ask "what's in my inbox" and "what's on today". Each sees only their own data.
6. A background job for the second account reads its own calendar.
7. Revoke Bulle at Registre for one account. Within 5 minutes its tools return `ErrNotConnected`.

## Later, not in this plan

- Write tools (send mail, create event, move task): grant `*.write` in the seed and route those tools through Bulle's approval gate.
- Sender-constrained tokens (DPoP, RFC 9449) once porte can verify proofs. The 5-minute TTL is the mitigation until then.
- Expose each app's read API as an MCP server so other agents reuse it. The token model above already matches the MCP authorization spec.

## Sources

- RFC 8693, OAuth 2.0 Token Exchange: https://www.rfc-editor.org/rfc/rfc8693.html
- RFC 9700, Best Current Practice for OAuth 2.0 Security: https://www.rfc-editor.org/rfc/rfc9700.html
- RFC 9700 summary: https://oauth.net/2/oauth-best-practice/
- MCP security best practices (token passthrough): https://modelcontextprotocol.io/specification/2026-07-28/basic/security_best_practices
- MCP authorization, audience validation: https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization
- MCP authorization security considerations: https://modelcontextprotocol.io/specification/2026-07-28/basic/authorization/security-considerations
- Auth0, On-Behalf-Of token exchange: https://auth0.com/docs/secure/call-apis-on-users-behalf/on-behalf-of-token-exchange
- ID-JAG draft (considered, not chosen): https://datatracker.ietf.org/doc/draft-ietf-oauth-identity-assertion-authz-grant/02
- Okta Cross App Access / AI agent token exchange: https://developer.okta.com/docs/guides/ai-agent-token-exchange/authserver/main/
- Code read: `porte/oidc/bearer.go`, `porte/oidc/handlers.go`, `porte/porte.go`, `Registre/seed.yaml`, `Registre/apps/api/modules/oidc/storage.go`, zitadel/oidc v3.49.2 `pkg/op/token_exchange.go`
