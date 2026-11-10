# Deploying bulle.facile.studio

Runbook for the Bulle web app (Go API plus SvelteKit client in one container) and the host daemon it depends on. Written 2026-11-01 from a read-only survey of ruche, Dokploy, Registre, Casier and boite. Nothing here has been executed.

## How the suite deploys

- Host: ruche, the machine `hostname` reports as `ruche`. Docker 29 in swarm mode, overlay network `dokploy-network`, Traefik behind Cloudflare. 12 cores, 62 GB RAM, about 120 GB free on `/`.
- Control plane: Dokploy at https://gare.facile.studio. The `dokploy` CLI on ruche is already authenticated against it (config in `~/.bun/install/global/node_modules/@dokploy/cli/config.json`, keys `url` and `token`).
- Every suite app is a Dokploy compose in project `Facile`, environment `production` (project id `HzqqotIk3BPxGsh4Y5pqX`, environment id `PQVxFx6mkl4nn0ZAFCjwV`). Sonde is compose `J3jWscC3f_QzNu6eGVsPf`: source GitHub `FacileStudio/Sonde`, compose path `./docker-compose.yml`, autoDeploy on, no Dokploy domains. A push to the tracked branch deploys.
- Routing: Traefik labels inside `docker-compose.yml`, one router pair per app, network `dokploy-network`. Labels replace what Dokploy injects.
- No sibling has a deploy workflow. Sonde, Capsule and Plume only run filet in Actions; Opus has a tag release workflow; Sablier publishes a GHCR image on tags. Deploys are Dokploy autoDeploy, so this repo gets no `deploy.yml`. Pushing `main` of this repo deploys the web app once the compose exists (see step 5), and CI does not gate it.
- Production env: the container never calls casier. Casier holds the values, `casier push dokploy <composeId> -p bulle -e prod` writes them into the Dokploy compose environment, and the compose file forwards each variable by name. The push replaces the whole Dokploy environment with the casier set, so casier must hold every key.
- Auth: Registre (https://sso.facile.studio). Clients are declared in `Registre/seed.yaml`, upserted at Registre boot. Registre has no CLI for clients; there is also an admin API (`POST /api/admin/clients`, cookie auth), which this runbook does not use.
- Worker host: boite-worker must run on ruche, because the container mounts its Unix socket from the host. ruche has `/dev/kvm` (`root:kvm`, mode 0660), `qemu-system-x86_64`, `genisoimage` and `mcopy`. `mkfs.vfat` (package `dosfstools`) is missing.

## Findings that change the plan

1. `bulle.facile.studio` has no DNS record today (`dig` returns nothing). `sonde.facile.studio` resolves to two Cloudflare proxy addresses, so records live in Cloudflare (nameservers alexa.ns and alexis.ns). No wildcard exists.
2. No Bulle compose exists in Dokploy. Nothing is live on that hostname, so `gare.yml` and `site/` are not serving anything now.
3. `apps/api/go.sum` was missing entries for `golang.org/x/text` and `golang.org/x/sync`. The workspace hid it locally (`go.work.sum`), but `docker build` has no `go.work` and failed at `go build`. Fixed on this branch with `GOWORK=off go mod tidy`.
4. The repo had no `.dockerignore`, so `COPY apps/client/` would have copied a local `node_modules` and `.svelte-kit` into the image. Added Sonde's.
5. The compose bind-mounted the socket file `/run/boite/worker.sock`. The worker recreates the socket on restart, and a file bind mount keeps the old deleted inode, so the API would lose the worker after any worker restart. The compose now mounts the directory `/run/boite`.
6. The compose gave uid 65532 no way to open a socket owned `root:<group>` mode 0660. It now has `group_add: ["${BOITE_SOCK_GID:-2001}"]`.
7. The Dockerfile is byte-identical to Sonde's. With the fixes above it builds locally (`docker build .` succeeded, image holds `/api` and `/client/index.html`, `docker compose config` is valid). It was not run against Postgres.
8. The worker docs disagree: `boite/README.md` says `POST /boxes` returns 201 after the box is up, `boite/docs/arch.md` says 202 with a creating status. Settle this before writing the Phase 3 client.
9. The worker, snapshot endpoints and `BOITE_CONTROL_PUBKEY` exist only on boite branch `feat/worker`, unpushed. Build from that checkout or push it first.
10. The control key is a private key and casier pushes the env as `KEY=value` lines, so a multi-line PEM would corrupt the Dokploy env. Store it base64-encoded on one line: the API reads `BULLE_BOX_KEY` (falling back to the older `BOITE_CONTROL_KEY`), accepts the PEM, its base64 or a path to either, and publishes the matching public key to each box.
11. The API does not read `BOITE_*`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY` or `ANTENNE_*` yet. The compose forwards them anyway, so the first deploy works with only the OIDC and database keys.
12. `casier` on ruche is not logged in, so casier contents for project `bulle` could not be listed. Run `casier projects` after `casier login` to see whether the project exists.

## Runbook

Run every command on ruche unless noted. Run the `jq` and `curl` blocks from a directory outside the repo, since the repo `mise.toml` is not trusted and the `jq` shim fails inside it.

### 0. Merge the branch

Merge this branch into `main` locally. Do not push `main` yet if you want the first deploy to be deliberate; with no Dokploy compose, a push deploys nothing, so pushing early is safe but pointless until step 5.

### 1. Install boite-worker on ruche

The compose mounts `/run/boite`, so the worker comes first. The computer network setting (Settings, Computer) needs boite-worker v0.12.0 or later. An older worker still serves open computers, but it refuses an offline one, so that computer fails to start rather than boot with the network open. Run v0.13.1 or later with `passt` installed (Debian package `passt`): open computers then run behind passt with `--no-map-gw`, so a computer cannot reach services the host binds to 127.0.0.1. ruche runs v0.13.1 since 2026-11-08 (previous binary kept as `~/.local/bin/boite-worker.v-pre-0.13.1.bak`).

```sh
sudo apt install dosfstools
sudo groupadd -g 2001 boite-sock
cd ~/Code/Facile/boite
go build -o boite-worker ./cmd/boite-worker
sudo install boite-worker /usr/local/bin/
sudo install -m 0644 scripts/boite-worker.service /etc/systemd/system/
sudo install -d -m 0700 /etc/boite
sudo install -d -m 0755 /var/lib/boite/cache
sudo cp ~/.boite/cache/boite.qcow2 /var/lib/boite/cache/boite.qcow2
ssh-keygen -t ed25519 -N '' -C bulle-control -f /tmp/bulle-control
ip -4 addr show docker0
```

The `cp` seeds the base image so the worker does not download 2 GB from https://boite.facile.studio/base.qcow2. Skip it to let the worker download and checksum it. Note the `docker0` address (10.128.0.1 on ruche) for the next file.

Create `/etc/boite/worker.env` with `sudoedit`, mode 0600:

```text
BOITE_WORKER_TOKEN=<value of BOITE_WORKER_TOKEN from step 2>
BOITE_WORKER_SOCK=/run/boite/worker.sock
BOITE_WORKER_SOCK_MODE=0660
BOITE_WORKER_SOCK_GROUP=2001
BOITE_SSH_BIND=10.128.0.1
BOITE_CONTROL_PUBKEY=<contents of /tmp/bulle-control.pub>
```

Then:

```sh
sudo systemctl enable --now boite-worker
systemctl status boite-worker
ls -l /run/boite
curl --unix-socket /run/boite/worker.sock -H "Authorization: Bearer $BOITE_WORKER_TOKEN" http://worker/boxes
```

Expected: socket `srw-rw---- root boite-sock`, and the curl returns `[]`. Keep `/tmp/bulle-control` (private half) until step 2 stores it, then delete both files. `BOITE_SSH_BIND` must be a private address, never `0.0.0.0`. Whether a container on the compose network can reach `10.128.0.1` is untested; the test is in step 6. If it fails, give the compose default network a fixed subnet and bind to its gateway instead.

The unit uses `KillMode=process`, so restarting the worker or redeploying the web app never stops a running box. Existing boxes under `~/.boite` (yours) are untouched because the worker uses `/var/lib/boite`.

### 2. Set production secrets in casier

```sh
casier login
casier projects
```

Use `casier secrets set -p bulle -e prod <KEY> <VALUE>` for each key. If `bulle` is missing from `casier projects`, create the project in the Casier UI at https://casier.facile.studio first. Generate each random value into a shell variable and reuse it; never write it to a file in the repo.

| Key | Value and where it comes from |
|---|---|
| `POSTGRES_PASSWORD` | `openssl rand -hex 24`. Must be set before the first deploy: the `bulle_db_data` volume is initialised with it and a later change does not rewrite the volume. |
| `OIDC_CLIENT_SECRET` | `openssl rand -hex 32`. The same value goes into Registre in step 3. |
| `BOITE_WORKER_TOKEN` | `openssl rand -hex 32`. The same value goes into `/etc/boite/worker.env`. |
| `BULLE_BOX_KEY` | `base64 -w0 /tmp/bulle-control`, the private key from step 1. The compose also accepts the older name `BOITE_CONTROL_KEY`. |
| `ANTHROPIC_API_KEY` | A human key from the Anthropic console. Not generated by any tool here. |
| `OPENAI_API_KEY` | Optional, a human key. Omit to leave the backend off. |
| `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_REDIRECT_URL`, `OIDC_SUCCESS_URL`, `OIDC_CLI_AUDIENCE`, `OIDC_MACHINE_AUDIENCE`, `OIDC_SCOPE_PREFIX`, `CORS_ALLOWED_ORIGINS` | Optional. The compose defaults are already the production values (`https://sso.facile.studio`, `bulle`, `https://bulle.facile.studio/api/auth/oidc/callback`, `https://bulle.facile.studio/`, `facile-cli`, `bulle`, `bulle`). The machine audience lets another suite app (Glouton) call Bulle with a Registre token-exchange JWT; the scope prefix limits it to `bulle.read` for reads and `bulle.write` for writes. Set them only to be explicit. |
| `TRUSTED_PROXIES`, `APP_VERSION` | Optional. `TRUSTED_PROXIES` defaults to `private,cloudflare` (Traefik and Cloudflare in front, so per-IP limits see the visitor). `APP_VERSION` is a build arg that GET /version reports; set it to the release tag at deploy. |
| `SUITE_COURRIER_MCP_URL` (and `SUITE_COURRIER_MCP_RESOURCE` when Bulle dials an internal address) | Optional. Courrier's MCP endpoint, `https://courrier.facile.studio/mcp`. Leave empty to switch mail off. See `docs/suite-access.md`. |
| `SUITE_AGENDA_MCP_URL` (and `SUITE_AGENDA_MCP_RESOURCE` when Bulle dials an internal address) | Optional. Agenda's MCP endpoint, `https://agenda.facile.studio/mcp`. Leave empty to switch the calendar off. |
| `SUITE_OPUS_URL`, `SUITE_SABLIER_URL`, `SUITE_NUAGE_URL`, `SUITE_MYCELIUM_URL`, `SUITE_GLOUTON_URL`, and the matching `SUITE_<APP>_AUDIENCE` | Optional, set a URL and its audience together. Leave empty to switch an app's tools off. URL is the app origin (`https://courrier.facile.studio`). Audience is the app's Registre client id from `seed.yaml`. Set them only after Registre serves token exchange and the app accepts `OIDC_MACHINE_AUDIENCE`. See `docs/plan-suite-delegation.md`. |
| `NTFY_URL`, `NTFY_TOKEN`, `PUSH_TOPIC_SECRET` | Optional. Push through a self-hosted ntfy server. `NTFY_URL` is the server origin (`http` or `https`), `PUSH_TOPIC_SECRET` is `openssl rand -hex 32` and derives each user's private topic (`bulle-` plus 32 hex characters of HMAC-SHA256 over the user id). Push is off unless both are set; the API boots either way. `NTFY_TOKEN` is a bearer token, only for a server with access control. Changing the secret changes every topic, so clients must resubscribe. Notifications are generic ("Approval needed", "<job> is ready") and never carry model output. |
| `JOURNAL_URL`, `JOURNAL_TOKEN` | Optional. A Journal token minted in Journal's admin UI. Sonde's Dokploy env sets none, so skipping is the suite norm. |
| `ANTENNE_INSTANCE`, `ANTENNE_APP`, `ANTENNE_SECRET`, `ANTENNE_INSTANCE_ID` | Optional. Values come from Antenne; copy Sonde's `ANTENNE_INSTANCE` and mint a secret for app `bulle`. |

Do not set `DATABASE_URL`: the compose builds it from `POSTGRES_*`. Check coverage:

```sh
casier check apps/api/.env.example -p bulle -e prod
```

It reports keys in the example that are missing remotely. `DATABASE_URL`, `PORT`, `LOG_LEVEL` and `TEST_DATABASE_URL` showing as missing is expected.

### 3. Register the Registre client

Registre registers clients from `Registre/seed.yaml` at boot (upsert on `client_id`, never deletes). It prints nothing and returns no secret: you choose the secret, Registre reads it from an environment variable and stores an argon2id hash. This is the same path Sonde took (commit `ee9f2fa feat(seed): register sonde as an oidc client`).

In the Registre repo, `~/Code/Facile/Registre`:

1. Append to `seed.yaml`:

```yaml
  - client_id: bulle
    name: Bulle
    backchannel_logout_uri: https://bulle.facile.studio/api/auth/backchannel-logout
    secret_env: REGISTRE_SEED_BULLE_SECRET
    redirect_uris:
      - https://bulle.facile.studio/api/auth/oidc/callback
    scopes: [openid, profile, email, offline_access]
```

   The redirect is compared byte for byte, `/api` prefix included. For local dev against Registre, add `http://localhost:5173/api/auth/oidc/callback` as a second redirect and use a separate dev client rather than sharing the production secret.
2. Add `REGISTRE_SEED_BULLE_SECRET: ${REGISTRE_SEED_BULLE_SECRET:-}` to the environment block of `Registre/docker-compose.yml`, next to `REGISTRE_SEED_SONDE_SECRET`. A name missing from that block never reaches the container and the boot fails.
3. Add the secret to the Registre compose environment in Dokploy (compose id `BkLI7Qn4IxrfL3Q9YDE_M`). Registre keeps its secrets in the Dokploy environment, not casier (its `docs/deployment.md`), so `casier push dokploy` would wipe them. Use a read, append, write round trip, and compare key sets:

```sh
cfg=~/.bun/install/global/node_modules/@dokploy/cli/config.json
url=$(jq -r .url $cfg)
key=$(jq -r .token $cfg)
id=BkLI7Qn4IxrfL3Q9YDE_M
curl -s -A curl/8 -H "x-api-key: $key" -G "$url/api/trpc/compose.one" --data-urlencode "input={\"json\":{\"composeId\":\"$id\"}}" | jq -r .result.data.json.env > /tmp/registre.env.before
cp /tmp/registre.env.before /tmp/registre.env.after
printf '\nREGISTRE_SEED_BULLE_SECRET=%s\n' "$secret" >> /tmp/registre.env.after
cut -d= -f1 /tmp/registre.env.before | sort > /tmp/k.before
cut -d= -f1 /tmp/registre.env.after | sort > /tmp/k.after
diff /tmp/k.before /tmp/k.after
curl -s -A curl/8 -H "x-api-key: $key" -H 'content-type: application/json' "$url/api/trpc/compose.update" -d "$(jq -n --arg id $id --rawfile env /tmp/registre.env.after '{json:{composeId:$id,env:$env}}')"
```

   `$secret` is the value stored as `OIDC_CLIENT_SECRET` in step 2. The diff must show exactly one added key. Delete the `/tmp/registre.env.*` files afterwards: they hold every Registre secret.
4. Commit and push Registre `main`. Registre autoDeploys, and the seed is applied on boot. If the Registre compose tracks another branch, check with `compose.one` first.

What it returns: nothing. Registre logs the applied seed at boot. Verify with discovery and an authorize request (a 302 to `/login?auth_request=...` means the client and redirect are known; a 400 naming the redirect means they are not):

```sh
curl -s https://sso.facile.studio/.well-known/openid-configuration | jq .issuer
curl -s -o /dev/null -w '%{http_code} %{redirect_url}\n' 'https://sso.facile.studio/authorize?client_id=bulle&response_type=code&scope=openid&redirect_uri=https%3A%2F%2Fbulle.facile.studio%2Fapi%2Fauth%2Foidc%2Fcallback&state=x'
```

### 4. DNS

In the Cloudflare dashboard for `facile.studio`, add a record for `bulle` with the same type, target and proxy setting as the existing `sonde` record (proxied). ruche's egress address is 65.108.2.107, but copy `sonde`'s origin record rather than assuming it. No CLI path exists on ruche for this. Check:

```sh
dig +short bulle.facile.studio
```

It should return Cloudflare addresses. Traefik gets its certificate through the `letsencrypt` resolver the first time the host is requested.

### 5. Create the Dokploy compose and deploy

Dokploy UI at https://gare.facile.studio: project Facile, environment production, Create Service, Compose.

- Name: `Bulle`
- Type: Docker Compose
- Provider: GitHub, repository `FacileStudio/Bulle`, branch `main`, compose path `./docker-compose.yml`
- Autodeploy: on
- Domains: none (the compose labels own routing)

Then push the environment from casier and deploy. `<composeId>` is shown in the compose URL:

```sh
export DOKPLOY_URL=https://gare.facile.studio
export DOKPLOY_API_KEY=$(jq -r .token ~/.bun/install/global/node_modules/@dokploy/cli/config.json)
casier push dokploy <composeId> -p bulle -e prod
```

Then press Deploy in the UI, or push `main` (autoDeploy triggers). The build takes about two minutes. Order matters: push the casier env first, because the first start creates the Postgres volume with whatever `POSTGRES_PASSWORD` is set.

From here, every push to `main` of this repo redeploys the web app. CI (`ci.yml`) runs on the same push and does not block it, so merge through a pull request.

### 6. Verify

```sh
sh scripts/verify-deploy.sh bulle
```

Nine checks, all must pass: `/` is HTML, SPA fallback, a hashed asset, a missing bundle 404s, `/api/auth/config` is JSON, no doubled prefix, `/api/health`, `/api/ready`, POST on an unknown route is refused. Then log in at https://bulle.facile.studio with a Registre account; the login must end on `/`.

Checks that need a person: the SSO round trip in a browser, and that the container can reach the worker:

```sh
docker exec $(docker ps -qf name=bulle-) /api healthcheck
sudo docker run --rm --network $(docker network ls --format '{{.Name}}' | grep bulle) busybox nc -z -w3 10.128.0.1 22
```

The second line only tests routing to the bridge address; it fails with connection refused (reachable) or timeout (blocked). Timeout means the SSH bind plan needs the fixed-subnet fallback from step 1. The socket itself cannot be exercised until Phase 3 code exists.

### 7. Rollback

- Web app, bad deploy: revert the commit on `main` and push (autoDeploy redeploys the previous tree), or in Dokploy set the compose branch to a known-good commit-bearing branch and redeploy. To take the site down, stop the compose in the UI (`compose.stop`); Postgres data stays in volume `bulle_db_data`.
- Database: migrations apply on startup and the files carry `-- +goose Down`. `docker run <image> migrate status` reads state without starting a server (needs `DATABASE_URL`). Roll the schema back only if the previous release cannot read it.
- Secrets: re-run `casier secrets set` and `casier push dokploy <composeId> -p bulle -e prod`, then redeploy.
- Registre: remove the `bulle` entry from `seed.yaml` for the record, but Registre never deletes clients, so the client keeps working until you remove it by hand through its admin API. Rotate the secret with `POST /api/admin/clients/bulle/secret` (admin cookie).
- Worker: `sudo systemctl stop boite-worker` leaves boxes running (`KillMode=process`). To remove it entirely: `sudo systemctl disable --now boite-worker`, then `sudo rm /usr/local/bin/boite-worker /etc/systemd/system/boite-worker.service /etc/boite/worker.env`. Box data stays in `/var/lib/boite`.

## What only the human can do

- [ ] `casier login` on ruche, create project `bulle` if absent, and run the step 2 `secrets set` commands (needs an interactive browser login).
- [ ] Supply `ANTHROPIC_API_KEY` (and `OPENAI_API_KEY` if wanted).
- [ ] Add the `bulle` DNS record in Cloudflare.
- [ ] Run the Registre steps 3.1 to 3.4 and approve the Registre push and redeploy (outward, affects every suite login).
- [ ] Run the sudo commands in step 1 (apt, groupadd, install, systemctl). The agent has no passwordless sudo.
- [ ] Push boite branch `feat/worker` and decide where the worker binary is built (ruche from the checkout, or a release).
- [ ] Create the Dokploy compose, then run `casier push dokploy`, then the first deploy.
- [ ] Decide on the 201 versus 202 contract for `POST /boxes` before Phase 3.
- [ ] Merge this repo through a pull request, since `main` deploys.
- [ ] Remove `gare.yml` and `site/` when the SPA landing ships.

## Open risks

- ruche is also a workstation: boxes share its 62 GB RAM and 12 cores with every suite app. Cap the default box size (worker defaults: 2 CPUs, 2G, 20G disk) and the number of users before opening sign-in.
- Box SSH ports sit on the host in the 2222 to 2322 range, bound to the docker0 address. Confirm no firewall rule exposes them publicly.
- A web redeploy interrupts live agent runs (plan Phase 4, step 25 handles this). It never stops a box.
