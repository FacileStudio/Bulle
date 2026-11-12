# The encre release the artifact validator comes from. Bump it with the
# @facile/encre pin in apps/client/package.json so the API and the editor
# agree on the format.
ARG ENCRE_VERSION=0.10.0

FROM oven/bun:1.4.4 AS client-build
WORKDIR /client
COPY apps/client/package.json apps/client/bun.lock* ./
RUN bun install --frozen-lockfile
COPY apps/client/ .
RUN bun run build

# encre's release ships the `encre` CLI as a static binary. It is fetched for
# the image's architecture and checked against the release's checksums.txt.
# Nothing here depends on this repository: the layer is cached until
# ENCRE_VERSION moves.
FROM alpine:3.21 AS encre-fetch
ARG ENCRE_VERSION
ARG TARGETARCH
RUN set -eu; apk add --no-cache curl; \
    base="https://github.com/FacileStudio/encre/releases/download/v${ENCRE_VERSION}"; \
    archive="encre_${ENCRE_VERSION}_linux_${TARGETARCH:-amd64}.tar.gz"; cd /tmp; \
    curl -fsSLO "$base/$archive" && curl -fsSLO "$base/checksums.txt"; \
    grep " $archive\$" checksums.txt | sha256sum -c -; \
    tar -xzf "$archive" -C /usr/local/bin encre; rm "$archive" checksums.txt; \
    /usr/local/bin/encre --version

FROM oven/bun:1.4.4 AS browser-build
WORKDIR /browser
COPY apps/browser/package.json apps/browser/bun.lock ./
RUN bun install --frozen-lockfile
COPY apps/browser/ .
RUN bun run build

FROM golang:1.26.11-alpine AS api-build

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /repo

COPY go.work go.work.sum ./
COPY core/go.mod core/go.sum ./core/
COPY apps/cli/go.mod apps/cli/go.sum ./apps/cli/
COPY apps/api/go.mod apps/api/go.sum ./apps/api/

RUN cd core && go mod download && cd ../apps/api && go mod download

COPY core ./core
COPY apps/api ./apps/api
COPY apps/client/package.json ./apps/client/package.json
COPY scripts/version.sh ./scripts/version.sh

WORKDIR /repo/apps/api

RUN APP_VERSION="$(sh /repo/scripts/version.sh)" && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -tags goolm -trimpath -ldflags="-s -w -X main.version=${APP_VERSION}" -o bin/api .

FROM alpine:3.21

RUN apk add --no-cache openssh-client ca-certificates \
    && addgroup -g 65532 nonroot \
    && adduser -u 65532 -G nonroot -h /home/nonroot -D nonroot

COPY --from=api-build /repo/apps/api/bin/api /api
COPY --from=client-build /client/build /client
COPY --from=browser-build /browser/dist/bulle-browser.mjs /browser/bulle-browser.mjs
COPY --from=encre-fetch /usr/local/bin/encre /usr/local/bin/encre

ENV CLIENT_DIR=/client
ENV ENCRE_BIN=/usr/local/bin/encre

EXPOSE 4000

USER nonroot:nonroot

ENTRYPOINT ["/api"]
