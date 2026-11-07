FROM oven/bun:1 AS client-build
WORKDIR /client
COPY apps/client/package.json apps/client/bun.lock* ./
RUN bun install --frozen-lockfile
COPY apps/client/ .
RUN bun run build

FROM oven/bun:1 AS browser-build
WORKDIR /browser
COPY apps/browser/package.json apps/browser/bun.lock ./
RUN bun install --frozen-lockfile
COPY apps/browser/ .
RUN bun run build

FROM golang:1.26-alpine AS api-build

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

ENV CLIENT_DIR=/client

EXPOSE 4000

USER nonroot:nonroot

ENTRYPOINT ["/api"]
