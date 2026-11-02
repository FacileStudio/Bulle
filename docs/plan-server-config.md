# Plan: server-held config and provider keys

Status: approved 2026-11-02. Phases 1 to 3 are being built. Phases 4 and 5 are later.

## Goal

The Bulle server holds provider keys. The CLI keeps a server URL and a login token. The server proxies model calls, so a laptop never sees a vendor key. Local-only mode (config.yml, env, `api_key_command`) keeps working unchanged, but the documented recommended flow is the server.

## Decisions

- Key delivery: proxy. The API exposes `/api/llm/{backend}/*` and injects the real key.
- Key owner: per user, with the instance env key (`ANTHROPIC_API_KEY` etc.) as fallback.
- Local mode stays first-class. Docs recommend the server and describe local mode as the standalone option.
- Scope now: phases 1 to 3.

## Contracts (fixed, so tracks build in parallel)

### Migration `apps/api/migrations/00006_provider_credentials.sql`

```sql
CREATE TABLE provider_credentials (
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    backend    TEXT   NOT NULL,
    ciphertext BYTEA  NOT NULL,
    nonce      BYTEA  NOT NULL,
    last4      TEXT   NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, backend)
);
```

### Package `apps/api/internal/credentials`

```go
type Store interface {
    Put(ctx context.Context, userID int64, backend, plaintext string) error
    Delete(ctx context.Context, userID int64, backend string) error
    List(ctx context.Context, userID int64) ([]Entry, error)
    Resolve(ctx context.Context, userID int64, backend string) (string, error)
}
type Entry struct {
    Backend   string
    Last4     string
    UpdatedAt time.Time
}
```

- AES-256-GCM. Master key from env `BULLE_CREDENTIALS_KEY` (32 bytes, base64). Associated data is `"<userID>:<backend>"`.
- `Resolve` returns the user's key, else the instance env key (`providerKey` logic in `apps/api/internal/runner/tools.go`), else `ErrNoKey`.
- Backends allowed: `anthropic`, `openai`, `google`, `openrouter`.
- No call ever returns plaintext to an HTTP client.

### HTTP module `apps/api/modules/credentials` (auth required)

- `GET /credentials` returns `[{backend, last4, updated_at, source: "user"|"instance"|"none"}]` for all four backends.
- `PUT /credentials/{backend}` with `{"api_key": "..."}`.
- `DELETE /credentials/{backend}`.
- `POST /credentials/{backend}/test` makes a minimal call to the vendor and returns `{ok, error?}`.
- The runner resolves keys through `credentials.Store.Resolve` for the session's user instead of `os.Getenv`.

### Proxy `apps/api/modules/llm`

- `ANY /api/llm/{backend}/*` with `Authorization: Bearer <bulle token>` (the same porte session token the CLI stores).
- Resolves the key for the token's user, swaps the auth header (`x-api-key` for anthropic, `Authorization: Bearer` for openai and openrouter, `x-goog-api-key` for google), forwards the body, streams the response without buffering (flush per chunk).
- Strips hop-by-hop headers and the inbound Bulle token. Never logs bodies or keys.
- 401 for a bad token, 402-style JSON error `{"error":"no key for <backend>"}` with status 409 when no key resolves.

### CLI

- Config: `server: {url: ...}` group in `core/settings`. Token stored in the OS keyring, fallback file `$XDG_DATA_HOME/bulle/token` mode 0600. `BULLE_TOKEN` and `BULLE_SERVER_URL` env override.
- Commands: `bulle login [url]` (browser plus loopback via porte `flow=cli`, with `state` nonce required back; `--device` fallback prints URL and code if the server supports it, else paste-a-code), `bulle logout`, `bulle whoami`.
- Provider resolution: when a server is configured and the user has not set a local `provider.api_key`, `api_key_command`, or a vendor env key, set the backend `base_url = <server>/api/llm/<backend>` and the key to the Bulle token. Local settings still win (precedence unchanged: flag > env > profile > file > server > default).
- The anthropic backend must accept `base_url` (`core/provider`, SDK `option.WithBaseURL`).
- `bulle status` gains: server reachable, token valid, backends with a key.

### Web UI

- New page under `apps/client/src/routes/(app)/settings/providers`: one row per backend, masked `last4`, source badge, save, remove, test. Write-only inputs. muse tokens and components. Link it from the nav.

### Docs

- `docs/configuration.md`: new "Using a Bulle server" section placed first, marked recommended. Existing provider sections get a "standalone mode" note. Precedence table gains the server layer. Document `BULLE_SERVER_URL`, `BULLE_TOKEN`, `BULLE_CREDENTIALS_KEY`.
- `README.md` quick start leads with `bulle login`.
- `apps/api/.env.example` adds `BULLE_CREDENTIALS_KEY`.
- `example.bulle.yml` must stay byte-identical to the README scaffold block: change both or neither.

## Tracks and file ownership

| Track | Owns |
|---|---|
| A. Credentials backend | migration 00006, `apps/api/internal/credentials/`, `apps/api/modules/credentials/`, runner key resolution, `router.go` registration, `.env.example`, `env.go` |
| B. Proxy | `apps/api/modules/llm/` and its router registration line, the proxy test |
| C. CLI | `apps/cli/cmd/login*.go`, `apps/cli/cmd/status*.go`, `core/settings/` server group and provider resolution, `core/provider/` anthropic base_url |
| D. Web UI | `apps/client/` only |
| E. Docs | `docs/configuration.md`, `README.md`, `CHANGELOG.md` is left to the release |

Track B depends on the `credentials.Store` interface only. It builds against a fake until A lands.

## Later

- Phase 4: profiles in Postgres, `/model` lists them from the server, server layer in precedence.
- Phase 5: deprecation notes for `api_key_command`, CI token labels in the UI.

## Sources

- https://clerk.com/blog/adding-clerk-auth-to-your-cli.md
- https://docs.litellm.ai/docs/proxy/virtual_keys.md
- https://docs.cloud.google.com/kms/docs/envelope-encryption
